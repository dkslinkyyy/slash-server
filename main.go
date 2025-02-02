package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader configures the WebSocket connection.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins (adjust for production use)
	},
}

// HandleWebSocket handles WebSocket requests from clients.
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Client connected")

	// Listen for messages from the client.
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		fmt.Printf("Received: %s\n", message)

		// Echo the message back to the client.
		if err := conn.WriteMessage(messageType, message); err != nil {
			fmt.Println("Error writing message:", err)
			break
		}
	}

	fmt.Println("Client disconnected")
}

func main() {
	http.HandleFunc("/ws", HandleWebSocket)

	port := "8080"
	fmt.Printf("WebSocket server started at ws://localhost:%s/ws\n", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
