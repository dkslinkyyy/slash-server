package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"golang.org/x/exp/rand"
)

// Config struct for holding configuration
type Config struct {
	Local   bool `mapstructure:"local"`
	Service struct {
		Name string `mapstructure:"name"`
		ID   string `mapstructure:"id"`
	} `mapstructure:"service"`
	Server struct {
		URL  string `mapstructure:"URL"`
		Port string `mapstructure:"port"`
	} `mapstructure:"server"`
	Proxy struct {
		RegistrationURL string `mapstructure:"registration_url"`
	} `mapstructure:"proxy"`
}

// ChatServer struct for WebSocket server
type ChatServer struct {
	activeUsers map[*websocket.Conn]string
	mu          sync.Mutex
	config      Config
}

func LoadConfig(path string) (Config, error) {
	var config Config
	viper.SetConfigName("config") // Name of the config file (without extension)
	viper.SetConfigType("yaml")   // File format
	viper.AddConfigPath(path)     // Path to look for the config file

	if err := viper.ReadInConfig(); err != nil {
		return config, fmt.Errorf("error reading config file: %v", err)
	}
	if err := viper.Unmarshal(&config); err != nil {
		return config, fmt.Errorf("error unmarshalling config: %v", err)
	}

	// Resolve placeholders if `local` is false
	if !config.Local {
		if envName, exists := os.LookupEnv(config.Service.Name); exists {
			config.Service.Name = envName
		}
		if envID, exists := os.LookupEnv(config.Service.ID); exists {
			config.Service.ID = envID
		}
		//		if envPublicURL, exists := os.LookupEnv(config.Server.URL); exists {
		//			config.Service.URL = envPublicURL + "/connect/" + config.Service.ID
		//		}
	} else {
		config.Service.Name = "local_service"
		config.Service.ID = string(generateRandomID())
		config.Server.URL = "localhost:" + config.Server.Port + "/connect/" + config.Service.ID
		log.Printf("Generated ID: %s", config.Service.ID) // Debugging line
	}
	return config, nil
}

func generateRandomID() string {
	// Ensure rand source is seeded for better randomness
	source := rand.NewSource(uint64(time.Now().UnixNano()))
	randGen := rand.New(source)

	// Create a byte slice to store the random ID
	randomBytes := make([]byte, 16) // 16 bytes for a 128-bit ID

	// Generate random bytes
	_, err := randGen.Read(randomBytes)
	if err != nil {
		fmt.Println("Error generating random bytes:", err)
		return ""
	}

	// Return the ID as a hexadecimal string
	return hex.EncodeToString(randomBytes)
}

// Function to register the server
func (server *ChatServer) registerServer() error {
	// Prepare JSON body for registration
	serverInfo := map[string]string{
		"serviceName":       server.config.Service.Name,
		"serviceIdentifier": server.config.Service.ID,
		"websocketURL":      server.config.Server.URL,
		"available":         "true",
	}

	// Marshal the server info to JSON
	data, err := json.Marshal(serverInfo)
	if err != nil {
		return fmt.Errorf("error marshalling server info: %v", err)
	}

	// Define the registration URL based on the local flag
	var registrationURL string
	if server.config.Local {
		// If local is true, send request to localhost:3000
		registrationURL = "http://localhost:3030/webservers/register"
	} else {
		// Otherwise, use the configured proxy URL
		registrationURL = server.config.Proxy.RegistrationURL
	}

	// Call the sendRegisterRequest function to send the POST request
	err = sendRegisterRequest(registrationURL, data)
	if err != nil {
		return fmt.Errorf("error sending registration request: %v", err)
	}

	log.Printf("Server registered with response status: OK")
	return nil
}

func sendRegisterRequest(url string, data []byte) error {
	// Send the POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("error sending POST request: %v", err)
	}
	defer resp.Body.Close()

	// Check if the response status is not OK
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	return nil
}

// Function to handle client WebSocket connections
func (server *ChatServer) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}
	defer conn.Close()
	log.Println("Client connected")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		log.Printf("Received: %s", msg)
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("Error sending message:", err)
			break
		}
	}
	log.Println("Client disconnected")
}

func (server *ChatServer) Start() {
	// Register the server upon startup
	if err := server.registerServer(); err != nil {
		log.Fatalf("Error registering server: %v", err)
	}

	r := mux.NewRouter()

	serviceID := server.config.Service.ID

	r.HandleFunc("/connect/"+serviceID, server.handleConnection)
	port := server.config.Server.Port
	log.Printf("Server started at port %s...\n", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func main() {
	config, err := LoadConfig(".")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	server := &ChatServer{
		activeUsers: make(map[*websocket.Conn]string),
		config:      config,
	}
	server.Start()
}
