package editapi

import (
	"coder/internal/cache"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// EditAPITool is a tool for editing an API URL in a module
type EditAPITool struct{}

// NewEditAPITool creates a new edit API tool
func NewEditAPITool() (*EditAPITool, error) {
	return &EditAPITool{}, nil
}

// Info returns information about the tool
func (t *EditAPITool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "editAPI",
		Desc: "Edit an API URL in a module",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"type": {
				Desc:     "The type of API (one of: listAPI, createAPI, getAPI, updateAPI, deleteAPI)",
				Type:     schema.String,
				Required: true,
			},
			"url": {
				Desc:     "The new API URL (e.g., '/api/crud/fieldModel/fieldModels')",
				Type:     schema.String,
				Required: true,
			},
		}),
	}, nil
}

// IsInvokable indicates that this tool can be invoked
func (t *EditAPITool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *EditAPITool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
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

	// Update the API URL
	mapCur[params.Type] = params.URL

	// Save to cache
	jsonCur, _ := json.Marshal(mapCur)
	moduleCache := cache.NewModuleCacheData(infoCache.ModuleName, infoCache.ModuleCode, infoCache.Support, string(jsonCur))
	cache.ModuleCacheInstance.Set(cache.CacheKey(userReq.ConversationID), moduleCache, cache.DefaultCacheExpiration)

	mockResponse := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("API '%s' has been edited successfully", params.Type),
		"params":  params,
	}
	result, err := json.MarshalIndent(mockResponse, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}
