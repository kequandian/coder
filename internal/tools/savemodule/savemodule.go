package savemodule

import (
	"bytes"
	"coder/api"
	"coder/app"
	"coder/internal/cache"
	"coder/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// SaveModuleTool is a tool for saving module configuration
type SaveModuleTool struct {
}

// NewSaveModuleTool creates a new save module tool
func NewSaveModuleTool() (*SaveModuleTool, error) {
	return &SaveModuleTool{}, nil
}

// Info returns information about the tool
func (t *SaveModuleTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "saveModule",
		Desc: "Save the configuration of a module",
	}, nil
}

// APIResponse is the standardized API response structure
type APIResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// IsInvokable indicates that this tool can be invoked
func (t *SaveModuleTool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *SaveModuleTool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
	// 获取state
	userReq, ok := ctx.Value(config.StateKey).(*api.ChatRequest)
	if !ok {
		return "", fmt.Errorf("state not found in context")
	}
	log.Printf("Processing LocalTool calls in message: %v", userReq)

	// 获取cache
	cacheKey := cache.CacheKey(userReq.ConversationID)
	info, ok := cache.ModuleCacheInstance.Get(cacheKey)
	if !ok {
		return "", fmt.Errorf("failed to get cache")
	}
	infoCache := info.(*cache.ModuleCacheData)
	log.Printf("Cache key: %v, Module info: %+v", cacheKey, infoCache)

	// Create the request payload
	payload := map[string]interface{}{
		"moduleName": infoCache.ModuleName,
		"moduleCode": infoCache.ModuleCode,
		"configJson": infoCache.Cur,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to create request payload: %w", err)
	}
	log.Printf("payload: %s", string(jsonPayload))
	// Build the request URL
	reqURL := fmt.Sprintf("%s/dynamicForm/config", app.ConfigClient.BaseURL)

	// Create a new request
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set the content type header
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := app.ConfigClient.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the response
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check the response code
	if apiResp.Code != 200 {
		return "", fmt.Errorf("API error: %s", apiResp.Msg)
	}

	return string(body), nil
}
