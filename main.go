package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Upgrader configures the WebSocket connection.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins (adjust for production use)
	},
}

var (
	clients   = make(map[*websocket.Conn]bool) // Active WebSocket clients
	broadcast = make(chan Message)             // Channel for broadcasting messages
	mu        sync.Mutex                        // Mutex to protect concurrent access
)

// Message structure
type Message struct {
	Type string `json:"type"` // "message" or "typing"
	User string `json:"user"`
	Text string `json:"text,omitempty"` // Only for normal messages
}

// HandleWebSocket handles WebSocket requests from clients.
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	// Register the client
	mu.Lock()
	clients[conn] = true
	mu.Unlock()

	fmt.Println("Client connected")

	// Listen for messages from the client
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		// Decode the incoming message
		var msg Message
		err = json.Unmarshal(msgBytes, &msg)
		if err != nil {
			fmt.Println("Error decoding message:", err)
			continue
		}

		// Broadcast the message to all clients
		broadcast <- msg
	}

	// Remove the client when they disconnect
	mu.Lock()
	delete(clients, conn)
	mu.Unlock()

	fmt.Println("Client disconnected")
}

// Broadcast messages to all connected clients
func handleMessages() {
	for {
		msg := <-broadcast

		mu.Lock()
		for client := range clients {
			msgBytes, err := json.Marshal(msg)
			if err != nil {
				fmt.Println("Error encoding message:", err)
				continue
			}

			err = client.WriteMessage(websocket.TextMessage, msgBytes)
			if err != nil {
				fmt.Println("Error writing message:", err)
				client.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}
}

func main() {
	http.HandleFunc("/ws", HandleWebSocket)

	port := "8080"
	fmt.Printf("WebSocket server started at ws://localhost:%s/ws\n", port)

	// Start the message broadcasting routine
	go handleMessages()

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
