package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

type ChatServer struct {
	activeUsers map[*websocket.Conn]string // Track active users by connection
	mu          sync.Mutex                 // Mutex to protect concurrent access to activeUsers
}

func NewChatServer() *ChatServer {
	return &ChatServer{
		activeUsers: make(map[*websocket.Conn]string),
	}
}

func (server *ChatServer) handleConnection(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		fmt.Println("Failed to upgrade connection:", err)
		return
	}
	defer conn.Close()
	log.Println("Client connected")

	for {
		// Read a message from the client
		_, msg, err := conn.ReadMessage()
		if err != nil {
			// Handle error or if connection is closed
			log.Println("Error reading message:", err)
			break
		}

		// Parse JSON (just for the sake of example, assuming the message is simple)
		var message map[string]interface{}
		if err := json.Unmarshal(msg, &message); err != nil {
			log.Println("Error unmarshalling JSON:", err)
			continue
		}

		// Log the received message (optional)
		log.Printf("Received: %v\n", message)

		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("Error sending message:", err)
			break
		}

		log.Println("Message sent back to client")
	}

	log.Println("Client disconnected")
}

func getPublicIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80") // Use a dummy connection to get the public IP
	if err != nil {
		return "Unknown"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func (server *ChatServer) Start() {
	http.HandleFunc("/", server.handleConnection)

	// Get the PORT environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Fallback for local testing
	}

	// Bind to 0.0.0.0 and listen on the Railway-provided port
	fmt.Printf("Server started at port %s...\n", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func main() {
	server := NewChatServer()
	server.Start()
}
