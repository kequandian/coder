package deletesearch

import (
	"coder/internal/cache"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// DeleteSearchTool is a tool for deleting a search condition from a module
type DeleteSearchTool struct{}

// NewDeleteSearchTool creates a new delete search tool
func NewDeleteSearchTool() (*DeleteSearchTool, error) {
	return &DeleteSearchTool{}, nil
}

// Info returns information about the tool
func (t *DeleteSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "deleteSearch",
		Desc: "Delete a search condition from a module",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"fields": {
				Desc:     "The fields identifier for the search (e.g., 'templateName')",
				Type:     schema.Array,
				Required: true,
			},
		}),
	}, nil
}

// IsInvokable indicates that this tool can be invoked
func (t *DeleteSearchTool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *DeleteSearchTool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
	// Parse the arguments
	var params struct {
		Fields []string `json:"fields"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate required parameters
	for _, field := range params.Fields {
		if field == "" {
			return "", fmt.Errorf("field cannot be empty")
		}
	}

	infoCache, mapCur, _, userReq, err := cache.DecodeModuleFromCtx(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to decode module from context: %w", err)
	}
	fieldsMap := make(map[string]bool)
	for _, field := range params.Fields {
		fieldsMap[field] = true
	}

	searchFields := mapCur["searchFields"].([]interface{})
	for index, field := range searchFields {
		if _, ok := fieldsMap[field.(map[string]interface{})["field"].(string)]; ok {
			searchFields = append(searchFields[:index], searchFields[index+1:]...)
		}
	}
	mapCur["searchFields"] = searchFields

	jsonCur, _ := json.Marshal(mapCur)
	moduleCache := cache.NewModuleCacheData(infoCache.ModuleName, infoCache.ModuleCode, infoCache.Support, string(jsonCur))
	cache.ModuleCacheInstance.Set(cache.CacheKey(userReq.ConversationID), moduleCache, cache.DefaultCacheExpiration)

	jsonCur, _ = json.Marshal(params.Fields)
	mockResponse := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Search field '%s' has been deleted successfully", string(jsonCur)),
		"params":  params,
	}
	result, err := json.MarshalIndent(mockResponse, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}
