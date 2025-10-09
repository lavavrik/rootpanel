package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/lavavrik/go-sm/api"
	"github.com/lavavrik/go-sm/api/kv"
	"github.com/lavavrik/go-sm/stats"
)

// Config represents the application configuration
type Config struct {
	Server ServerConfig   `json:"server"`
	Stats  StatsConfig    `json:"stats"`
	Redis  kv.RedisConfig `json:"redis"`
}

// ServerConfig represents server-specific configuration
type ServerConfig struct {
	Port string `json:"port"`
}

type StatsConfig struct {
	Enabled         bool   `json:"enabled"`
	IntervalSeconds int    `json:"interval_seconds"`
	StatsFolder     string `json:"stats_folder"`
}

// loadConfig loads configuration from config.json file
func loadConfig() (*Config, error) {
	configFile := "config.json"

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return &Config{
			Server: ServerConfig{
				Port: "8080",
			},
		}, nil
	}

	// Read config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return &config, nil
}

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Register API handlers
	api.RegisterHandlers()

	// Get port from config, with environment variable override
	port := config.Server.Port
	log.Printf("Using port from config file: %s", port)

	// Start server
	addr := "127.0.0.1:" + port
	log.Printf("Starting server on %s", addr)
	log.Printf("API endpoints:")
	log.Printf("  GET /        - API documentation")
	log.Printf("  GET /health  - Health check")
	log.Printf("  GET /files   - List files")

	if config.Stats.Enabled {
		ticker := time.NewTicker(time.Duration(config.Stats.IntervalSeconds) * time.Second)

		log.Println("Stats collection is enabled.")
		done := make(chan bool)
		go func() {
			// 4. Use a for-select loop to wait for ticks or a stop signal.
			for {
				select {
				case <-done:
					log.Println("Goroutine is stopping!")
					return
				case <-ticker.C:
					stats.WriteDataPoints()
				}
			}
		}()
	} else {
		log.Println("Stats collection is disabled.")
	}

	fmt.Println("Redis configuration:", config.Redis)

	err = kv.Connect(config.Redis)
	if err != nil {
		log.Fatal("Failed to connect to Redis: ", err)
	} else {
		log.Println("Connected to Redis successfully.")
	}

	defer kv.Close()

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}
