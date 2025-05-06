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
// AddActionTool 是用于向模块添加操作的工具
type AddActionTool struct{}

// NewAddActionTool creates a new add action tool
// NewAddActionTool 创建并返回一个新的AddActionTool实例
func NewAddActionTool() (*AddActionTool, error) {
	return &AddActionTool{}, nil
}

// Info returns information about the tool
// Info 返回工具的基本信息，包括名称、描述和参数定义
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
// IsInvokable 返回true表示该工具可以被调用
func (t *AddActionTool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
// InvokableRun 执行工具的主要逻辑，添加操作到模块中
// 参数：
//
//	ctx - 上下文
//	args - 包含title, type和options的JSON字符串
//
// 返回值：
//
//	成功返回操作结果，失败返回错误信息
func (t *AddActionTool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
	// 解析输入参数
	var params struct {
		Title   string                 `json:"title"`
		Type    string                 `json:"type"`
		Options map[string]interface{} `json:"options"`
	}

	// 将JSON字符串反序列化为结构体
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// 验证必填参数
	if params.Title == "" {
		return "", fmt.Errorf("title cannot be empty")
	}
	if params.Type == "" {
		return "", fmt.Errorf("type cannot be empty")
	}
	if params.Options == nil {
		return "", fmt.Errorf("options cannot be empty")
	}

	// 从上下文中解码模块信息
	infoCache, mapCur, _, userReq, err := cache.DecodeModuleFromCtx(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to decode module from context: %w", err)
	}

	// 获取当前模块的tableActions，如果不存在则初始化
	tableAction, ok := mapCur["tableActions"].([]interface{})
	if !ok {
		tableAction = make([]interface{}, 0)
	}

	// 创建title到index的映射，用于快速查找已存在的action
	title2Index := make(map[string]int)
	for index, action := range tableAction {
		actionMap := action.(map[string]interface{})
		if title, ok := actionMap["title"].(string); ok {
			title2Index[title] = index
		}
	}

	// 如果action已存在则更新，否则添加新action
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

	// 将修改后的模块信息序列化并保存到缓存
	jsonCur, _ := json.Marshal(mapCur)
	moduleCache := cache.NewModuleCacheData(infoCache.ModuleName, infoCache.ModuleCode, infoCache.Support, string(jsonCur))
	cache.ModuleCacheInstance.Set(cache.CacheKey(userReq.ConversationID), moduleCache, cache.DefaultCacheExpiration)

	// 构造并返回成功响应
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
