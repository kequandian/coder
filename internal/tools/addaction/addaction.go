package addaction

import (
	"coder/internal/cache"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// AddActionTool is a tool for adding an action to a module
type AddActionTool struct{}

// NewAddActionTool creates a new add action tool
func NewAddActionTool() (*AddActionTool, error) {
	return &AddActionTool{}, nil
}

// Info returns information about the tool
func (t *AddActionTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "addAction",
		Desc: "Add an action to a module",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"title": {
				Desc:     "The title of the action (e.g., '添加')",
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
func (t *AddActionTool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *AddActionTool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
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

	// In a real implementation, we would add the action to the module
	infoCache, mapCur, _, userReq, err := cache.DecodeModuleFromCtx(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to decode module from context: %w", err)
	}

	tableAction, ok := mapCur["tableActions"].([]interface{})
	if !ok {
		// Initialize tableAction if it doesn't exist
		tableAction = make([]interface{}, 0)
	}

	title2Index := make(map[string]int)
	for index, action := range tableAction {
		actionMap := action.(map[string]interface{})
		if title, ok := actionMap["title"].(string); ok {
			title2Index[title] = index
		}
	}

	if index, ok := title2Index[params.Title]; ok {
		tableAction[index] = map[string]interface{}{
			"title":   params.Title,
			"type":    params.Type,
			"options": params.Options,
		}
	} else {
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
		"message": fmt.Sprintf("Action for range '%s' (%s) has been added successfully with %d actions",
			params.Title, params.Type, len(params.Options)),
		"params": params,
	}
	result, err := json.MarshalIndent(mockResponse, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}
