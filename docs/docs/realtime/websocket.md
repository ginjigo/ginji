# WebSocket

Build real-time bidirectional communication with WebSocket support.

## Basic Usage

```go
app.Get("/ws", func(c *ginji.Context) {
    c.WebSocket(func(conn *ginji.WebSocketConn) {
        defer conn.Close()
        
        for {
            messageType, data, err := conn.ReadMessage()
            if err != nil {
                break
            }
            
            // Echo back
            conn.WriteMessage(messageType, data)
        }
    })
})
```

## Manual Upgrade

For more control over the upgrade process:

```go
upgrader := ginji.NewWebSocketUpgrader(ginji.DefaultWebSocketConfig())

app.Get("/ws", func(c *ginji.Context) {
    conn, err := upgrader.Upgrade(c)
    if err != nil {
        c.JSON(ginji.StatusBadRequest, ginji.H{"error": err.Error()})
        return
    }
    defer conn.Close()

    // Handle connection
    handleWebSocket(conn)
})
```

## Configuration

```go
config := ginji.WebSocketConfig{
    ReadBufferSize:   8192,
    WriteBufferSize:  8192,
    HandshakeTimeout: 10 * time.Second,
    
    CheckOrigin: func(c *ginji.Context) bool {
        origin := c.Header("Origin")
        return origin == "https://yourdomain.com"
    },
}

upgrader := ginji.NewWebSocketUpgrader(config)
```

## JSON Messages

Send and receive JSON easily:

```go
type Message struct {
    Type string `json:"type"`
    Data string `json:"data"`
}

conn.WriteJSON(Message{
    Type: "notification",
    Data: "Hello!",
})

var msg Message
conn.ReadJSON(&msg)
```

## Broadcasting with Hub

The Hub manages multiple connections and broadcasting:

```go
hub := ginji.NewHub()
go hub.Run() // Start hub in background

app.Get("/chat", func(c *ginji.Context) {
    upgrader := ginji.NewWebSocketUpgrader(ginji.DefaultWebSocketConfig())
    conn, _ := upgrader.Upgrade(c)
    
    // Register connection
    hub.Register(conn)
    defer hub.Unregister(conn)
    
    // Read messages and broadcast
    for {
        _, msg, err := conn.ReadMessage()
        if err != nil {
            break
        }
        hub.Broadcast(msg) // Send to all connected clients
    }
})
```

## Complete Chat Example

```go
package main

import (
    "github.com/ginjigo/ginji"
)

func main() {
    app := ginji.New()
    hub := ginji.NewHub()
    go hub.Run()

    app.Get("/chat", func(c *ginji.Context) {
        upgrader := ginji.NewWebSocketUpgrader(ginji.DefaultWebSocketConfig())
        conn, err := upgrader.Upgrade(c)
        if err != nil {
            return
        }

        hub.Register(conn)
        defer hub.Unregister(conn)

        // Send welcome message
        conn.WriteJSON(ginji.H{
            "type": "system",
            "message": "Welcome to the chat!",
        })

        // Handle messages
        for {
            var msg map[string]any
            if err := conn.ReadJSON(&msg); err != nil {
                break
            }

            // Broadcast to all clients
            hub.Broadcast([]byte(msg["message"].(string)))
        }
    })

    app.Listen(":3000")
}
```

## Client Example

```javascript
const ws = new WebSocket('ws://localhost:3000/chat');

ws.onopen = () => {
    console.log('Connected!');
    ws.send(JSON.stringify({
        type: 'message',
        message: 'Hello, server!'
    }));
};

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log('Received:', data);
};

ws.onerror = (error) => {
    console.error('WebSocket error:', error);
};

ws.onclose = () => {
    console.log('Disconnected');
};
```

## Ping/Pong

Keep connections alive:

```go
go func() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        if err := conn.Ping(); err != nil {
            return
        }
    }
}()
```

## Message Types

- `ginji.TextMessage` - Text data
- `ginji.BinaryMessage` - Binary data
- `ginji.PingMessage` - Ping frame
- `ginji.PongMessage` - Pong frame
- `ginji.CloseMessage` - Close connection

## Best Practices

1. **Close connections** - Always defer `conn.Close()`
2. **Handle errors** - Check for read/write errors
3. **Use timeouts** - Set reasonable handshake timeouts
4. **Validate origin** - Implement `CheckOrigin` in production
5. **Use Hub** - For broadcasting to multiple clients
6. **Ping/Pong** - Keep connections alive
7. **Rate limiting** - Prevent message spam

## Use Cases

- 💬 Chat applications
- 🎮 Real-time games
- 📊 Live dashboards
- 🔔 Notifications
- 📝 Collaborative editing
- 📈 Live data streaming
