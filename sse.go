package ginji

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// SSEEvent represents a Server-Sent Event.
type SSEEvent struct {
	// ID is the event ID (optional)
	ID string

	// Event is the event type (optional)
	Event string

	// Data is the event data
	Data string

	// Retry is the reconnection time in milliseconds (optional)
	Retry int
}

// SSEStream represents an SSE stream.
type SSEStream struct {
	ctx           *Context
	lastEventID   string
	keepAlive     time.Duration
	keepAliveDone chan struct{}
}

// NewSSEStream creates a new SSE stream.
func NewSSEStream(c *Context) *SSEStream {
	// Set headers for SSE
	c.SetHeader("Content-Type", "text/event-stream")
	c.SetHeader("Cache-Control", "no-cache")
	c.SetHeader("Connection", "keep-alive")
	c.SetHeader("X-Accel-Buffering", "no") // Disable nginx buffering

	return &SSEStream{
		ctx:           c,
		keepAlive:     15 * time.Second,
		keepAliveDone: make(chan struct{}),
	}
}

// Send sends an SSE event.
func (s *SSEStream) Send(event SSEEvent) error {
	var sb strings.Builder

	// Write ID
	if event.ID != "" {
		sb.WriteString(fmt.Sprintf("id: %s\n", event.ID))
		s.lastEventID = event.ID
	}

	// Write event type
	if event.Event != "" {
		sb.WriteString(fmt.Sprintf("event: %s\n", event.Event))
	}

	// Write data (can be multi-line)
	lines := strings.Split(event.Data, "\n")
	for _, line := range lines {
		sb.WriteString(fmt.Sprintf("data: %s\n", line))
	}

	// Write retry
	if event.Retry > 0 {
		sb.WriteString(fmt.Sprintf("retry: %d\n", event.Retry))
	}

	// End event with double newline
	sb.WriteString("\n")

	// Write to response
	_, err := s.ctx.Res.Write([]byte(sb.String()))
	if err != nil {
		return err
	}

	// Flush immediately
	if flusher, ok := s.ctx.Res.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}

// SendData is a convenience method to send just data.
func (s *SSEStream) SendData(data string) error {
	return s.Send(SSEEvent{Data: data})
}

// SendJSON sends JSON data as an event.
func (s *SSEStream) SendJSON(v any) error {
	data, err := jsonMarshal(v)
	if err != nil {
		return err
	}
	return s.SendData(string(data))
}

// SendEvent sends an event with type and data.
func (s *SSEStream) SendEvent(eventType, data string) error {
	return s.Send(SSEEvent{
		Event: eventType,
		Data:  data,
	})
}

// SetKeepAlive sets the keep-alive interval.
func (s *SSEStream) SetKeepAlive(d time.Duration) {
	s.keepAlive = d
}

// StartKeepAlive starts sending keep-alive comments.
func (s *SSEStream) StartKeepAlive() {
	if s.keepAlive <= 0 {
		return
	}

	ticker := time.NewTicker(s.keepAlive)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// Send keep-alive comment
				s.ctx.Res.Write([]byte(": keep-alive\n\n"))
				if flusher, ok := s.ctx.Res.(http.Flusher); ok {
					flusher.Flush()
				}
			case <-s.keepAliveDone:
				return
			}
		}
	}()
}

// StopKeepAlive stops the keep-alive goroutine.
func (s *SSEStream) StopKeepAlive() {
	close(s.keepAliveDone)
}

// LastEventID returns the last sent event ID.
func (s *SSEStream) LastEventID() string {
	return s.lastEventID
}

// GetLastEventID returns the last event ID from the client (from Last-Event-ID header).
func (s *SSEStream) GetLastEventID() string {
	return s.ctx.Header("Last-Event-ID")
}

// SSE creates an SSE stream and calls the handler.
func (c *Context) SSE(handler func(*SSEStream)) {
	stream := NewSSEStream(c)
	handler(stream)
}

// SSEBroadcaster manages multiple SSE connections for broadcasting.
type SSEBroadcaster struct {
	clients map[chan SSEEvent]bool
	mu      sync.RWMutex
}

// NewSSEBroadcaster creates a new SSE broadcaster.
func NewSSEBroadcaster() *SSEBroadcaster {
	return &SSEBroadcaster{
		clients: make(map[chan SSEEvent]bool),
	}
}

// Register registers a new client.
func (b *SSEBroadcaster) Register(client chan SSEEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[client] = true
}

// Unregister unregisters a client.
func (b *SSEBroadcaster) Unregister(client chan SSEEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.clients[client]; ok {
		delete(b.clients, client)
		close(client)
	}
}

// Broadcast broadcasts an event to all clients.
func (b *SSEBroadcaster) Broadcast(event SSEEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for client := range b.clients {
		select {
		case client <- event:
		default:
			// Client channel is full, skip
		}
	}
}

// Count returns the number of connected clients.
func (b *SSEBroadcaster) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

// ServeSSE serves SSE to a client.
func (b *SSEBroadcaster) ServeSSE(c *Context) {
	stream := NewSSEStream(c)
	stream.StartKeepAlive()
	defer stream.StopKeepAlive()

	// Create client channel
	client := make(chan SSEEvent, 10)
	b.Register(client)
	defer b.Unregister(client)

	// Send events to client
	for event := range client {
		if err := stream.Send(event); err != nil {
			return
		}
	}
}
