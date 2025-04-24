package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"coder/app"
)

// MCPClient represents a client for a specific MCP server
type MCPClient struct {
	name        string
	description string
	url         string
	client      *client.SSEMCPClient
	tools       []tool.BaseTool
}

// MCPManager manages multiple MCP clients
type MCPManager struct {
	clients map[string]*MCPClient
	tools   map[string]tool.BaseTool
	mu      sync.RWMutex
}

// NewMCPManager creates a new MCP manager with the provided configuration
func NewMCPManager() *MCPManager {
	return &MCPManager{
		clients: make(map[string]*MCPClient),
		tools:   make(map[string]tool.BaseTool),
	}
}

// Initialize initializes all enabled MCP clients
func (m *MCPManager) Initialize(ctx context.Context) error {
	if !app.Config.MCP.Enabled {
		log.Println("MCP is disabled in configuration")
		return nil
	}

	// Initialize each enabled client
	for _, clientCfg := range app.Config.MCP.Clients {
		if !clientCfg.Enabled {
			log.Printf("MCP client %s is disabled, skipping", clientCfg.Name)
			continue
		}

		log.Printf("Initializing MCP client: %s (%s)", clientCfg.Name, clientCfg.URL)

		// Create a new MCP client
		mcpClient := &MCPClient{
			name:        clientCfg.Name,
			description: clientCfg.Description,
			url:         clientCfg.URL,
		}

		// Connect to MCP server
		err := mcpClient.connect(ctx)
		if err != nil {
			log.Printf("Failed to connect to MCP server %s: %v", clientCfg.Name, err)
			continue
		}

		log.Printf("Connected to MCP server %s", clientCfg.Name)
		// Store the client
		m.mu.Lock()
		m.clients[clientCfg.Name] = mcpClient
		m.mu.Unlock()

		// Register all tools from this client
		m.registerTools(ctx, mcpClient)
	}

	// Log the number of clients and tools
	m.mu.RLock()
	defer m.mu.RUnlock()

	log.Printf("MCP initialization complete: %d clients connected, %d tools available",
		len(m.clients), len(m.tools))

	return nil
}

// GetAllTools returns all available tools from all connected MCP clients
func (m *MCPManager) GetAllTools() []tool.BaseTool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tools := make([]tool.BaseTool, 0, len(m.tools))
	for _, t := range m.tools {
		tools = append(tools, t)
	}
	return tools
}

// GetClientByName returns a specific MCP client by name
func (m *MCPManager) GetClientByName(name string) (*MCPClient, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[name]
	return client, exists
}

// registerTools registers all tools from a client
func (m *MCPManager) registerTools(ctx context.Context, mcpClient *MCPClient) {
	// Get tools from the client
	tools, err := mcpClient.GetTools(ctx)
	if err != nil {
		log.Printf("Failed to get tools from MCP client %s: %v", mcpClient.name, err)
		return
	}

	// Register each tool
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, tool := range tools {
		info, err := tool.Info(ctx)
		if err != nil {
			log.Printf("Failed to get info for tool: %v", err)
			continue
		}

		// Add a prefix to avoid name collisions
		toolKey := fmt.Sprintf("%s.%s", mcpClient.name, info.Name)
		m.tools[toolKey] = tool

		log.Printf("Registered tool: %s (%s)", toolKey, info.Desc)
	}

	// Store tools in the client
	mcpClient.tools = tools
}

// connect establishes a connection to the MCP server
func (c *MCPClient) connect(ctx context.Context) error {
	// Create a new MCP client
	cli, err := client.NewSSEMCPClient(c.url)
	if err != nil {
		return fmt.Errorf("failed to create MCP client: %w", err)
	}

	// Start the client
	err = cli.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start MCP client: %w", err)
	}

	// Initialize the MCP client
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    fmt.Sprintf("eino-coder-%s", c.name),
		Version: "1.0.0",
	}

	_, err = cli.Initialize(ctx, initRequest)
	if err != nil {
		return fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	c.client = cli
	return nil
}

// GetTools retrieves tools from the MCP server
func (c *MCPClient) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	if c.client == nil {
		return nil, fmt.Errorf("MCP client not connected")
	}

	tools, err := mcpp.GetTools(ctx, &mcpp.Config{Cli: c.client})
	if err != nil {
		return nil, fmt.Errorf("failed to get tools from MCP server: %w", err)
	}

	return tools, nil
}

// ExecuteTool executes a tool by name with the provided arguments
func (c *MCPClient) ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (string, error) {
	// Find the tool
	var selectedTool tool.BaseTool
	for _, t := range c.tools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		if info.Name == toolName {
			selectedTool = t
			break
		}
	}

	if selectedTool == nil {
		return "", fmt.Errorf("tool %s not found", toolName)
	}

	// Execute the tool
	invokableTool, ok := selectedTool.(tool.InvokableTool)
	if !ok {
		return "", fmt.Errorf("tool %s is not invokable", toolName)
	}

	// Convert arguments to JSON
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("failed to convert arguments to JSON: %w", err)
	}

	// Execute the tool
	result, err := invokableTool.InvokableRun(ctx, string(argsJSON))
	if err != nil {
		return "", fmt.Errorf("failed to execute tool %s: %w", toolName, err)
	}

	return result, nil
}

// Close closes the MCP client connection
func (c *MCPClient) Close() error {
	if c.client != nil {
		c.client.Close()
		c.client = nil
	}
	return nil
}

// Close closes all MCP client connections
func (m *MCPManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, client := range m.clients {
		if err := client.Close(); err != nil {
			log.Printf("Error closing MCP client %s: %v", name, err)
		}
	}

	m.clients = make(map[string]*MCPClient)
	m.tools = make(map[string]tool.BaseTool)
}

// HealthCheck checks if all clients are connected
func (m *MCPManager) HealthCheck() map[string]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	health := make(map[string]bool)
	for name, client := range m.clients {
		// A simple check - if we have tools, consider the client healthy
		health[name] = len(client.tools) > 0
	}

	return health
}

// ReconnectAll attempts to reconnect all disconnected clients
func (m *MCPManager) ReconnectAll(ctx context.Context) {
	for _, clientCfg := range app.Config.MCP.Clients {
		if !clientCfg.Enabled {
			continue
		}

		m.mu.RLock()
		client, exists := m.clients[clientCfg.Name]
		m.mu.RUnlock()

		if !exists {
			// Create new client
			mcpClient := &MCPClient{
				name:        clientCfg.Name,
				description: clientCfg.Description,
				url:         clientCfg.URL,
			}

			if err := mcpClient.connect(ctx); err != nil {
				log.Printf("Failed to connect MCP client %s: %v", clientCfg.Name, err)
				continue
			}

			m.mu.Lock()
			m.clients[clientCfg.Name] = mcpClient
			m.mu.Unlock()

			m.registerTools(ctx, mcpClient)
		} else {
			// Check if client needs reconnection
			needsReconnect := false

			m.mu.RLock()
			if len(client.tools) == 0 {
				needsReconnect = true
			}
			m.mu.RUnlock()

			if needsReconnect {
				log.Printf("Reconnecting MCP client: %s", clientCfg.Name)
				client.Close()

				if err := client.connect(ctx); err != nil {
					log.Printf("Failed to reconnect MCP client %s: %v", clientCfg.Name, err)
					continue
				}

				m.registerTools(ctx, client)
			}
		}
	}
}

// StartHealthChecker starts a goroutine to periodically check and reconnect clients
func (m *MCPManager) StartHealthChecker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.ReconnectAll(ctx)
			}
		}
	}()
}

// GetAllClients returns all MCP clients
func (m *MCPManager) GetAllClients() []*MCPClient {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients := make([]*MCPClient, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	return clients
}

// GetName returns the name of the client
func (c *MCPClient) GetName() string {
	return c.name
}

// GetClientTools returns the tools for a specific client
func (m *MCPManager) GetClientTools(client *MCPClient) []tool.BaseTool {
	if client == nil {
		return nil
	}

	// Return a copy of the tools slice to avoid direct modifications
	toolsCopy := make([]tool.BaseTool, len(client.tools))
	copy(toolsCopy, client.tools)
	return toolsCopy
}
