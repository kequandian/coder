package addoperation

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

// AddOperationTool is a tool for adding an operation to a module
type AddOperationTool struct{}

// NewAddOperationTool creates a new add operation tool
func NewAddOperationTool() (*AddOperationTool, error) {
	return &AddOperationTool{}, nil
}

// Info returns information about the tool
func (t *AddOperationTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "addOperation",
		Desc: "Add an operation to a module",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"title": {
				Desc:     "The title of the operation (e.g., '详情')",
				Type:     schema.String,
				Required: true,
			},
			"type": {
				Desc:     "The type of the operation (e.g., 'path')",
				Type:     schema.String,
				Required: true,
			},
			"options": {
				Desc:     "The options for the operation (e.g., outside, path)",
				Type:     schema.Object,
				Required: true,
			},
		}),
	}, nil
}

// IsInvokable indicates that this tool can be invoked
func (t *AddOperationTool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *AddOperationTool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
	// Parse the arguments
	var params struct {
		Title   string                 `json:"title"`
		Type    string                 `json:"type"`
		Options map[string]interface{} `json:"options"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate required parameters
	if params.Title == "" {
		return "", fmt.Errorf("title cannot be empty")
	}
	if params.Type == "" {
		return "", fmt.Errorf("type cannot be empty")
	}
	if params.Options == nil {
		return "", fmt.Errorf("options cannot be empty")
	}

	// In a real implementation, we would add the operation to the module
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
	title2Index := make(map[string]int)
	for index, field := range tableOperation {
		title2Index[field.(map[string]interface{})["title"].(string)] = index
	}
	if index, ok := title2Index[params.Title]; ok {
		tableOperation[index] = map[string]interface{}{
			"title":   params.Title,
			"type":    params.Type,
			"options": params.Options,
		}
	} else {
		tableOperation = append(tableOperation, map[string]interface{}{
			"title":   params.Title,
			"type":    params.Type,
			"options": params.Options,
		})
	}
	mapCur["tableOperation"] = tableOperation
	// Save to cache
	jsonCur, _ := json.Marshal(mapCur)
	moduleCache := cache.NewModuleCacheData(infoCache.ModuleName, infoCache.ModuleCode, infoCache.Support, string(jsonCur))
	cache.ModuleCacheInstance.Set(cache.CacheKey(userReq.ConversationID), moduleCache, cache.DefaultCacheExpiration)

	mockResponse := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Operation '%s' has been added successfully", params.Title),
		"params":  params,
	}
	result, err := json.MarshalIndent(mockResponse, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}
