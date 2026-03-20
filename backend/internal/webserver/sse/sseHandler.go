package sse_handler

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/r3labs/sse/v2"
)

var server *sse.Server
var mutex sync.Mutex

// RegisterRoute registers the SSE endpoint on the provided Gin router.
// The r3labs/sse server is wrapped with gin.WrapH so it integrates seamlessly
// with the Gin middleware chain.
func RegisterRoute(router *gin.Engine, path string) {
	mutex.Lock()
	defer mutex.Unlock()

	if server == nil {
		server = sse.New()
		server.CreateStream("messages")
	}

	router.GET(path, gin.WrapH(http.HandlerFunc(server.ServeHTTP)))

	go func() {
		// Send a reconnect event after a short delay so clients that connect
		// immediately after startup receive at least one event.
		time.Sleep(3 * time.Second)
		SendMessage("reconnect", "_")
	}()
}

// SendMessage publishes an SSE event to all connected clients.
func SendMessage(event string, message string) {
	mutex.Lock()
	defer mutex.Unlock()

	if server == nil {
		log.Println("SSE server not started")
		return
	}

	server.Publish("messages", &sse.Event{
		Event: []byte(event),
		Data:  []byte(message),
	})
}
