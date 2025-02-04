package socket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

// Message struct represents the JSON format for messages
type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

// Handle WebSocket connection
func handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()


	log.Println("New client connected")

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		log.Print(msg)

		log.Printf("[%s]: %s", msg.Username, msg.Message)
		broadcastMessage(msg, conn)
	}

	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()
}

// Broadcast message to all clients except the sender
func broadcastMessage(msg Message, sender *websocket.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		if client == sender {
			continue
		}
		err := client.WriteJSON(msg)
		if err != nil {
			log.Println("Broadcast error:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// Run starts the WebSocket server
func Run(address, websocketPath string) {
	http.HandleFunc(websocketPath, handleConnection)



	log.Printf("WebSocket server running on ws://%s%s", address, websocketPath)
	log.Fatal(http.ListenAndServe(address, nil))
}
