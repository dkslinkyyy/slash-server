package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

type ChatServer struct {
	activeUsers map[*websocket.Conn]string // Track active users by connection
	mu          sync.Mutex                 // Mutex to protect concurrent access to activeUsers
}

type ServerInfo struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

func NewChatServer() *ChatServer {
	return &ChatServer{
		activeUsers: make(map[*websocket.Conn]string),
	}
}

// Function to register the server with environment variables
func (server *ChatServer) registerServer() error {
	// Prepare JSON body for registration
	serverInfo := ServerInfo{
		Name:       os.Getenv("RAILWAY_SERVICE_NAME"), // Assuming NAME is an environment variable
		Identifier: os.Getenv("RAILWAY_SERVICE_ID"),   // Assuming IDENTIFIER is an environment variable
	}

	// Convert the struct to JSON
	data, err := json.Marshal(serverInfo)
	if err != nil {
		return fmt.Errorf("error marshalling server info: %v", err)
	}

	// Send POST request to the registration endpoint
	resp, err := http.Post("http://slash-proxy-production.up.railway.app/webservers/register", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("error sending POST request: %v", err)
	}
	defer resp.Body.Close()

	// Log the response
	log.Printf("Server registered with response status: %s", resp.Status)
	return nil
}

// Function to handle client WebSocket connections
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

// Handle connection request (GET request)
func (server *ChatServer) handleConnectionRequest(w http.ResponseWriter, r *http.Request) {
	// Just a simple example of connection request logic
	// In a real-world scenario, you'd check if the server is available and respond accordingly
	log.Println("Received connection request")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Connection successful")
}

func (server *ChatServer) Start() {
	// Register the server upon startup
	if err := server.registerServer(); err != nil {
		log.Fatalf("Error registering server: %v", err)
	}

	// Set up the HTTP routes
	http.HandleFunc("/webserver/requestConnection", server.handleConnectionRequest)
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
