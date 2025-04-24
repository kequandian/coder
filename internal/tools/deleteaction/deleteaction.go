package deleteaction

import (
	"coder/internal/cache"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// DeleteActionTool is a tool for deleting an action from a module
type DeleteActionTool struct{}

// NewDeleteActionTool creates a new delete action tool
func NewDeleteActionTool() (*DeleteActionTool, error) {
	return &DeleteActionTool{}, nil
}

// Info returns information about the tool
func (t *DeleteActionTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "deleteAction",
		Desc: "Delete an action from a module",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"title": {
				Desc:     "The title of the action to delete",
				Type:     schema.String,
				Required: true,
			},
		}),
	}, nil
}

// IsInvokable indicates that this tool can be invoked
func (t *DeleteActionTool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *DeleteActionTool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
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

	// In a real implementation, we would delete the action from the module
	infoCache, mapCur, _, userReq, err := cache.DecodeModuleFromCtx(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to decode module from context: %w", err)
	}

	tableAction, ok := mapCur["tableActions"].([]interface{})
	if !ok {
		// If tableAction doesn't exist, there's nothing to delete
		return fmt.Sprintf("No actions found with title: %s", params.Title), nil
	}

	for index, action := range tableAction {
		actionMap := action.(map[string]interface{})
		if title, ok := actionMap["title"].(string); ok && title == params.Title {
			tableAction = append(tableAction[:index], tableAction[index+1:]...)
			break
		}
	}

	mapCur["tableActions"] = tableAction

	// Save to cache
	jsonCur, _ := json.Marshal(mapCur)
	moduleCache := cache.NewModuleCacheData(infoCache.ModuleName, infoCache.ModuleCode, infoCache.Support, string(jsonCur))
	cache.ModuleCacheInstance.Set(cache.CacheKey(userReq.ConversationID), moduleCache, cache.DefaultCacheExpiration)

	mockResponse := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Action '%s' has been deleted successfully", params.Title),
		"params":  params,
	}
	result, err := json.MarshalIndent(mockResponse, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}
