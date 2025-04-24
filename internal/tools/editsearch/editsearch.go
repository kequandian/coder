package editsearch

import (
	"coder/internal/cache"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// EditSearchTool is a tool for editing a search condition in a module
type EditSearchTool struct{}

// NewEditSearchTool creates a new edit search tool
func NewEditSearchTool() (*EditSearchTool, error) {
	return &EditSearchTool{}, nil
}

// Info returns information about the tool
func (t *EditSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "editSearch",
		Desc: "Edit a search condition in a module",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"label": {
				Desc:     "The display label of the search field (e.g., '模板名称')",
				Type:     schema.String,
				Required: true,
			},
			"field": {
				Desc:     "The field identifier for the search (e.g., 'templateName')",
				Type:     schema.String,
				Required: true,
			},
			"type": {
				Desc:     "The type of the search field (e.g., 'search')",
				Type:     schema.String,
				Required: true,
			},
			"props": {
				Desc:     "The properties for the search field (e.g., placeholder text)",
				Type:     schema.Object,
				Required: true,
			},
		}),
	}, nil
}

// IsInvokable indicates that this tool can be invoked
func (t *EditSearchTool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *EditSearchTool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
	// Parse the arguments
	var params struct {
		Label string                 `json:"label"`
		Field string                 `json:"field"`
		Type  string                 `json:"type"`
		Props map[string]interface{} `json:"props"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate required parameters
	if params.Label == "" {
		return "", fmt.Errorf("label cannot be empty")
	}
	if params.Field == "" {
		return "", fmt.Errorf("field cannot be empty")
	}
	if params.Type == "" {
		return "", fmt.Errorf("type cannot be empty")
	}
	if params.Props == nil {
		return "", fmt.Errorf("props cannot be empty")
	}

	infoCache, mapCur, mapSupport, userReq, err := cache.DecodeModuleFromCtx(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to decode module from context: %w", err)
	}
	supportFields := mapSupport["searchFields"].([]interface{})
	supportFieldsMap := make(map[string]bool)
	for _, field := range supportFields {
		supportFieldsMap[field.(map[string]interface{})["field"].(string)] = true
	}

	if _, ok := supportFieldsMap[params.Field]; !ok {
		return "", fmt.Errorf("field not supported")
	}

	searchFields := mapCur["searchFields"].([]interface{})
	searchFieldsMap := make(map[string]int)
	for index, field := range searchFields {
		searchFieldsMap[field.(map[string]interface{})["field"].(string)] = index
	}

	if index, ok := searchFieldsMap[params.Field]; ok {
		searchFields[index] = map[string]interface{}{
			"label": params.Label,
			"field": params.Field,
			"type":  params.Type,
			"props": params.Props,
		}
	} else {
		searchFields = append(searchFields, map[string]interface{}{
			"label": params.Label,
			"field": params.Field,
			"type":  params.Type,
			"props": params.Props,
		})
	}
	mapCur["searchFields"] = searchFields

	jsonCur, _ := json.Marshal(mapCur)
	moduleCache := cache.NewModuleCacheData(infoCache.ModuleName, infoCache.ModuleCode, infoCache.Support, string(jsonCur))
	cache.ModuleCacheInstance.Set(cache.CacheKey(userReq.ConversationID), moduleCache, cache.DefaultCacheExpiration)

	mockResponse := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Search field '%s' has been edited successfully", params.Label),
		"params":  params,
	}
	result, err := json.MarshalIndent(mockResponse, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}
