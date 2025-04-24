package deletefield

import (
	"coder/internal/cache"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// DeleteFieldTool is a tool for deleting a field from a module
type DeleteFieldTool struct{}

// NewDeleteFieldTool creates a new delete field tool
func NewDeleteFieldTool() (*DeleteFieldTool, error) {
	return &DeleteFieldTool{}, nil
}

// Info returns information about the tool
func (t *DeleteFieldTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "deleteField",
		Desc: "Delete a field from a module",
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
func (t *DeleteFieldTool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *DeleteFieldTool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
	// Parse the arguments
	var params struct {
		RangeName string   `json:"range_name"`
		RangeCode string   `json:"range_code"`
		Fields    []string `json:"fields"`
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

	fieldsMap := make(map[string]bool)
	for _, field := range params.Fields {
		fieldsMap[field] = true
	}

	// tip := fmt.Sprintf("删除字段: %v", params.Fields)
	// 删除字段
	if params.RangeCode == "view" {
		viewConfig := mapCur["viewConfig"].([]interface{})
		for index, field := range viewConfig {
			if _, ok := fieldsMap[field.(map[string]interface{})["field"].(string)]; ok {
				viewConfig = append(viewConfig[:index], viewConfig[index+1:]...)
			}
		}
		mapCur["viewConfig"] = viewConfig
	} else if params.RangeCode == "create" {
		createFields := mapCur["createFields"].([]interface{})
		for index, field := range createFields {
			if _, ok := fieldsMap[field.(map[string]interface{})["field"].(string)]; ok {
				createFields = append(createFields[:index], createFields[index+1:]...)
			}
		}
		mapCur["createFields"] = createFields
	} else if params.RangeCode == "update" {
		updateFields := mapCur["updateFields"].([]interface{})
		for index, field := range updateFields {
			if _, ok := fieldsMap[field.(map[string]interface{})["field"].(string)]; ok {
				updateFields = append(updateFields[:index], updateFields[index+1:]...)
			}
		}
		mapCur["updateFields"] = updateFields
	} else if params.RangeCode == "list" {
		tableFields := mapCur["tableFields"].([]interface{})
		for index, tableField := range tableFields {
			if _, ok := fieldsMap[tableField.(map[string]interface{})["field"].(string)]; ok {
				tableFields = append(tableFields[:index], tableFields[index+1:]...)
			}
		}
		mapCur["tableFields"] = tableFields
	}

	jsonCur, _ := json.Marshal(mapCur)
	moduleCache := cache.NewModuleCacheData(infoCache.ModuleName, infoCache.ModuleCode, infoCache.Support, string(jsonCur))
	cache.ModuleCacheInstance.Set(cache.CacheKey(userReq.ConversationID), moduleCache, cache.DefaultCacheExpiration)

	mockResponse := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Field '%s' has been deleted successfully", params.Fields),
		"params":  params.Fields,
	}

	result, err := json.MarshalIndent(mockResponse, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}
