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
	clients   = make(map[*websocket.Conn]bool) 
	broadcast = make(chan Message)             
	mu        sync.Mutex                       
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

	
	mu.Lock()
	clients[conn] = true
	mu.Unlock()
	fmt.Println("Client connected")

	
	var msgBytes []byte
	_, msgBytes, err = conn.ReadMessage()
	if err != nil {
		fmt.Println("Error reading message:", err)
		return
	}

	
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break 
		}

		
		var msg Message
		err = json.Unmarshal(msgBytes, &msg)
		if err != nil {
			fmt.Println("Error decoding message:", err)
			continue
		}

		
		fmt.Printf("Received message: %+v\n", msg)

		
		broadcast <- msg
	}
}


func handleMessages() {
	for {
		msg := <-broadcast

		mu.Lock()
		for client := range clients {
			
			err := client.WriteMessage(websocket.TextMessage, msgBytes)
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
	fmt.Printf("WebSocket server started at ws:

	
	go handleMessages()

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
