package deleteoperation

import (
	"coder/api"
	"coder/internal/cache"
	"coder/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// DeleteOperationTool is a tool for deleting an operation from a module
type DeleteOperationTool struct{}

// NewDeleteOperationTool creates a new delete operation tool
func NewDeleteOperationTool() (*DeleteOperationTool, error) {
	return &DeleteOperationTool{}, nil
}

// Info returns information about the tool
func (t *DeleteOperationTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "deleteOperation",
		Desc: "Delete an operation from a module",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"title": {
				Desc:     "The title of the operation to delete",
				Type:     schema.String,
				Required: true,
			},
		}),
	}, nil
}

// IsInvokable indicates that this tool can be invoked
func (t *DeleteOperationTool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *DeleteOperationTool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
	// Parse the arguments
	var params struct {
		Title string `json:"title"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate required parameters
	if params.Title == "" {
		return "", fmt.Errorf("title cannot be empty")
	}

	// In a real implementation, we would delete the operation from the module
	userReq, ok := ctx.Value(config.StateKey).(*api.ChatRequest)
	if !ok {
		return "", fmt.Errorf("state not found in context")
	}
	log.Printf("Processing LocalTool calls in message: %v", userReq)

	infoCache, mapCur, _, userReq, err := cache.DecodeModuleFromCtx(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to decode module from context: %w", err)
	}

	tableOperation := mapCur["tableOperation"].([]interface{})
	for index, field := range tableOperation {
		if field.(map[string]interface{})["title"] == params.Title {
			tableOperation = append(tableOperation[:index], tableOperation[index+1:]...)
			break
		}
	}
	mapCur["tableOperation"] = tableOperation

	// Save to cache
	jsonCur, _ := json.Marshal(mapCur)
	moduleCache := cache.NewModuleCacheData(infoCache.ModuleName, infoCache.ModuleCode, infoCache.Support, string(jsonCur))
	cache.ModuleCacheInstance.Set(cache.CacheKey(userReq.ConversationID), moduleCache, cache.DefaultCacheExpiration)

	mockResponse := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Operation '%s' has been deleted successfully", params.Title),
		"params":  params,
	}
	result, err := json.MarshalIndent(mockResponse, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}
