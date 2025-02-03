package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	clients   = make(map[*websocket.Conn]bool) // connected clients
	broadcast = make(chan Message)             // broadcast channel
	mu        sync.Mutex                       // mutex to protect shared resources
)

type Message struct {
	Type string `json:"type"`
	User string `json:"user"`
	Text string `json:"text,omitempty"`
}

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

	// Add the client to the list
	mu.Lock()
	clients[conn] = true
	mu.Unlock()
	fmt.Println("Client connected")

	// Read messages from the client
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		// Unmarshal the message into a Message struct
		var msg Message
		err = json.Unmarshal(msgBytes, &msg)
		if err != nil {
			fmt.Println("Error decoding message:", err)
			continue
		}

		// Print the received message
		fmt.Printf("Received message: %+v\n", msg)

		// Broadcast the message
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		// Get the next message from the broadcast channel
		msg := <-broadcast

		// Lock access to the clients map
		mu.Lock()
		for client := range clients {
			// Send the message to all connected clients
			err := client.WriteMessage(websocket.TextMessage, []byte(msg.Text))
			if err != nil {
				fmt.Println("Error sending message:", err)
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
	fmt.Printf("WebSocket server started at ws://localhost:%s\n", port)

	// Start the message handling goroutine
	go handleMessages()

	// Start the HTTP server
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
