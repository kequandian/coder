package addapi

import (
	"coder/internal/cache"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// AddAPITool is a tool for adding an API URL to a module
type AddAPITool struct{}

// NewAddAPITool creates a new add API tool
func NewAddAPITool() (*AddAPITool, error) {
	return &AddAPITool{}, nil
}

// Info returns information about the tool
func (t *AddAPITool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "addAPI",
		Desc: "Add an API URL to a module",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"type": {
				Desc:     "The type of API (one of: listAPI, createAPI, getAPI, updateAPI, deleteAPI)",
				Type:     schema.String,
				Required: true,
			},
			"url": {
				Desc:     "The API URL to add (e.g., '/api/crud/fieldModel/fieldModels')",
				Type:     schema.String,
				Required: true,
			},
		}),
	}, nil
}

// IsInvokable indicates that this tool can be invoked
func (t *AddAPITool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *AddAPITool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
	// Parse the arguments
	var params struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate required parameters
	if params.Type == "" {
		return "", fmt.Errorf("type cannot be empty")
	}
	if params.URL == "" {
		return "", fmt.Errorf("url cannot be empty")
	}

	// Validate type is one of the allowed values
	validTypes := map[string]bool{
		"listAPI":   true,
		"createAPI": true,
		"getAPI":    true,
		"updateAPI": true,
		"deleteAPI": true,
	}
	if !validTypes[params.Type] {
		return "", fmt.Errorf("invalid type: %s. Must be one of: listAPI, createAPI, getAPI, updateAPI, deleteAPI", params.Type)
	}

	// Decode module from context
	infoCache, mapCur, _, userReq, err := cache.DecodeModuleFromCtx(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to decode module from context: %w", err)
	}

	// Add or update the API URL
	mapCur[params.Type] = params.URL

	// Save to cache
	jsonCur, _ := json.Marshal(mapCur)
	moduleCache := cache.NewModuleCacheData(infoCache.ModuleName, infoCache.ModuleCode, infoCache.Support, string(jsonCur))
	cache.ModuleCacheInstance.Set(cache.CacheKey(userReq.ConversationID), moduleCache, cache.DefaultCacheExpiration)

	mockResponse := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("API URL for range '%s' (%s) has been added successfully with %d APIs",
			params.Type, params.URL, len(params.URL)),
		"params": params,
	}
	result, err := json.MarshalIndent(mockResponse, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}
