package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

type webSocketHandler struct {
	upgrader websocket.Upgrader
}

func (wsh webSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := wsh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error %s when upgrading connection to websocket", err)
		return
	}
	defer c.Close()
}

func printLogo() {
	data, err := os.ReadFile("assets/logo.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	fmt.Println(string(data))
}

func main() {

	printLogo()

	webSocketHandler := webSocketHandler{
		upgrader: websocket.Upgrader{},
	}
	http.Handle("/", webSocketHandler)
	log.Println("Starting websocket on localhost:8080..")
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
