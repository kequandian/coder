package nodelog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	callbacks2 "github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/eino/utils/callbacks"
)

const (
	nodeKeyPlanner        = "planner"          // planner 智能体的节点 key
	nodeKeyExecutor       = "executor"         // executor 智能体的节点 key
	nodeKeyReviser        = "reviser"          // reviser 智能体的节点 key
	nodeKeyTools          = "tools"            // tools 执行器的节点 key
	nodeKeyPlannerToList  = "planner_to_list"  // planner->executor 之间的 converter 节点 key
	nodeKeyExecutorToList = "executor_to_list" // executor->reviser 之间的 converter 节点 key
	nodeKeyReviserToList  = "reviser_to_list"  // reviser->executor 之间的 converter 节点 key
	defaultMaxStep        = 100                // 默认的最大执行步骤数量
)

type coloredString struct {
	str  string
	code string
}

// Nodelog 利用 Eino 的 callback 机制，收集多智能体各步骤的实时输出.
type Nodelog struct {
	ch               chan coloredString
	currentAgentName string          // 当前智能体名称
	agentReasoning   map[string]bool // 智能体处在"推理"阶段还是"最终答案"阶段
	mu               sync.Mutex
	wg               sync.WaitGroup
}

func NewNodelog() *Nodelog {
	return &Nodelog{
		ch: make(chan coloredString),
		agentReasoning: map[string]bool{
			nodeKeyPlanner:  false,
			nodeKeyExecutor: false,
			nodeKeyReviser:  false,
		},
	}
}

func (s *Nodelog) PrintStream() {
	go func() {
		for m := range s.ch {
			fmt.Print(m.code + m.str + Reset)
		}
	}()
}

func (s *Nodelog) ToCallbackHandler() callbacks2.Handler {
	// 使用HandlerBuilder创建所有回调处理器
	builder := callbacks2.NewHandlerBuilder()

	// 所有节点通用回调
	builder.OnStartFn(func(ctx context.Context, info *callbacks2.RunInfo, input callbacks2.CallbackInput) context.Context {
		// 检查是否是模板节点
		if info.Component == "ChatTemplate" || info.Type == "prompt.ChatTemplate" || info.Name == "ChatTemplate" {
			s.ch <- coloredString{fmt.Sprintf("\n\n=======\nChatTemplate Start [%s]: \n=======\n", info.Name), Cyan}

			// 尝试将输入格式化为JSON
			inputJSON, err := sonic.MarshalIndent(input, "  ", "  ")
			if err == nil {
				s.ch <- coloredString{fmt.Sprintf("Template Input: %s\n", string(inputJSON)), Purple}
			} else {
				s.ch <- coloredString{fmt.Sprintf("Template Input: %v\n", input), Purple}
			}
		}
		return ctx
	})

	builder.OnEndFn(func(ctx context.Context, info *callbacks2.RunInfo, output callbacks2.CallbackOutput) context.Context {
		// 检查是否是模板节点
		if info.Component == "ChatTemplate" || info.Type == "prompt.ChatTemplate" || info.Name == "ChatTemplate" {
			s.ch <- coloredString{fmt.Sprintf("\n=======\nChatTemplate End [%s]: \n=======\n", info.Name), Cyan}

			// 尝试将输出格式化为JSON
			outputJSON, err := sonic.MarshalIndent(output, "  ", "  ")
			if err == nil {
				s.ch <- coloredString{fmt.Sprintf("Template Output: %s\n", string(outputJSON)), Green}
			} else {
				s.ch <- coloredString{fmt.Sprintf("Template Output: %v\n", output), Green}
			}
		}
		return ctx
	})

	// 为模型添加特定回调
	modelHandler := &callbacks.ModelCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks2.RunInfo, input *model.CallbackInput) context.Context {
			s.ch <- coloredString{fmt.Sprintf("\n\n=======\nLLM Request [%s]: \n=======\n", info.Name), Magenta}

			if input != nil {
				// 打印请求消息
				if len(input.Messages) > 0 {
					s.ch <- coloredString{"Request Messages:\n", Green}

					for i, msg := range input.Messages {
						s.ch <- coloredString{fmt.Sprintf("[%d] Role: %s\n", i, msg.Role), Purple}
						s.ch <- coloredString{fmt.Sprintf("Content: %s\n\n", msg.Content), White}
					}
				}

				// 尝试打印完整请求信息
				inputJSON, err := sonic.MarshalIndent(input, "  ", "  ")
				if err == nil {
					s.ch <- coloredString{fmt.Sprintf("Full Request: %s\n", string(inputJSON)), Gray}
				}
			}

			return ctx
		},
		OnEndWithStreamOutput: func(ctx context.Context, runInfo *callbacks2.RunInfo, output *schema.StreamReader[*model.CallbackOutput]) context.Context {
			name := runInfo.Name
			s.ch <- coloredString{fmt.Sprintf("\n\n=======\nLLM Response type:[%s]: \n=======\n", runInfo.Type), Cyan}
			if name != s.currentAgentName {
				s.ch <- coloredString{fmt.Sprintf("\n\n=======\nLLM Response name:[%s]: \n=======\n", name), Cyan}
				s.currentAgentName = name
			}

			s.wg.Add(1)

			go func() {
				defer output.Close()
				defer s.wg.Done()

				for {
					chunk, err := output.Recv()
					if err != nil {
						if errors.Is(err, io.EOF) {
							break
						}
						log.Fatalf("internal error: %s\n", err)
					}

					data1, _ := json.Marshal(chunk.Message)
					s.ch <- coloredString{"\nanswer data: \n", string(data1)}
					if len(chunk.Message.Content) > 0 {
						if s.agentReasoning[name] { // 切换到最终答案阶段
							s.ch <- coloredString{"\nanswer begin: \n", Green}
							s.mu.Lock()
							s.agentReasoning[name] = false
							s.mu.Unlock()
						}
						s.ch <- coloredString{chunk.Message.Content, Yellow}
					} else if reasoningContent, ok := deepseek.GetReasoningContent(chunk.Message); ok {
						if !s.agentReasoning[name] { // 切换到推理阶段
							s.ch <- coloredString{"\nreasoning begin: \n", Green}
							s.mu.Lock()
							s.agentReasoning[name] = true
							s.mu.Unlock()
						}
						s.ch <- coloredString{reasoningContent, White}
					}
				}
			}()

			return ctx
		},
	}

	// 为工具添加特定回调
	toolHandler := &callbacks.ToolCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks2.RunInfo, input *tool.CallbackInput) context.Context {
			arguments := make(map[string]any)
			err := sonic.Unmarshal([]byte(input.ArgumentsInJSON), &arguments)
			if err != nil {
				s.ch <- coloredString{fmt.Sprintf("\ncall %s: %s\n", info.Name, input.ArgumentsInJSON), Red}
				return ctx
			}

			formatted, err := sonic.MarshalIndent(arguments, "  ", "  ")
			if err != nil {
				s.ch <- coloredString{fmt.Sprintf("\ncall %s: %s\n", info.Name, input.ArgumentsInJSON), Red}
				return ctx
			}

			s.ch <- coloredString{fmt.Sprintf("\ncall %s: %s\n", info.Name, string(formatted)), Red}
			return ctx
		},
		OnEnd: func(ctx context.Context, info *callbacks2.RunInfo, output *tool.CallbackOutput) context.Context {
			response := make(map[string]any)
			err := sonic.Unmarshal([]byte(output.Response), &response)
			if err != nil {
				s.ch <- coloredString{fmt.Sprintf("\ncall %s: %s\n", info.Name, output.Response), Blue}
				return ctx
			}

			formatted, err := sonic.MarshalIndent(response, "  ", "  ")
			if err != nil {
				s.ch <- coloredString{fmt.Sprintf("\ncall %s: %s\n", info.Name, output.Response), Blue}
				return ctx
			}

			s.ch <- coloredString{fmt.Sprintf("\ncall %s result: %s\n", info.Name, string(formatted)), Blue}
			return ctx
		},
	}

	// 创建处理器帮助类
	helper := callbacks.NewHandlerHelper()
	helper.ChatModel(modelHandler)
	helper.Tool(toolHandler)

	// 返回生成的处理器
	return helper.Handler()
}

func (s *Nodelog) wait() {
	s.wg.Wait()
}

const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	White   = "\033[97m"
	Cyan    = "\033[36m"
	Gray    = "\033[37m"
	Purple  = "\033[35m"
	Magenta = "\033[95m"
	Orange  = "\033[38;5;208m"
)
