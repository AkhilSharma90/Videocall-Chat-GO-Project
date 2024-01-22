package chat

import (
	"bytes"
	"log"
	"time"

	"github.com/fasthttp/websocket"
)

// Constants defining time durations and message sizes
const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Special characters
var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Upgrader configuration for WebSocket connection
var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client represents a WebSocket client
type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
}

// readPump continuously reads messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	// Set up read limits and deadlines
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		// Read message from the WebSocket connection
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			// Handle unexpected close errors
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Trim leading/trailing whitespaces and replace newlines
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.Hub.broadcast <- message
	}
}

// writePump continuously writes messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			// Set write deadline
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				return
			}

			// Get the next writer for sending a message
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			// Write the message to the WebSocket connection
			w.Write(message)

			// Iterate through remaining messages in the channel and write them
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
			}

			// Close the writer
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			// Set write deadline for sending a ping message
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// PeerChatConn creates a new Client and starts the writePump and readPump goroutines
func PeerChatConn(c *websocket.Conn, hub *Hub) {
	client := &Client{Hub: hub, Conn: c, Send: make(chan []byte, 256)}
	client.Hub.register <- client

	// Start the writePump and readPump goroutines
	go client.writePump()
	client.readPump()
}
