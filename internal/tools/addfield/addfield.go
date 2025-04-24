package addfield

import (
	"coder/internal/cache"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// AddFieldTool is a tool for adding a field to a module
type AddFieldTool struct{}

// NewAddFieldTool creates a new add field tool
func NewAddFieldTool() (*AddFieldTool, error) {
	return &AddFieldTool{}, nil
}

// Info returns information about the tool
func (t *AddFieldTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "addField",
		Desc: "Add a field to a module",
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
				Desc:     "The fields in the range",
				Type:     schema.Array,
				Required: true,
			},
		}),
	}, nil
}

// IsInvokable indicates that this tool can be invoked
func (t *AddFieldTool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *AddFieldTool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
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

	// In a real implementation, here we would delete the field from the module
	// For this mock implementation, we return a success message
	infoCache, mapCur, mapSupport, userReq, err := cache.DecodeModuleFromCtx(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to decode module from context: %w", err)
	}

	if params.RangeCode == "view" {
		// 获取支持的field
		supportFields := mapSupport["viewConfig"].([]interface{})
		supportFieldsMap := make(map[string]bool)
		for _, field := range supportFields {
			supportFieldsMap[field.(map[string]interface{})["field"].(string)] = true
		}
		// 获取存在的field
		existFields := mapCur["viewConfig"].([]interface{})
		existFieldsMap := make(map[string]bool)
		for _, field := range existFields {
			existFieldsMap[field.(map[string]interface{})["field"].(string)] = true
		}

		viewConfig := mapCur["viewConfig"].([]interface{})
		for _, field := range params.Fields {
			if _, ok := supportFieldsMap[field["field"].(string)]; ok {
				if _, ok := existFieldsMap[field["field"].(string)]; !ok {
					viewConfig = append(viewConfig, field)
				}
			}
		}
		mapCur["viewConfig"] = viewConfig
	} else if params.RangeCode == "create" {
		// 获取支持的field
		supportFields := mapSupport["createFields"].([]interface{})
		supportFieldsMap := make(map[string]bool)
		for _, field := range supportFields {
			supportFieldsMap[field.(map[string]interface{})["field"].(string)] = true
		}
		// 获取存在的field
		existFields := mapCur["createFields"].([]interface{})
		existFieldsMap := make(map[string]bool)
		for _, field := range existFields {
			existFieldsMap[field.(map[string]interface{})["field"].(string)] = true
		}

		createFields := mapCur["createFields"].([]interface{})
		for _, field := range params.Fields {
			if _, ok := supportFieldsMap[field["field"].(string)]; ok {
				if _, ok := existFieldsMap[field["field"].(string)]; !ok {
					createFields = append(createFields, field)
				}
			}
		}
		mapCur["createFields"] = createFields
	} else if params.RangeCode == "update" {
		// 获取支持的field
		supportFields := mapSupport["updateFields"].([]interface{})
		supportFieldsMap := make(map[string]bool)
		for _, field := range supportFields {
			supportFieldsMap[field.(map[string]interface{})["field"].(string)] = true
		}
		// 获取存在的field
		existFields := mapCur["updateFields"].([]interface{})
		existFieldsMap := make(map[string]bool)
		for _, field := range existFields {
			existFieldsMap[field.(map[string]interface{})["field"].(string)] = true
		}

		updateFields := mapCur["updateFields"].([]interface{})
		for _, field := range params.Fields {
			if _, ok := supportFieldsMap[field["field"].(string)]; ok {
				if _, ok := existFieldsMap[field["field"].(string)]; !ok {
					updateFields = append(updateFields, field)
				}
			}
		}
		mapCur["updateFields"] = updateFields
	} else if params.RangeCode == "list" {
		// 获取支持的field
		supportFields := mapSupport["tableFields"].([]interface{})
		supportFieldsMap := make(map[string]bool)
		for _, field := range supportFields {
			supportFieldsMap[field.(map[string]interface{})["field"].(string)] = true
		}
		// 获取存在的field
		existFields := mapCur["tableFields"].([]interface{})
		existFieldsMap := make(map[string]bool)
		for _, field := range existFields {
			existFieldsMap[field.(map[string]interface{})["field"].(string)] = true
		}

		tableFields := mapCur["tableFields"].([]interface{})
		for _, field := range params.Fields {
			if _, ok := supportFieldsMap[field["field"].(string)]; ok {
				if _, ok := existFieldsMap[field["field"].(string)]; !ok {
					tableFields = append(tableFields, field)
				}
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
		"message": fmt.Sprintf("Field for range '%s' (%s) has been added successfully with %d fields",
			params.RangeName, params.RangeCode, len(params.Fields)),
		"params": params,
	}

	result, err := json.MarshalIndent(mockResponse, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}
