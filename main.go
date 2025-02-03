package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins (for testing only)
	},
}

var (
	clients   = make(map[*websocket.Conn]string) // Track connected clients and their usernames
	broadcast = make(chan Message)                // Channel for broadcasting messages
	mu        sync.Mutex                          // Mutex to prevent race conditions
)


// Message struct for WebSocket communication
type Message struct {
	Type string `json:"type"`           // "message" or "typing"
	User string `json:"user"`           // Username
	Text string `json:"text,omitempty"` // Chat message content (optional)
}

// WebSocket handler
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		return
	}
	defer func() {
		mu.Lock()
		delete(clients, conn)
		mu.Unlock()
		conn.Close()
		fmt.Println("Client disconnected")
	}()

	// Register the client and assign a username
	mu.Lock()
	clients[conn] = "" // Empty until username is set
	mu.Unlock()
	fmt.Println("Client connected")

	// Handle username setup (assume it's sent immediately after connection)
	var username string
	_, msgBytes, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("Error reading message:", err)
		return
	}
	// Assuming the first message is the username
	fmt.Println(msgBytes)
	fmt.Println(&username)

	err = json.Unmarshal(msgBytes, &username)
	if err != nil {
		fmt.Println("Error decoding username:", err)
		return
	}

	// Save the username
	mu.Lock()
	clients[conn] = username
	mu.Unlock()

	// Listen for messages
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break // Exit loop on error (client likely disconnected)
		}

		// Decode message
		var msg Message
		err = json.Unmarshal(msgBytes, &msg)
		if err != nil {
			fmt.Println("Error decoding message:", err)
			continue
		}

		// Debug: Log received message
		fmt.Printf("Received message: %+v\n", msg)

		// Send message to broadcast channel
		broadcast <- msg
	}
}


// Broadcast messages to all clients except the sender
func handleMessages() {
	for {
		msg := <-broadcast

		mu.Lock()
		for client, username := range clients {
			// Exclude the sender
			if username != msg.User { // Only broadcast to clients other than the sender
				msgBytes, err := json.Marshal(msg)
				if err != nil {
					fmt.Println("Error encoding message:", err)
					continue
				}

				// Debug: Log message before sending
				fmt.Printf("Broadcasting: %+v\n", msg)

				err = client.WriteMessage(websocket.TextMessage, msgBytes)
				if err != nil {
					fmt.Println("Error sending message:", err)
					client.Close()
					delete(clients, client)
				}
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
