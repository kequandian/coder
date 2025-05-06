package addoperation

import (
	"coder/api"
	"coder/internal/cache"
	"coder/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// AddOperationTool 是用于向模块添加操作的工具
// 该工具负责将新的操作添加到模块中，支持创建、更新和删除操作
type AddOperationTool struct{}

// NewAddOperationTool 创建并返回一个新的AddOperationTool实例
// 返回值: AddOperationTool指针和可能的错误信息
func NewAddOperationTool() (*AddOperationTool, error) {
	return &AddOperationTool{}, nil
}

// Info 返回工具的基本信息，包括名称、描述和参数定义
// 参数:
//
//	ctx - 上下文
//
// 返回值: ToolInfo指针和可能的错误信息
func (t *AddOperationTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "addOperation",
		Desc: "Add an operation to a module",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"title": {
				Desc:     "The title of the operation (e.g., '详情')",
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

// IsInvokable 返回true表示该工具可以被调用
// 该工具始终可调用，因此固定返回true
func (t *AddOperationTool) IsInvokable() bool {
	return true
}

// InvokableRun 执行工具的主要逻辑，添加操作到模块中
// 参数:
//
//	ctx - 上下文
//	args - 包含title, type和options的JSON字符串
//	_ - 可选参数列表
//
// 返回值: 操作结果字符串和可能的错误信息
// 参数：
//
//	ctx - 上下文
//	args - 包含title, type和options的JSON字符串
//
// 返回值：
//
//	成功返回操作结果，失败返回错误信息
func (t *AddOperationTool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
	// 解析输入参数
	// 将JSON字符串反序列化为结构体，包含title, type和options字段
	var params struct {
		Title   string                 `json:"title"`
		Type    string                 `json:"type"`
		Options map[string]interface{} `json:"options"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// 验证必填参数
	// 确保title, type和options都不为空，否则返回错误
	if params.Title == "" {
		return "", fmt.Errorf("title cannot be empty")
	}
	if params.Type == "" {
		return "", fmt.Errorf("type cannot be empty")
	}
	if params.Options == nil {
		return "", fmt.Errorf("options cannot be empty")
	}

	// 在实际实现中，我们将操作添加到模块中
	// 从上下文中获取用户请求信息，并记录日志
	userReq, ok := ctx.Value(config.StateKey).(*api.ChatRequest)
	if !ok {
		return "", fmt.Errorf("state not found in context")
	}
	log.Printf("Processing LocalTool calls in message: %v", userReq)

	// 从上下文中解码模块信息
	// 获取模块缓存数据和当前模块状态
	infoCache, mapCur, _, userReq, err := cache.DecodeModuleFromCtx(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to decode module from context: %w", err)
	}

	tableOperation := mapCur["tableOperation"].([]interface{})
	// 创建title到index的映射
	// 用于快速查找已存在的操作
	title2Index := make(map[string]int)
	for index, field := range tableOperation {
		title2Index[field.(map[string]interface{})["title"].(string)] = index
	}
	// 更新或添加操作
	// 如果操作已存在则更新，否则添加新操作
	if index, ok := title2Index[params.Title]; ok {
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
	// 保存到缓存
	// 将修改后的模块信息序列化并存储到缓存中，设置默认过期时间
	jsonCur, _ := json.Marshal(mapCur)
	moduleCache := cache.NewModuleCacheData(infoCache.ModuleName, infoCache.ModuleCode, infoCache.Support, string(jsonCur))
	cache.ModuleCacheInstance.Set(cache.CacheKey(userReq.ConversationID), moduleCache, cache.DefaultCacheExpiration)

	// 格式化响应
	// 构造成功响应信息，包含操作结果和参数详情
	mockResponse := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Operation '%s' has been added successfully", params.Title),
		"params":  params,
	}
	result, err := json.MarshalIndent(mockResponse, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}
