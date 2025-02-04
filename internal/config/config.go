package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Host          string `mapstructure:"host"`
		Port          int    `mapstructure:"port"`
		WebSocketPath string `mapstructure:"websocket_path"`
		LogLevel      string `mapstructure:"log_level"`
	} `mapstructure:"server"`
}

var AppConfig Config

func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // Look in root directory
	viper.AddConfigPath("./config") // Look in ./config directory
	viper.AutomaticEnv() // Allow env variables to override config values

	// Try to read the config file
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error loading config file: %v", err)
	}

	// Unmarshal the config into AppConfig
	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	// Debugging print
	fmt.Printf("Loaded Config: %+v\n", AppConfig)
}
