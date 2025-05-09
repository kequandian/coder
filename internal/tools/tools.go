package tools

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"coder/internal/tools/addaction"
	"coder/internal/tools/addapi"
	"coder/internal/tools/addfield"
	"coder/internal/tools/addoperation"
	"coder/internal/tools/addsearch"
	"coder/internal/tools/deleteaction"
	"coder/internal/tools/deleteapi"
	"coder/internal/tools/deletefield"
	"coder/internal/tools/deleteoperation"
	"coder/internal/tools/deletesearch"
	"coder/internal/tools/editaction"
	"coder/internal/tools/editapi"
	"coder/internal/tools/editfield"
	"coder/internal/tools/editoperation"
	"coder/internal/tools/editsearch"
	"coder/internal/tools/genfield"
	"coder/internal/tools/saveentity"
	"coder/internal/tools/savemodule"
	"coder/internal/tools/viewmodule"
)

// ToolManager 管理所有本地工具
type ToolManager struct {
	tools map[string]tool.BaseTool
	mu    sync.RWMutex
}

// NewToolManager 创建一个新的工具管理器
func NewToolManager() *ToolManager {
	return &ToolManager{
		tools: make(map[string]tool.BaseTool),
	}
}

// Initialize 初始化所有工具
func (tm *ToolManager) Initialize(ctx context.Context) error {
	// 初始化模块生成工具
	// genCodeTool, err := gencode.NewGenCodeTool(ctx, cfg.OpenAI.APIKey, cfg.OpenAI.BaseURL, cfg.OpenAI.ModelID)
	// if err != nil {
	// 	return fmt.Errorf("failed to initialize code generation tool: %w", err)
	// }

	// // 注册工具
	// if err := tm.RegisterTool(ctx, genCodeTool); err != nil {
	// 	return fmt.Errorf("failed to register code generation tool: %w", err)
	// }

	// 初始化查看模块工具
	viewModuleTool, err := viewmodule.NewViewModuleTool()
	if err != nil {
		return fmt.Errorf("failed to initialize view module tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, viewModuleTool); err != nil {
		return fmt.Errorf("failed to register view module tool: %w", err)
	}

	// 初始化保存模块工具
	saveModuleTool, err := savemodule.NewSaveModuleTool()
	if err != nil {
		return fmt.Errorf("failed to initialize save module tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, saveModuleTool); err != nil {
		return fmt.Errorf("failed to register save module tool: %w", err)
	}

	// 初始化添加字段工具
	addFieldTool, err := addfield.NewAddFieldTool()
	if err != nil {
		return fmt.Errorf("failed to initialize add field tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, addFieldTool); err != nil {
		return fmt.Errorf("failed to register add field tool: %w", err)
	}

	// 初始化编辑字段工具
	editFieldTool, err := editfield.NewEditFieldTool()
	if err != nil {
		return fmt.Errorf("failed to initialize edit field tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, editFieldTool); err != nil {
		return fmt.Errorf("failed to register edit field tool: %w", err)
	}

	// 初始化删除字段工具
	deleteFieldTool, err := deletefield.NewDeleteFieldTool()
	if err != nil {
		return fmt.Errorf("failed to initialize delete field tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, deleteFieldTool); err != nil {
		return fmt.Errorf("failed to register delete field tool: %w", err)
	}

	// 初始化添加搜索工具
	addSearchTool, err := addsearch.NewAddSearchTool()
	if err != nil {
		return fmt.Errorf("failed to initialize add search tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, addSearchTool); err != nil {
		return fmt.Errorf("failed to register add search tool: %w", err)
	}

	// 初始化编辑搜索工具
	editSearchTool, err := editsearch.NewEditSearchTool()
	if err != nil {
		return fmt.Errorf("failed to initialize edit search tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, editSearchTool); err != nil {
		return fmt.Errorf("failed to register edit search tool: %w", err)
	}

	// 初始化删除搜索工具
	deleteSearchTool, err := deletesearch.NewDeleteSearchTool()
	if err != nil {
		return fmt.Errorf("failed to initialize delete search tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, deleteSearchTool); err != nil {
		return fmt.Errorf("failed to register delete search tool: %w", err)
	}

	// 初始化添加操作工具
	addOperationTool, err := addoperation.NewAddOperationTool()
	if err != nil {
		return fmt.Errorf("failed to initialize add operation tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, addOperationTool); err != nil {
		return fmt.Errorf("failed to register add operation tool: %w", err)
	}

	// 初始化编辑操作工具
	editOperationTool, err := editoperation.NewEditOperationTool()
	if err != nil {
		return fmt.Errorf("failed to initialize edit operation tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, editOperationTool); err != nil {
		return fmt.Errorf("failed to register edit operation tool: %w", err)
	}

	// 初始化删除操作工具
	deleteOperationTool, err := deleteoperation.NewDeleteOperationTool()
	if err != nil {
		return fmt.Errorf("failed to initialize delete operation tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, deleteOperationTool); err != nil {
		return fmt.Errorf("failed to register delete operation tool: %w", err)
	}

	// 初始化添加动作工具
	addActionTool, err := addaction.NewAddActionTool()
	if err != nil {
		return fmt.Errorf("failed to initialize add action tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, addActionTool); err != nil {
		return fmt.Errorf("failed to register add action tool: %w", err)
	}

	// 初始化编辑动作工具
	editActionTool, err := editaction.NewEditActionTool()
	if err != nil {
		return fmt.Errorf("failed to initialize edit action tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, editActionTool); err != nil {
		return fmt.Errorf("failed to register edit action tool: %w", err)
	}

	// 初始化删除动作工具
	deleteActionTool, err := deleteaction.NewDeleteActionTool()
	if err != nil {
		return fmt.Errorf("failed to initialize delete action tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, deleteActionTool); err != nil {
		return fmt.Errorf("failed to register delete action tool: %w", err)
	}

	// 初始化添加API工具
	addAPITool, err := addapi.NewAddAPITool()
	if err != nil {
		return fmt.Errorf("failed to initialize add API tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, addAPITool); err != nil {
		return fmt.Errorf("failed to register add API tool: %w", err)
	}

	// 初始化编辑API工具
	editAPITool, err := editapi.NewEditAPITool()
	if err != nil {
		return fmt.Errorf("failed to initialize edit API tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, editAPITool); err != nil {
		return fmt.Errorf("failed to register edit API tool: %w", err)
	}

	// 初始化删除API工具
	deleteAPITool, err := deleteapi.NewDeleteAPITool()
	if err != nil {
		return fmt.Errorf("failed to initialize delete API tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, deleteAPITool); err != nil {
		return fmt.Errorf("failed to register delete API tool: %w", err)
	}

	// 初始化生成字段工具
	genfieId, err := genfield.NewGenFieldTool()
	if err != nil {
		return fmt.Errorf("failed to initialize genfieId tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, genfieId); err != nil {
		return fmt.Errorf("failed to register genfieId tool: %w", err)
	}

	// 初始化保存实体工具
	saveEntityTool, err := saveentity.NewSaveEntityTool()
	if err != nil {
		return fmt.Errorf("failed to initialize save entity tool: %w", err)
	}
	if err := tm.RegisterTool(ctx, saveEntityTool); err != nil {
		return fmt.Errorf("failed to register save entity tool: %w", err)
	}

	log.Printf("Tool manager initialized with %d tools", len(tm.tools))
	return nil
}

// RegisterTool 注册一个工具
func (tm *ToolManager) RegisterTool(ctx context.Context, tool tool.BaseTool) error {
	info, err := tool.Info(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tool info: %w", err)
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 检查工具名称是否已存在
	if _, exists := tm.tools[info.Name]; exists {
		return fmt.Errorf("tool with name '%s' already registered", info.Name)
	}

	tm.tools[info.Name] = tool
	log.Printf("Registered tool: %s (%s)", info.Name, info.Desc)
	return nil
}

// GetAllTools 获取所有注册的工具
func (tm *ToolManager) GetAllTools() []tool.BaseTool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tools := make([]tool.BaseTool, 0, len(tm.tools))
	for _, t := range tm.tools {
		tools = append(tools, t)
	}
	return tools
}

// GetToolByName 根据名称获取工具
func (tm *ToolManager) GetToolByName(name string) (tool.BaseTool, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	t, exists := tm.tools[name]
	return t, exists
}

// GetAllToolInfos 获取所有工具的信息
func (tm *ToolManager) GetAllToolInfos(ctx context.Context) ([]*schema.ToolInfo, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	toolInfos := make([]*schema.ToolInfo, 0, len(tm.tools))
	for _, t := range tm.tools {
		info, err := t.Info(ctx)
		if err != nil {
			log.Printf("Failed to get info for tool: %v", err)
			continue
		}
		toolInfos = append(toolInfos, info)
	}
	return toolInfos, nil
}

// ExecuteTool 执行指定的工具
func (tm *ToolManager) ExecuteTool(ctx context.Context, toolName string, arguments string) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			buf := new(bytes.Buffer)
			buf.WriteString("Recovered from panic in ExecuteTool:\n")
			buf.WriteString(fmt.Sprintf("%v", r))
			buf.WriteString("\n")
			buf.WriteString(string(debug.Stack()))
			log.Printf(buf.String())
		}
	}()

	// 获取工具
	tm.mu.RLock()
	selectTools, exists := tm.tools[toolName]
	tm.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("tool '%s' not found", toolName)
	}

	// 检查工具是否可调用
	invokableTool, ok := selectTools.(tool.InvokableTool)
	if !ok {
		return "", fmt.Errorf("tool '%s' is not invokable", toolName)
	}

	// 执行工具
	return invokableTool.InvokableRun(ctx, arguments)
}
