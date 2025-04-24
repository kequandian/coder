package editoperation

import (
	"coder/internal/cache"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// EditOperationTool is a tool for editing an operation in a module
type EditOperationTool struct{}

// NewEditOperationTool creates a new edit operation tool
func NewEditOperationTool() (*EditOperationTool, error) {
	return &EditOperationTool{}, nil
}

// Info returns information about the tool
func (t *EditOperationTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "editOperation",
		Desc: "Edit an operation in a module",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"old_title": {
				Desc:     "The current title of the operation to edit",
				Type:     schema.String,
				Required: true,
			},
			"title": {
				Desc:     "The new title of the operation (e.g., '详情')",
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
func (t *EditOperationTool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *EditOperationTool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
	// Parse the arguments
	var params struct {
		OldTitle string                 `json:"old_title"`
		Title    string                 `json:"title"`
		Type     string                 `json:"type"`
		Options  map[string]interface{} `json:"options"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate required parameters
	if params.OldTitle == "" {
		return "", fmt.Errorf("old_title cannot be empty")
	}
	if params.Title == "" {
		return "", fmt.Errorf("title cannot be empty")
	}
	if params.Type == "" {
		return "", fmt.Errorf("type cannot be empty")
	}
	if params.Options == nil {
		return "", fmt.Errorf("options cannot be empty")
	}

	infoCache, mapCur, _, userReq, err := cache.DecodeModuleFromCtx(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to decode module from context: %w", err)
	}

	tableOperation := mapCur["tableOperation"].([]interface{})
	title2Index := make(map[string]int)
	for index, field := range tableOperation {
		title2Index[field.(map[string]interface{})["title"].(string)] = index
	}
	if index, ok := title2Index[params.OldTitle]; ok {
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
		"message": fmt.Sprintf("Operation '%s' has been edited successfully", params.Title),
		"params":  params,
	}
	result, err := json.MarshalIndent(mockResponse, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}
