package sse

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
)

// Map to store SSE client connections
var (
	clients = make(map[chan uint8]bool)
	mutex   = sync.Mutex{}
)

// Handle SSE connection
func SseHandler(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	messageChan := make(chan uint8)
	mutex.Lock()
	clients[messageChan] = true
	mutex.Unlock()

	// Send updates while connection is open
	for {
		select {
		case msg := <-messageChan:
			fmt.Fprintf(c.Writer, "data: %d\n\n", msg)
			c.Writer.Flush()
		case <-c.Writer.CloseNotify():
			mutex.Lock()
			delete(clients, messageChan)
			mutex.Unlock()
			close(messageChan)
			return
		}
	}
}

// Notify all connected clients
func Notify(message uint8) {
	mutex.Lock()
	defer mutex.Unlock()
	for client := range clients {
		client <- message
	}
}
