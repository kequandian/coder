package viewmodule

import (
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
	"net/url"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ViewModuleTool is a tool for getting module configuration
type ViewModuleTool struct {
}

// NewViewModuleTool creates a new view module tool
func NewViewModuleTool() (*ViewModuleTool, error) {
	return &ViewModuleTool{}, nil
}

// Info returns information about the tool
func (t *ViewModuleTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "viewModule",
		Desc: "Get/Load the configuration of a module by module name and module code",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"module_name": {
				Desc:     "The name of the module chinese name",
				Type:     schema.String,
				Required: true,
			},
			"module_code": {
				Desc:     "The code of the module code, little english from the module name",
				Type:     schema.String,
				Required: true,
			},
		}),
	}, nil
}

// APIResponse is the standardized API response structure
type APIResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// ModuleConfigData is the data structure for GetModuleConfig response
type ModuleConfigData struct {
	Support string `json:"support"`
	Cur     string `json:"cur"`
}

// IsInvokable indicates that this tool can be invoked
func (t *ViewModuleTool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *ViewModuleTool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
	// Parse the arguments
	var params struct {
		ModuleName string `json:"module_name"`
		ModuleCode string `json:"module_code"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	if params.ModuleName == "" || params.ModuleCode == "" {
		return "", fmt.Errorf("module_name and module_code cannot be empty")
	}

	// 获取state
	userReq, ok := ctx.Value(config.StateKey).(*api.ChatRequest)
	if !ok {
		return "", fmt.Errorf("state not found in context")
	}
	log.Printf("Processing LocalTool calls in message: %v", userReq)

	cacheKey := cache.CacheKey(userReq.ConversationID)
	// If not in cache, fetch from API
	fmt.Println("Cache miss for module", params.ModuleName, params.ModuleCode, "fetching from API")

	// Build the request URL with query parameters
	log.Printf("app.ConfigClient.BaseURL: %v", app.ConfigClient.BaseURL)
	reqURL := fmt.Sprintf("%s/dynamicForm/config", app.ConfigClient.BaseURL)

	// Create a URL with query parameters
	baseURL, err := url.Parse(reqURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	log.Printf("baseURL: %v", baseURL)
	// Add query parameters
	query := baseURL.Query()
	query.Add("moduleName", params.ModuleName)
	query.Add("moduleCode", params.ModuleCode)
	baseURL.RawQuery = query.Encode()
	log.Printf("baseURL.String(): %v", baseURL.String())
	// Create a new request
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

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
	apiResp.Data = &ModuleConfigData{}

	fmt.Println("body", string(body))
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check the response code
	if apiResp.Code != 200 {
		return "", fmt.Errorf("API error: %s", apiResp.Msg)
	}

	// Convert the data to the expected structure
	respData, ok := apiResp.Data.(*ModuleConfigData)
	if !ok {
		return "", fmt.Errorf("unexpected response data format")
	}

	cur := respData.Cur
	if cur == "" {
		cur = respData.Support
	}

	if cur == "" {
		return "", fmt.Errorf("module config is empty")
	}

	// Store in cache
	moduleCache := cache.NewModuleCacheData(params.ModuleName, params.ModuleCode, respData.Support, cur)
	cache.ModuleCacheInstance.Set(cacheKey, moduleCache, cache.DefaultCacheExpiration)

	fmt.Println("Cached module", params.ModuleName, params.ModuleCode, "with key", cacheKey)

	return cur, nil
}
