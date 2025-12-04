# Server-Sent Events (SSE)

Unidirectional server-to-client event streaming.

## Basic Usage

```go
app.Get("/events", func(c *ginji.Context) {
    c.SSE(func(stream *ginji.SSEStream) {
        for i := 0; i < 10; i++ {
            stream.SendData(fmt.Sprintf("Event %d", i))
            time.Sleep(time.Second)
        }
    })
})
```

## Send Events

```go
stream.Send(ginji.SSEEvent{
    ID:    "1",
    Event: "update",
    Data:  "New data available",
})
```

## Send JSON

```go
stream.SendJSON(ginji.H{
    "type": "notification",
    "message": "Hello!",
})
```

## Event Types

```go
stream.SendEvent("notification", "You have a new message")
stream.SendEvent("update", "Data refreshed")
```

## Keep-Alive

Prevent connection timeouts:

```go
c.SSE(func(stream *ginji.SSEStream) {
    stream.StartKeepAlive() // Sends keep-alive comments
    defer stream.StopKeepAlive()
    
    // Your event logic...
})
```

## Broadcasting

Broadcast to multiple clients:

```go
broadcaster := ginji.NewSSEBroadcaster()

// Endpoint for clients
app.Get("/stream", func(c *ginji.Context) {
    broadcaster.ServeSSE(c)
})

// Send from anywhere
broadcaster.Broadcast(ginji.SSEEvent{
    Event: "notification",
    Data:  "Server update!",
})
```

## Live Dashboard Example

```go
package main

import (
    "fmt"
    "time"
    "github.com/ginjigo/ginji"
)

var broadcaster = ginji.NewSSEBroadcaster()

func main() {
    app := ginji.New()

    //  SSE endpoint
    app.Get("/dashboard", func(c *ginji.Context) {
        broadcaster.ServeSSE(c)
    })

    // Start background updates
    go sendUpdates()

    app.Listen(":3000")
}

func sendUpdates() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        stats := getSystemStats()
        
        broadcaster.Broadcast(ginji.SSEEvent{
            ID:    fmt.Sprintf("%d", time.Now().Unix()),
            Event: "stats",
            Data:  stats,
        })
    }
}

func getSystemStats() string {
    return `{"cpu": 45, "memory": 60, "connections": 125}`
}
```

## Client Example

```javascript
const eventSource = new EventSource('http://localhost:3000/dashboard');

eventSource.addEventListener('stats', (event) => {
    const data = JSON.parse(event.data);
    updateDashboard(data);
});

eventSource.addEventListener('notification', (event) => {
    showNotification(event.data);
});

eventSource.onerror = (error) => {
    console.error('SSE error:', error);
};
```

## Resume on Reconnect

SSE supports automatic reconnection with last event ID:

```go
c.SSE(func(stream *ginji.SSEStream) {
    // Get last event ID from client
    lastID := stream.GetLastEventID()
    
    // Resume from that point
    events := getEventsSince(lastID)
    for _, event := range events {
        stream.Send(event)
    }
})
```

## Configuration

```go
stream := ginji.NewSSEStream(c)
stream.SetKeepAlive(15 * time.Second) // Custom keep-alive interval
stream.StartKeepAlive()
```

## Features

- ✅ Standard SSE protocol
- ✅ Event types and IDs
- ✅ Multi-line data support
- ✅ Keep-alive mechanism
- ✅ Broadcaster for multiple clients
- ✅ Last-Event-ID tracking
- ✅ Auto-reconnect support

## Best Practices

1. **Use Keep-Alive** - Prevent proxy/firewall timeouts
2. **Set Event IDs** - Enable client resumption
3. **Use Broadcaster** - For many-to-many communication
4. **Handle Disconnects** - Clean up resources
5. **Set Retry** - Guide client reconnection timing
6. **Compress** - Use gzip for large data

## Use Cases

- 📊 Live dashboards
- 🔔 Real-time notifications
- 📈 Stock price updates
- 🎯 Progress tracking
- 📰 News feeds
- 🏃 Activity streams

## SSE vs WebSocket

**Use SSE when:**
- Unidirectional (server → client)
- Simple setup needed
- Browser compatibility important
- Auto-reconnect wanted

**Use WebSocket when:**
- Bidirectional communication needed
- Lower latency required
- Binary data transmission
- Custom protocols needed
