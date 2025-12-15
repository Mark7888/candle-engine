package sse_handler

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/r3labs/sse/v2"
)

var server *sse.Server
var mutex sync.Mutex

// StartServer starts the server
func StartServer(path string) *sse.Server {
	mutex.Lock()
	defer mutex.Unlock()

	if server == nil {
		server = sse.New()
		server.CreateStream("messages")

		http.HandleFunc(path, server.ServeHTTP)
	}

	go func() {
		// send a reload message after 3 seconds to ensure client connects properly
		time.Sleep(3 * time.Second)
		SendMessage("reconnect", "_")
	}()

	return server
}

// send message
func SendMessage(event string, message string) {
	mutex.Lock()
	defer mutex.Unlock()

	if server == nil {
		log.Println("Server not started")
		return
	}

	server.Publish("messages", &sse.Event{
		Event: []byte(event),
		Data:  []byte(message),
	})
}
