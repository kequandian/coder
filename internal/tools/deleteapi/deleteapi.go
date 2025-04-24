package deleteapi

import (
	"coder/internal/cache"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// DeleteAPITool is a tool for deleting an API URL from a module
type DeleteAPITool struct{}

// NewDeleteAPITool creates a new delete API tool
func NewDeleteAPITool() (*DeleteAPITool, error) {
	return &DeleteAPITool{}, nil
}

// Info returns information about the tool
func (t *DeleteAPITool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "deleteAPI",
		Desc: "Delete an API URL from a module",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"type": {
				Desc:     "The type of API to delete (one of: listAPI, createAPI, getAPI, updateAPI, deleteAPI)",
				Type:     schema.String,
				Required: true,
			},
		}),
	}, nil
}

// IsInvokable indicates that this tool can be invoked
func (t *DeleteAPITool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *DeleteAPITool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
	// Parse the arguments
	var params struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate required parameters
	if params.Type == "" {
		return "", fmt.Errorf("type cannot be empty")
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

	// Delete the API URL
	delete(mapCur, params.Type)

	// Save to cache
	jsonCur, _ := json.Marshal(mapCur)
	moduleCache := cache.NewModuleCacheData(infoCache.ModuleName, infoCache.ModuleCode, infoCache.Support, string(jsonCur))
	cache.ModuleCacheInstance.Set(cache.CacheKey(userReq.ConversationID), moduleCache, cache.DefaultCacheExpiration)

	mockResponse := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("API '%s' has been deleted successfully", params.Type),
		"params":  params,
	}
	result, err := json.MarshalIndent(mockResponse, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}
