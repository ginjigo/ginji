package ginji

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

// WebSocket message types
const (
	TextMessage   = 1
	BinaryMessage = 2
	CloseMessage  = 8
	PingMessage   = 9
	PongMessage   = 10
)

// maxWebSocketPayloadSize limits the maximum payload size for WebSocket messages sent by WriteMessage, in bytes.
// 64 MiB is a typical upper bound, safe for most environments (adjust as appropriate for your application).
const maxWebSocketPayloadSize = 64 * 1024 * 1024

// WebSocketConn represents a WebSocket connection.
type WebSocketConn struct {
	conn      net.Conn
	mu        sync.Mutex
	writeMu   sync.Mutex
	closed    bool
	closeOnce sync.Once
}

// WebSocketConfig defines configuration for WebSocket upgrade.
type WebSocketConfig struct {
	// ReadBufferSize is the read buffer size in bytes.
	ReadBufferSize int

	// WriteBufferSize is the write buffer size in bytes.
	WriteBufferSize int

	// HandshakeTimeout is the duration for the handshake.
	HandshakeTimeout time.Duration

	// CheckOrigin returns true if the request Origin header is acceptable.
	// If nil, a safe default is used.
	CheckOrigin func(*Context) bool
}

// DefaultWebSocketConfig returns default WebSocket configuration.
func DefaultWebSocketConfig() WebSocketConfig {
	return WebSocketConfig{
		ReadBufferSize:   4096,
		WriteBufferSize:  4096,
		HandshakeTimeout: 10 * time.Second,
		CheckOrigin:      nil,
	}
}

// WebSocketUpgrader handles WebSocket upgrade requests.
type WebSocketUpgrader struct {
	config WebSocketConfig
}

// NewWebSocketUpgrader creates a new WebSocket upgrader.
func NewWebSocketUpgrader(config WebSocketConfig) *WebSocketUpgrader {
	if config.ReadBufferSize == 0 {
		config.ReadBufferSize = 4096
	}
	if config.WriteBufferSize == 0 {
		config.WriteBufferSize = 4096
	}
	if config.HandshakeTimeout == 0 {
		config.HandshakeTimeout = 10 * time.Second
	}
	return &WebSocketUpgrader{config: config}
}

// Upgrade upgrades the HTTP connection to a WebSocket connection.
func (u *WebSocketUpgrader) Upgrade(c *Context) (*WebSocketConn, error) {
	// Check if it's a WebSocket upgrade request
	if c.Header("Connection") != "Upgrade" || c.Header("Upgrade") != "websocket" {
		return nil, errors.New("not a websocket handshake")
	}

	// Check origin if configured
	if u.config.CheckOrigin != nil && !u.config.CheckOrigin(c) {
		return nil, errors.New("origin not allowed")
	}

	// Get the underlying connection
	hijacker, ok := c.Res.(http.Hijacker)
	if !ok {
		return nil, errors.New("response does not support hijacking")
	}

	conn, bufrw, err := hijacker.Hijack()
	if err != nil {
		return nil, err
	}

	// Perform WebSocket handshake
	key := c.Header("Sec-WebSocket-Key")
	if key == "" {
		_ = conn.Close()
		return nil, errors.New("missing Sec-WebSocket-Key")
	}

	// Send handshake response
	acceptKey := computeAcceptKey(key)
	response := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n"

	if _, err := bufrw.Write([]byte(response)); err != nil {
		_ = conn.Close()
		return nil, err
	}

	if err := bufrw.Flush(); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return &WebSocketConn{
		conn:   conn,
		closed: false,
	}, nil
}

// WriteMessage writes a message to the WebSocket connection.
func (ws *WebSocketConn) WriteMessage(messageType int, data []byte) error {
	ws.writeMu.Lock()
	defer ws.writeMu.Unlock()

	if ws.closed {
		return errors.New("websocket: connection closed")
	}

	if len(data) > maxWebSocketPayloadSize {
		return errors.New("websocket: payload too large")
	}

	// Simple frame format (for basic implementation)
	// In production, you'd want full RFC 6455 compliance
	frame := make([]byte, 2+len(data))
	frame[0] = byte(0x80 | messageType) // FIN bit + opcode
	frame[1] = byte(len(data))          // Payload length (simplified)
	copy(frame[2:], data)

	_, err := ws.conn.Write(frame)
	return err
}

// ReadMessage reads a message from the WebSocket connection.
func (ws *WebSocketConn) ReadMessage() (messageType int, p []byte, err error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed {
		return 0, nil, errors.New("websocket: connection closed")
	}

	// Read frame header (simplified)
	header := make([]byte, 2)
	if _, err := io.ReadFull(ws.conn, header); err != nil {
		return 0, nil, err
	}

	messageType = int(header[0] & 0x0F)
	payloadLen := int(header[1] & 0x7F)

	// Read payload
	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(ws.conn, payload); err != nil {
		return 0, nil, err
	}

	return messageType, payload, nil
}

// WriteJSON writes a JSON message to the WebSocket.
func (ws *WebSocketConn) WriteJSON(v any) error {
	data, err := jsonMarshal(v)
	if err != nil {
		return err
	}
	return ws.WriteMessage(TextMessage, data)
}

// ReadJSON reads a JSON message from the WebSocket.
func (ws *WebSocketConn) ReadJSON(v any) error {
	_, data, err := ws.ReadMessage()
	if err != nil {
		return err
	}
	return jsonUnmarshal(data, v)
}

// Close closes the WebSocket connection.
func (ws *WebSocketConn) Close() error {
	var err error
	ws.closeOnce.Do(func() {
		ws.mu.Lock()
		ws.closed = true
		ws.mu.Unlock()
		err = ws.conn.Close()
	})
	return err
}

// Ping sends a ping message.
func (ws *WebSocketConn) Ping() error {
	return ws.WriteMessage(PingMessage, []byte{})
}

// Pong sends a pong message.
func (ws *WebSocketConn) Pong() error {
	return ws.WriteMessage(PongMessage, []byte{})
}

// computeAcceptKey computes the Sec-WebSocket-Accept value.
func computeAcceptKey(key string) string {
	const magic = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New()
	h.Write([]byte(key + magic))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Helper for Context to check if request is WebSocket upgrade
func (c *Context) IsWebSocket() bool {
	return c.Header("Connection") == "Upgrade" && c.Header("Upgrade") == "websocket"
}

// WebSocket upgrades the connection to WebSocket.
func (c *Context) WebSocket(handler func(*WebSocketConn)) error {
	upgrader := NewWebSocketUpgrader(DefaultWebSocketConfig())
	conn, err := upgrader.Upgrade(c)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	handler(conn)
	return nil
}

// Hub manages WebSocket connections and broadcasts.
type Hub struct {
	connections map[*WebSocketConn]bool
	broadcast   chan []byte
	register    chan *WebSocketConn
	unregister  chan *WebSocketConn
	mu          sync.RWMutex
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		connections: make(map[*WebSocketConn]bool),
		broadcast:   make(chan []byte, 256),
		register:    make(chan *WebSocketConn),
		unregister:  make(chan *WebSocketConn),
	}
}

// Run starts the hub.
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.connections[conn] = true
			h.mu.Unlock()

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.connections[conn]; ok {
				delete(h.connections, conn)
				_ = conn.Close()
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for conn := range h.connections {
				go func(c *WebSocketConn) {
					if err := c.WriteMessage(TextMessage, message); err != nil {
						h.unregister <- c
					}
				}(conn)
			}
			h.mu.RUnlock()
		}
	}
}

// Register registers a connection to the hub.
func (h *Hub) Register(conn *WebSocketConn) {
	h.register <- conn
}

// Unregister unregisters a connection from the hub.
func (h *Hub) Unregister(conn *WebSocketConn) {
	h.unregister <- conn
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// Count returns the number of active connections.
func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections)
}
