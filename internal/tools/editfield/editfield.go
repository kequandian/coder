package editfield

import (
	"context"
	"encoding/json"
	"fmt"

	"coder/internal/cache"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// EditFieldTool is a tool for editing a field in a module
type EditFieldTool struct{}

// NewEditFieldTool creates a new edit field tool
func NewEditFieldTool() (*EditFieldTool, error) {
	return &EditFieldTool{}, nil
}

// Info returns information about the tool
func (t *EditFieldTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "editField",
		Desc: "Edit a field in a module",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"range_name": {
				Desc:     "The name of the range (e.g., 详情，创建/新建/添加, 更新/修改, 列表/分页/表格)",
				Type:     schema.String,
				Required: true,
			},
			"range_code": {
				Desc:     "The code of the range (e.g., view, create, update, list)",
				Type:     schema.String,
				Required: true,
			},
			"fields": {
				Desc:     "The updated fields in the range",
				Type:     schema.Array,
				Required: true,
			},
		}),
	}, nil
}

// IsInvokable indicates that this tool can be invoked
func (t *EditFieldTool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *EditFieldTool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
	// Parse the arguments
	var params struct {
		RangeName string                   `json:"range_name"`
		RangeCode string                   `json:"range_code"`
		Fields    []map[string]interface{} `json:"fields"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	if params.RangeName == "" || params.RangeCode == "" {
		return "", fmt.Errorf("range_name and range_code cannot be empty")
	}

	if len(params.Fields) == 0 {
		return "", fmt.Errorf("fields cannot be empty")
	}

	infoCache, mapCur, _, userReq, err := cache.DecodeModuleFromCtx(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to decode module from context: %w", err)
	}

	if params.RangeCode == "view" {
		// 获取存在的field
		existFields := mapCur["viewConfig"].([]interface{})
		existFieldsMap := make(map[string]int)
		for index, field := range existFields {
			existFieldsMap[field.(map[string]interface{})["field"].(string)] = index
		}

		viewConfig := mapCur["viewConfig"].([]interface{})
		for _, field := range params.Fields {
			if index, ok := existFieldsMap[field["field"].(string)]; ok {
				viewConfig[index] = field
			}
		}
		mapCur["viewConfig"] = viewConfig
	} else if params.RangeCode == "create" {
		// 获取存在的field
		existFields := mapCur["createFields"].([]interface{})
		existFieldsMap := make(map[string]int)
		for index, field := range existFields {
			existFieldsMap[field.(map[string]interface{})["field"].(string)] = index
		}

		createFields := mapCur["createFields"].([]interface{})
		for _, field := range params.Fields {
			if index, ok := existFieldsMap[field["field"].(string)]; ok {
				createFields[index] = field
			}
		}
		mapCur["createFields"] = createFields
	} else if params.RangeCode == "update" {
		// 获取存在的field
		existFields := mapCur["updateFields"].([]interface{})
		existFieldsMap := make(map[string]int)
		for index, field := range existFields {
			existFieldsMap[field.(map[string]interface{})["field"].(string)] = index
		}

		updateFields := mapCur["updateFields"].([]interface{})
		for _, field := range params.Fields {
			if index, ok := existFieldsMap[field["field"].(string)]; ok {
				updateFields[index] = field
			}
		}
		mapCur["updateFields"] = updateFields
	} else if params.RangeCode == "list" {
		// 获取存在的field
		existFields := mapCur["tableFields"].([]interface{})
		existFieldsMap := make(map[string]int)
		for index, field := range existFields {
			existFieldsMap[field.(map[string]interface{})["field"].(string)] = index
		}

		tableFields := mapCur["tableFields"].([]interface{})
		for _, field := range params.Fields {
			if index, ok := existFieldsMap[field["field"].(string)]; ok {
				tableFields[index] = field
			}
		}
		mapCur["tableFields"] = tableFields
	}

	jsonCur, _ := json.Marshal(mapCur)
	// Store in cache
	moduleCache := cache.NewModuleCacheData(infoCache.ModuleName, infoCache.ModuleCode, infoCache.Support, string(jsonCur))
	cache.ModuleCacheInstance.Set(cache.CacheKey(userReq.ConversationID), moduleCache, cache.DefaultCacheExpiration)

	mockResponse := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Field for range '%s' (%s) has been updated successfully with %d fields",
			params.RangeName, params.RangeCode, len(params.Fields)),
		"params": params,
	}

	result, err := json.MarshalIndent(mockResponse, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}
