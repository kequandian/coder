package gencode

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ModuleField represents a field in a module
type ModuleField struct {
	NameCN   string `json:"name_cn"`           // Chinese name of the field
	NameEN   string `json:"name_en"`           // English name of the field
	Type     string `json:"type"`              // SQL type (VARCHAR, INT, etc.)
	Required bool   `json:"required"`          // Is the field required
	Comment  string `json:"comment,omitempty"` // Additional comments about the field
}

// ModuleDefinition represents a module definition
type ModuleDefinition struct {
	ModuleName string        `json:"module_name"` // Module name
	Fields     []ModuleField `json:"fields"`      // Module fields
}

// GenCodeTool is a tool for generating code for business modules
type GenCodeTool struct {
	apiKey       string
	baseURL      string
	modelID      string
	chatModel    model.ChatModel
	systemPrompt string
}

// NewGenCodeTool creates a new code generation tool
func NewGenCodeTool(ctx context.Context, apiKey, baseURL, modelID string) (*GenCodeTool, error) {
	// Create the tool
	tool := &GenCodeTool{
		apiKey:  apiKey,
		baseURL: baseURL,
		modelID: modelID,
		systemPrompt: `You are a database schema expert who helps design database tables for business applications.
When given a module name, you need to infer what fields would be appropriate for that module and return them in a structured format.
Each field should have:
1. A Chinese name (name_cn)
2. An English name (name_en) - this should be a good variable name in camelCase or snake_case
3. An SQL data type (type) - use standard SQL types like VARCHAR(255), INT, DATETIME, TEXT, etc.
4. Whether it's required (required) - true/false
5. Optional comment for additional information

For common modules, include standard fields like:
- id fields (usually INT or BIGINT, PRIMARY KEY)
- creation and modification timestamps (created_at, updated_at as DATETIME)
- status fields when appropriate (status as TINYINT or VARCHAR)
- relationship fields where appropriate (e.g., user_id, product_id)

Example module types might include:
- User management (用户管理)
- Order processing (订单处理)
- Product management (产品管理)
- Task tracking (任务追踪)
- Permission systems (权限系统)
- Content management (内容管理)

Return your response as a JSON object containing a module_name field and a fields array with the properties described above.`,
	}

	// Configure and create the model
	modelConfig := &openai.ChatModelConfig{
		Model:  modelID,
		APIKey: apiKey,
	}

	if baseURL != "" {
		modelConfig.BaseURL = baseURL
	}

	chatModel, err := openai.NewChatModel(ctx, modelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model for code generation tool: %w", err)
	}

	tool.chatModel = chatModel
	return tool, nil
}

// Info returns information about the tool
func (t *GenCodeTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "generateModuleSchema",
		Desc: "Generates database schema definition for a business module based on the module name. Returns field definitions with Chinese and English names and SQL types.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"module_name": {
				Desc:     "The name of the business module to generate schema for",
				Type:     schema.String,
				Required: true,
			},
		}),
	}, nil
}

// IsInvokable indicates that this tool can be invoked
func (t *GenCodeTool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *GenCodeTool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
	// Parse the arguments
	var params struct {
		ModuleName string `json:"module_name"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	if params.ModuleName == "" {
		return "", fmt.Errorf("module_name cannot be empty")
	}

	// Create messages for the model
	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: t.systemPrompt,
		},
		{
			Role:    schema.User,
			Content: fmt.Sprintf("Generate a database schema for a %s module. Return only the JSON, no explanations.", params.ModuleName),
		},
	}

	// Generate the module schema
	startTime := time.Now()
	log.Printf("Generating module schema for %s", params.ModuleName)
	msg, err := t.chatModel.Generate(ctx, messages)
	log.Printf("Generated module schema for %s: %s cost:%d", params.ModuleName, msg.Content, time.Since(startTime))
	if err != nil {
		return "", fmt.Errorf("failed to generate module schema: %w", err)
	}

	// Extract JSON content from the response
	content := msg.Content

	// Check if the response contains a code block
	if strings.Contains(content, "```json") {
		// Extract content between JSON code blocks
		parts := strings.Split(content, "```json")
		if len(parts) > 1 {
			jsonPart := strings.Split(parts[1], "```")[0]
			content = strings.TrimSpace(jsonPart)
		}
	} else if strings.Contains(content, "```") {
		// Try to extract from regular code blocks
		parts := strings.Split(content, "```")
		if len(parts) > 1 {
			content = strings.TrimSpace(parts[1])
		}
	}

	// Clean up the content and validate it's proper JSON
	var moduleDefinition ModuleDefinition
	if err := json.Unmarshal([]byte(content), &moduleDefinition); err != nil {
		log.Printf("Error parsing generated schema: %v. Raw content: %s", err, content)
		return "", fmt.Errorf("generated schema is not valid JSON: %w", err)
	}

	// Ensure the module name is set
	if moduleDefinition.ModuleName == "" {
		moduleDefinition.ModuleName = params.ModuleName
	}

	// Return the formatted result
	result, err := json.MarshalIndent(moduleDefinition, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format module schema: %w", err)
	}

	return string(result), nil
}
