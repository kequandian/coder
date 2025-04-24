package editaction

import (
	"coder/internal/cache"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// EditActionTool is a tool for editing an action in a module
type EditActionTool struct{}

// NewEditActionTool creates a new edit action tool
func NewEditActionTool() (*EditActionTool, error) {
	return &EditActionTool{}, nil
}

// Info returns information about the tool
func (t *EditActionTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "editAction",
		Desc: "Edit an action in a module",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"old_title": {
				Desc:     "The current title of the action to edit",
				Type:     schema.String,
				Required: true,
			},
			"title": {
				Desc:     "The new title of the action (e.g., '添加')",
				Type:     schema.String,
				Required: true,
			},
			"type": {
				Desc:     "The type of the action (e.g., 'path')",
				Type:     schema.String,
				Required: true,
			},
			"options": {
				Desc:     "The options for the action (e.g., style, path)",
				Type:     schema.Object,
				Required: true,
			},
		}),
	}, nil
}

// IsInvokable indicates that this tool can be invoked
func (t *EditActionTool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *EditActionTool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
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

	// In a real implementation, we would edit the action in the module
	infoCache, mapCur, _, userReq, err := cache.DecodeModuleFromCtx(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to decode module from context: %w", err)
	}

	tableAction, ok := mapCur["tableActions"].([]interface{})
	if !ok {
		// Initialize tableAction if it doesn't exist
		tableAction = make([]interface{}, 0)
	}

	foundAndUpdated := false
	for index, action := range tableAction {
		actionMap := action.(map[string]interface{})
		if title, ok := actionMap["title"].(string); ok && title == params.OldTitle {
			tableAction[index] = map[string]interface{}{
				"title":   params.Title,
				"type":    params.Type,
				"options": params.Options,
			}
			foundAndUpdated = true
			break
		}
	}

	if !foundAndUpdated {
		// If we didn't find and update an existing action, add a new one
		tableAction = append(tableAction, map[string]interface{}{
			"title":   params.Title,
			"type":    params.Type,
			"options": params.Options,
		})
	}

	mapCur["tableActions"] = tableAction

	// Save to cache
	jsonCur, _ := json.Marshal(mapCur)
	moduleCache := cache.NewModuleCacheData(infoCache.ModuleName, infoCache.ModuleCode, infoCache.Support, string(jsonCur))
	cache.ModuleCacheInstance.Set(cache.CacheKey(userReq.ConversationID), moduleCache, cache.DefaultCacheExpiration)

	mockResponse := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Action '%s' has been edited successfully", params.Title),
		"params":  params,
	}
	result, err := json.MarshalIndent(mockResponse, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}
