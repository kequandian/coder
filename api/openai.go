package api

import (
	"github.com/cloudwego/eino/schema"
)

// ChatRequest represents an OpenAI API chat completion request
type ChatRequest struct {
	ID string `json:"id"`
	// 对话id
	ConversationID string `json:"conversation_id"`
	// 模型
	Model string `json:"model"`
	// 消息
	Messages []OpenAIMessage `json:"messages"`
	// 是否流式
	Stream bool `json:"stream"`
	// 温度
	Temperature float32 `json:"temperature,omitempty"`
	// 最大tokens
	MaxTokens int `json:"max_tokens,omitempty"`
}

// OpenAIMessage represents a message in an OpenAI API request or response
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse represents an OpenAI API chat completion response
type ChatResponse struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []ChatResponseChoice `json:"choices"`
	Usage   ChatResponseUsage    `json:"usage"`
}

// ChatResponseChoice represents a choice in an OpenAI API response
type ChatResponseChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// ChatResponseUsage represents token usage in an OpenAI API response
type ChatResponseUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatStreamResponse represents a streaming response from the OpenAI API
type ChatStreamResponse struct {
	ID      string                     `json:"id"`
	Object  string                     `json:"object"`
	Created int64                      `json:"created"`
	Model   string                     `json:"model"`
	Choices []ChatStreamResponseChoice `json:"choices"`
}

// ChatStreamResponseChoice represents a choice in a streaming OpenAI API response
type ChatStreamResponseChoice struct {
	Index        int           `json:"index"`
	Delta        OpenAIMessage `json:"delta"`
	FinishReason *string       `json:"finish_reason"`
}

// ConvertToSchema converts OpenAI messages to Eino schema messages
func ConvertToSchema(messages []OpenAIMessage) []*schema.Message {
	result := make([]*schema.Message, 0, len(messages))
	for _, msg := range messages {
		switch msg.Role {
		case "system":
			result = append(result, schema.SystemMessage(msg.Content))
		case "user":
			result = append(result, schema.UserMessage(msg.Content))
		case "assistant":
			result = append(result, schema.AssistantMessage(msg.Content, nil))
		}
	}
	return result
}

// ConvertToOpenAI converts Eino schema messages to OpenAI messages
func ConvertToOpenAI(messages []*schema.Message) []OpenAIMessage {
	result := make([]OpenAIMessage, 0, len(messages))
	for _, msg := range messages {
		result = append(result, OpenAIMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}
	return result
}
