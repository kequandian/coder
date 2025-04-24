package app

import (
	"fmt"
	"net/http"
	"time"

	"coder/config"
)

// App represents the application container for shared resources
var (
	Config       *config.Config
	ConfigClient *HTTPClient
)

// HTTPClient represents a configured HTTP client with its own base URL
type HTTPClient struct {
	Client  *http.Client
	BaseURL string
}

// Initialize initializes the application
func Init() {
	// Load configuration
	config, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %s", err.Error()))
	}
	Config = config

	// Create HTTP clients
	ConfigClient = &HTTPClient{
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		BaseURL: config.HTTPClient.Config,
	}
}
