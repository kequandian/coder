package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"coder/api"
	"coder/app"
	"coder/internal/config"
	"coder/internal/mcp"
	"coder/internal/models"
	"coder/internal/tools"
	"coder/nodelog"
)

// Agent represents an Eino agent for chat
type Agent struct {
	einoGraph   compose.Runnable[map[string]any, *schema.Message]
	mcpManager  *mcp.MCPManager
	toolManager *tools.ToolManager
}

// New creates a new chat agent
func New(ctx context.Context) (*Agent, error) {
	// Initialize MCP manager
	mcpManager := mcp.NewMCPManager()
	err := mcpManager.Initialize(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MCP manager: %w", err)
	}

	// Start MCP health checker
	mcpManager.StartHealthChecker(ctx)

	// Initialize Tool manager
	toolManager := tools.NewToolManager()
	err = toolManager.Initialize(ctx)
	if err != nil {
		mcpManager.Close() // Clean up MCP connections on error
		return nil, fmt.Errorf("failed to initialize tool manager: %w", err)
	}

	// Create the Eino graph with tools
	einoGraph, err := createEinoGraph(ctx, mcpManager, toolManager)
	if err != nil {
		mcpManager.Close() // Clean up MCP connections on error
		return nil, err
	}

	return &Agent{
		einoGraph:   einoGraph,
		mcpManager:  mcpManager,
		toolManager: toolManager,
	}, nil
}

// Generate generates a single response from the agent
func (a *Agent) Generate(ctx context.Context, req *api.ChatRequest, systemPrompt string, chatHistory []*schema.Message, userQuery string) (*schema.Message, error) {
	// Use the graph with branch logic to handle tool calls
	message, err := a.einoGraph.Invoke(ctx, map[string]any{
		"system_prompt": systemPrompt,
		"chat_history":  chatHistory,
		"user_query":    userQuery,
	})

	if err != nil {
		return nil, err
	}

	return message, nil
}

// Stream streams responses from the agent
func (a *Agent) Stream(ctx context.Context, req *api.ChatRequest, systemPrompt string, chatHistory []*schema.Message, userQuery string) (*schema.StreamReader[*schema.Message], error) {
	// For streaming, we use the graph with branch logic
	printer := nodelog.NewNodelog() // 创建一个中间结果打印器
	printer.PrintStream()           // 开始异步输出到 console
	handler := printer.ToCallbackHandler()
	newCtx := context.WithValue(ctx, config.StateKey, req)
	return a.einoGraph.Stream(newCtx, map[string]any{
		"system_prompt": systemPrompt,
		"chat_history":  chatHistory,
		"user_query":    userQuery,
	}, compose.WithCallbacks(handler))
}

// Close cleans up resources used by the agent
func (a *Agent) Close() {
	if a.mcpManager != nil {
		a.mcpManager.Close()
	}
}

// ToolCall represents a simplified tool call structure for extraction
type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

// extractToolCalls extracts tool calls from message, handling different schema versions
func extractToolCalls(message *schema.Message) []ToolCall {
	var toolCalls []ToolCall

	// 检查OpenAI格式的ToolCalls
	if len(message.ToolCalls) > 0 {
		for _, tc := range message.ToolCalls {
			if tc.Function.Name != "" {
				toolCalls = append(toolCalls, ToolCall{
					ID:        tc.ID,
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				})
			}
		}
		return toolCalls
	}

	// 尝试从消息内容解析JSON
	var response struct {
		ToolCalls []struct {
			ID       string `json:"id"`
			Function struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			} `json:"function"`
		} `json:"tool_calls"`
	}

	// 尝试将内容解析为JSON对象
	if err := json.Unmarshal([]byte(message.Content), &response); err == nil {
		if len(response.ToolCalls) > 0 {
			for _, tc := range response.ToolCalls {
				if tc.Function.Name != "" {
					toolCalls = append(toolCalls, ToolCall{
						ID:        tc.ID,
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					})
				}
			}
		}
	}

	return toolCalls
}

// createEinoGraph creates a graph for the chat agent with branch logic for tool handling
func createEinoGraph(ctx context.Context, mcpManager *mcp.MCPManager, toolManager *tools.ToolManager) (compose.Runnable[map[string]any, *schema.Message], error) {
	// 创建模板
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage("{system_prompt}"),
		schema.MessagesPlaceholder("chat_history", true),
		schema.UserMessage("{user_query}"),
	)

	// 创建聊天模型
	var chatModel model.ChatModel
	var err error
	modelConfig := &openai.ChatModelConfig{
		Model:  app.Config.OpenAI.ModelID,
		APIKey: app.Config.OpenAI.APIKey,
	}

	if app.Config.OpenAI.BaseURL != "" {
		modelConfig.BaseURL = app.Config.OpenAI.BaseURL
	}

	if app.Config.OpenAI.MaxTokens > 0 {
		maxTokens := app.Config.OpenAI.MaxTokens
		modelConfig.MaxTokens = &maxTokens
	}

	chatModel, err = openai.NewChatModel(ctx, modelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	// 组合所有工具
	allTools := append(toolManager.GetAllTools(), mcpManager.GetAllTools()...)

	// 如果有工具，绑定到模型
	if len(allTools) > 0 {
		toolInfos := make([]*schema.ToolInfo, 0, len(allTools))

		// 添加MCP工具
		for _, t := range allTools {
			info, err := t.Info(ctx)
			if err != nil {
				fmt.Printf("Failed to get tool info: %v\n", err)
				continue
			}
			json, _ := json.Marshal(info)
			log.Printf("toolInfos: %v", string(json))
			toolInfos = append(toolInfos, info)
		}

		if len(toolInfos) > 0 {
			err = chatModel.BindTools(toolInfos)
			if err != nil {
				return nil, fmt.Errorf("failed to bind tools to model: %w", err)
			}
		}
	}

	// 定义节点名称常量
	const (
		nodePrompt    = "prompt"     // 提示模板节点
		nodeModel     = "model"      // 模型节点
		nodeLocalTool = "LocalTool"  // 本地工具节点
		nodeMcpTool   = "InvorkTool" // MCP调用工具节点
	)

	// 创建分支函数，用于判断是否有工具调用以及调用哪个工具
	branch := compose.NewStreamGraphBranch(func(ctx context.Context, input *schema.StreamReader[*schema.Message]) (string, error) {
		log.Printf("branch start Received streaming message")
		defer input.Close()

		var hasToolCall bool
		var isLocalTool bool
		msgCount := 0
		startTime := time.Now()
		for {
			// 添加超时检查
			if time.Since(startTime) > 3*time.Second {
				log.Printf("Exceeded 3 second timeout, stopping stream check")
				break
			}

			// 添加消息数量检查
			if msgCount >= 1 {
				log.Printf("Received message chunk, stopping stream check")
				break
			}
			msgCount++

			msg, err := input.Recv()
			if err != nil {
				break
			}

			toolCalls := extractToolCalls(msg)
			if len(toolCalls) > 0 {
				log.Printf("Detected tool calls in stream: %v", toolCalls)
				hasToolCall = true

				// 检查是否是本地工具
				for _, tc := range toolCalls {
					if _, exists := toolManager.GetToolByName(tc.Name); exists {
						isLocalTool = true
						break
					}
				}
				break
			}
		}

		// 根据是否有工具调用决定路由
		if hasToolCall {
			log.Printf("Stream complete, routing to tool execution")
			if isLocalTool {
				log.Printf("Using LocalTool")
				return nodeLocalTool, nil
			}
			return nodeMcpTool, nil
		}

		log.Printf("Stream complete, no tools detected")
		return compose.END, nil
	}, map[string]bool{compose.END: true, nodeMcpTool: true, nodeLocalTool: true})

	// 创建本地工具节点
	localToolNode := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
		if input == nil {
			return nil, fmt.Errorf("no message received")
		}
		// 获取原始消息
		originalMessage := input
		log.Printf("Processing LocalTool calls in message: %v", originalMessage)

		// 提取工具调用
		toolCalls := extractToolCalls(originalMessage)
		if len(toolCalls) == 0 {
			// 如果没有工具调用，直接返回原始消息
			return originalMessage, nil
		}

		// 创建工具调用结果列表
		toolResultMessages := make([]*schema.Message, 0, len(toolCalls))

		// 处理每个工具调用
		for _, tc := range toolCalls {
			// 检查是否是本地工具
			if _, exists := toolManager.GetToolByName(tc.Name); !exists {
				continue // 跳过非本地工具
			}

			log.Printf("Processing LocalTool call: %s with arguments: %s", tc.Name, tc.Arguments)

			// 执行本地工具
			result, err := toolManager.ExecuteTool(ctx, tc.Name, tc.Arguments)

			// 创建工具调用结果消息
			toolResultMsg := &schema.Message{
				Role:       "tool",
				ToolCallID: tc.ID,
				Name:       tc.Name,
			}

			// 处理执行结果或错误
			if err != nil {
				log.Printf("Error executing LocalTool %s: %v", tc.Name, err)
				toolResultMsg.Content = fmt.Sprintf("Error executing tool: %v", err)
			} else {
				toolResultMsg.Content = fmt.Sprintf("\n```json\n%s\n```", result)
			}

			// 添加工具结果消息
			toolResultMessages = append(toolResultMessages, toolResultMsg)
		}

		// 通过再次调用模型来获取最终结果
		finalMsg := &schema.Message{
			Role:    "assistant",
			Content: "我已经处理了您的请求，结果如下:\n\n",
		}

		// 添加每个工具的结果
		for _, toolResult := range toolResultMessages {
			finalMsg.Content += fmt.Sprintf("**工具**: %s\n**结果**: %s\n\n", toolResult.Name, toolResult.Content)
		}

		return finalMsg, nil
	})

	// 将格式化工具执行结果的Lambda函数
	invorkTool := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
		if input == nil {
			return nil, fmt.Errorf("no message received")
		}

		// 获取原始消息
		originalMessage := input
		log.Printf("Processing tool calls in message: %v", originalMessage)

		// 提取工具调用
		toolCalls := extractToolCalls(originalMessage)
		if len(toolCalls) == 0 {
			// 如果没有工具调用，直接返回原始消息
			return originalMessage, nil
		}

		log.Printf("Found %d tool calls to process", len(toolCalls))

		// 创建工具调用结果列表
		content := "没处理结果"
		toolName := ""
		for _, tc := range toolCalls {
			// 跳过本地工具调用，因为它有自己的处理节点
			if _, exists := toolManager.GetToolByName(tc.Name); exists {
				continue
			}

			log.Printf("Processing MCP tool call: %s with arguments: %s", tc.Name, tc.Arguments)

			// 调用工具并获取结果
			result, err := executeMCPTool(ctx, mcpManager, tc.Name, tc.Arguments)
			toolName = tc.Name
			// 处理执行结果或错误
			if err != nil {
				log.Printf("Error executing tool %s: %v", tc.Name, err)
				content = fmt.Sprintf("Error executing tool: %v", err)
			} else {
				callToolResult := &models.MCPResult{}
				err = json.Unmarshal([]byte(result), &callToolResult)
				if err != nil {
					log.Printf("Error unmarshalling tool result: %v", err)
					content = result
				} else {
					content = fmt.Sprintf("%v", callToolResult.Content[0].Text)
				}
			}
			break
		}

		// 通过再次调用模型来获取最终结果
		// 这里应该使用一个新的模型调用，但为了简化流程，我们将返回一个特殊的消息
		// 在真实系统中，应该重新调用模型或使用另一个节点来处理
		finalMsg := &schema.Message{
			Role:    "assistant",
			Content: "我已经处理了您的工具1调用请求，结果如下:\n\n",
		}

		// 添加每个工具的结果
		finalMsg.Content += fmt.Sprintf("**工具**: %s\n**结果**: %s\n\n", toolName, content)
		return finalMsg, nil
	})

	// 创建图实例
	g := compose.NewGraph[map[string]any, *schema.Message]()
	// 添加节点
	_ = g.AddChatTemplateNode(nodePrompt, template, compose.WithNodeName("ChatTemplate"))
	_ = g.AddChatModelNode(nodeModel, chatModel, compose.WithNodeName("LLMModel"))
	// 添加本地工具节点
	_ = g.AddLambdaNode(nodeLocalTool, localToolNode, compose.WithNodeName("LocalToolExecutor"))
	// 添加MCP工具节点
	_ = g.AddLambdaNode(nodeMcpTool, invorkTool, compose.WithNodeName("McpToolExecutor"))
	// 连接节点
	_ = g.AddEdge(compose.START, nodePrompt)
	_ = g.AddEdge(nodePrompt, nodeModel)
	_ = g.AddBranch(nodeModel, branch)        // 添加分支，决定是否需要调用工具
	_ = g.AddEdge(nodeLocalTool, compose.END) // 本地工具结果返回
	_ = g.AddEdge(nodeMcpTool, compose.END)   // MCP工具结果返回

	// 编译图
	r, err := g.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile graph: %w", err)
	}

	return r, nil
}

// executeMCPTool 执行MCP工具调用
func executeMCPTool(ctx context.Context, mcpManager *mcp.MCPManager, toolName string, arguments string) (string, error) {
	// 解析参数
	var args map[string]interface{}
	// 尝试解析参数
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		log.Printf("Warning: Cannot parse arguments after fixing: %v", err)
		args = make(map[string]interface{})
	}
	// 否则，尝试在所有服务器上查找并执行
	clients := mcpManager.GetAllClients()
	for _, client := range clients {
		tools := mcpManager.GetClientTools(client)
		for _, t := range tools {
			info, err := t.Info(ctx)
			if err != nil {
				continue
			}

			if info.Name == toolName {
				// 找到匹配的工具，执行
				result, err := client.ExecuteTool(ctx, toolName, args)
				if err == nil {
					return result, nil
				}
				log.Printf("Failed to execute tool %s on %s: %v", toolName, client.GetName(), err)
			}
		}
	}

	return "", fmt.Errorf("tool '%s' not found or execution failed on all servers", toolName)
}

// GetToolsInfo returns information about MCP tools and local tools
func (a *Agent) GetToolsInfo(ctx context.Context) []map[string]interface{} {
	var result []map[string]interface{}
	// 获取MCP工具信息
	if a.mcpManager != nil {
		mcpTools := a.mcpManager.GetAllTools()
		for _, t := range mcpTools {
			info, err := t.Info(ctx)
			if err != nil {
				continue
			}

			toolInfo := map[string]interface{}{
				"name":        info.Name,
				"description": info.Desc,
				"type":        "mcp",
			}

			result = append(result, toolInfo)
		}
	}

	// 获取本地工具信息
	if a.toolManager != nil {
		localTools := a.toolManager.GetAllTools()
		for _, t := range localTools {
			info, err := t.Info(ctx)
			if err != nil {
				continue
			}

			toolInfo := map[string]interface{}{
				"name":        info.Name,
				"description": info.Desc,
				"type":        "local",
			}

			result = append(result, toolInfo)
		}
	}

	return result
}
