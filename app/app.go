package app

import (
	"fmt"
	"net/http"
	"time"

	"coder/config"
)

// App represents the application container for shared resources
var (
	// Config 存储应用程序的全局配置信息
	Config *config.Config
	// ConfigClient 是用于与配置服务通信的HTTP客户端
	ConfigClient *HTTPClient
)

// HTTPClient represents a configured HTTP client with its own base URL
type HTTPClient struct {
	// Client 是标准HTTP客户端实例
	Client *http.Client
	// BaseURL 是HTTP客户端的基础URL
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
