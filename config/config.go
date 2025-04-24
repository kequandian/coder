package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

// Config holds the application configuration
type Config struct {
	Server     ServerConfig `toml:"server"`
	OpenAI     OpenAIConfig `toml:"openai"`
	Chat       ChatConfig   `toml:"chat"`
	LogPath    string       `toml:"log_path"`
	MCP        MCPConfig    `toml:"mcp"`
	HTTPClient HttpClient   `toml:"httpclient"`
}

// ServerConfig contains server configuration
type ServerConfig struct {
	Port            int           `toml:"port"`
	Host            string        `toml:"host"`
	EnableCORS      bool          `toml:"enable_cors"`
	ShutdownTimeout time.Duration `toml:"shutdown_timeout"`
}

// OpenAIConfig contains OpenAI API configuration
type OpenAIConfig struct {
	APIKey    string `toml:"api_key"`
	ModelID   string `toml:"model_id"`
	BaseURL   string `toml:"base_url"`
	MaxTokens int    `toml:"max_tokens"`
}

// ChatConfig contains chat configuration
type ChatConfig struct {
	SystemPrompt     string `toml:"system_prompt"`
	MaxHistoryLength int    `toml:"max_history_length"`
}

// MCPConfig contains Model Control Protocol configuration
type MCPConfig struct {
	Enabled bool        `toml:"enabled"`
	Clients []MCPClient `toml:"clients"`
}

// MCPClient contains configuration for an individual MCP client
type MCPClient struct {
	Name        string `toml:"name"`
	Enabled     bool   `toml:"enabled"`
	URL         string `toml:"url"`
	Description string `toml:"description"`
}

type HttpClient struct {
	Config string `toml:"config"`
}

// LoadConfig loads configuration from file and command line flags
func LoadConfig() (*Config, error) {
	config := &Config{}
	// Create config directory if it doesn't exist
	configDir := filepath.Dir("config.toml")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}
	}
	// Load existing config
	if _, err := toml.DecodeFile("config.toml", config); err != nil {
		panic(fmt.Sprintf("failed to decode config file: %s", err.Error()))
	}

	return config, nil
}
