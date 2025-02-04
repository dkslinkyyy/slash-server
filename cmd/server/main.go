package main

import (
	"fmt"
	"log"
	"os"

	"slashserver/internal/config"
	"slashserver/internal/socket"
)

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

	// Load configuration
	config.LoadConfig()

	// Validate Config
	if config.AppConfig.Server.Host == "" || config.AppConfig.Server.Port == 0 || config.AppConfig.Server.WebSocketPath == "" {
		log.Fatal("Invalid configuration. Please check config.yml or environment variables.")
	}

	address := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	websocketPath := config.AppConfig.Server.WebSocketPath


	log.Printf("test")
	log.Printf("Starting WebSocket server at ws://%s%s", address, websocketPath)

	// Start WebSocket server
	socket.Run(address, websocketPath)
}