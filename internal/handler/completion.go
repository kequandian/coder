package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"

	"coder/api"
	"coder/app"
	"coder/internal/agent"
	"coder/internal/cache"
	"coder/internal/models"
)

// Handler handles HTTP requests
type Handler struct {
	agent         *agent.Agent
	conversations *models.ConversationStore
}

// New creates a new handler
func New(agent *agent.Agent) *Handler {
	return &Handler{
		agent:         agent,
		conversations: models.NewConversationStore(),
	}
}

// HandleChatCompletion handles chat completion requests
func (h *Handler) HandleChatCompletion(w http.ResponseWriter, r *http.Request) {
	var req api.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		req.Model = app.Config.OpenAI.ModelID
	}

	// 如果请求中没有指定temperature，使用配置中的默认值
	if req.Temperature == 0 {
		req.Temperature = 0.7 // 默认值
	}

	// 如果请求中没有指定max_tokens，使用配置中的默认值
	if req.MaxTokens == 0 && app.Config.OpenAI.MaxTokens > 0 {
		req.MaxTokens = app.Config.OpenAI.MaxTokens
	}

	ctx := r.Context()

	// Prepare Eino input
	schemaMessages := api.ConvertToSchema(req.Messages)

	// Streaming response
	if req.Stream {
		h.handleStreamingResponse(ctx, w, &req, schemaMessages)
	} else {
		h.handleNonStreamingResponse(ctx, w, &req, schemaMessages)
	}
}

// handleStreamingResponse handles streaming chat completion requests
func (h *Handler) handleStreamingResponse(ctx context.Context, w http.ResponseWriter, req *api.ChatRequest, schemaMessages []*schema.Message) {
	// Initialize SSE writer
	sseWriter, err := NewSSEWriter(w)
	if err != nil {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Generate random ID for this chat completion
	created := unixTimestamp()
	// Process last user message and prepare chat history
	userQuery, chatHistory := extractQueryAndHistory(schemaMessages)

	// Stream response using Eino

	// 从缓存中获取模块名称 模块代码
	// 获取cache
	cacheKey := cache.CacheKey(req.ConversationID)
	info, ok := cache.ModuleCacheInstance.Get(cacheKey)
	if ok {
		// 根据缓存数据类型添加不同的上下文信息
		switch v := info.(type) {
		case *cache.ModuleCacheData:
			chatHistory = append(chatHistory, &schema.Message{
				Role:    "user",
				Content: "模块名称：" + v.ModuleName + "，模块代码：" + v.ModuleCode,
			})
			chatHistory = append(chatHistory, &schema.Message{
				Role:    "user",
				Content: fmt.Sprintf("这个是当前模块约束的配置，所有增加都需要在该配置里：%s", v.Support),
			})
			chatHistory = append(chatHistory, &schema.Message{
				Role:    "user",
				Content: fmt.Sprintf("这个是最新的配置，所有的调整都是基于该配置调整的：%s", v.Cur),
			})
		case *cache.EntityCacheData:
			chatHistory = append(chatHistory, &schema.Message{
				Role:    "user",
				Content: "实体名称：" + v.EntityName,
			})
			chatHistory = append(chatHistory, &schema.Message{
				Role:    "user",
				Content: fmt.Sprintf("实体相关配置：%s", v.Config),
			})
		}
	}

	sr, err := h.agent.Stream(ctx, req, app.Config.Chat.SystemPrompt, chatHistory, userQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer sr.Close()

	content := ""
	for {
		chunk, err := sr.Recv()
		if errors.Is(err, io.EOF) {
			// Send final chunk with finish_reason
			finishReason := "stop"
			finalResponse := api.ChatStreamResponse{
				ID:      req.ID,
				Object:  "chat.completion.chunk",
				Created: created,
				Model:   req.Model,
				Choices: []api.ChatStreamResponseChoice{
					{
						Index:        0,
						Delta:        api.OpenAIMessage{},
						FinishReason: &finishReason,
					},
				},
			}

			// Write final response and done marker
			sseWriter.WriteEvent(finalResponse)
			sseWriter.WriteDone()
			break
		}
		if err != nil {
			log.Printf("Error receiving stream chunk: %v", err)
			break
		}

		content += chunk.Content

		// Create OpenAI-compatible chunk response
		response := api.ChatStreamResponse{
			ID:      req.ID,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   req.Model,
			Choices: []api.ChatStreamResponseChoice{
				{
					Index: 0,
					Delta: api.OpenAIMessage{
						Role:    "assistant",
						Content: chunk.Content,
					},
					FinishReason: nil,
				},
			},
		}

		sseWriter.WriteEvent(response)
	}
}

// handleNonStreamingResponse handles non-streaming chat completion requests
func (h *Handler) handleNonStreamingResponse(ctx context.Context, w http.ResponseWriter, req *api.ChatRequest, schemaMessages []*schema.Message) {
	// Process last user message and prepare chat history
	userQuery, chatHistory := extractQueryAndHistory(schemaMessages)

	// Generate response using Eino
	result, err := h.agent.Generate(ctx, req, app.Config.Chat.SystemPrompt, chatHistory, userQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create OpenAI-compatible response
	response := api.ChatResponse{
		ID:      req.ID,
		Object:  "chat.completion",
		Created: unixTimestamp(),
		Model:   req.Model,
		Choices: []api.ChatResponseChoice{
			{
				Index: 0,
				Message: api.OpenAIMessage{
					Role:    "assistant",
					Content: result.Content,
				},
				FinishReason: "stop",
			},
		},
		Usage: api.ChatResponseUsage{
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// extractQueryAndHistory extracts the user query and chat history from messages
func extractQueryAndHistory(messages []*schema.Message) (string, []*schema.Message) {
	var userQuery string
	var chatHistory []*schema.Message

	if len(messages) > 0 && messages[len(messages)-1].Role == "user" {
		userQuery = messages[len(messages)-1].Content
	} else {
		userQuery = "Hello"
	}

	if len(messages) > 1 {
		chatHistory = messages[:len(messages)-1]
	} else {
		chatHistory = []*schema.Message{}
	}

	return userQuery, chatHistory
}

func unixTimestamp() int64 {
	return time.Now().Unix()
}

// HandleHealthCheck handles health check requests
func (h *Handler) HandleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
