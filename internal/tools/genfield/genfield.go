package genfield

import (
	"coder/api"
	"coder/internal/cache"
	"coder/internal/config"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// GenFieldTool is a tool for generating form field configurations
type GenFieldTool struct{}

// NewGenFieldTool creates a new gen field tool
func NewGenFieldTool() (*GenFieldTool, error) {
	return &GenFieldTool{}, nil
}

// Info returns information about the tool
func (t *GenFieldTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "genField",
		Desc: "Generate complete form configurations including entity and field attributes",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"entity": {
				Desc:     "Entity information containing entityName, name and note",
				Type:     schema.Object,
				Required: true,
			},
			"attributes": {
				Desc:     "An array of field definitions, each containing attributeName (must be in English), componentType (e.g. input, select), fieldName, fieldType (e.g. string, number), placeholder and required. When componentType is select, must include options field.",
				Type:     schema.Array,
				Required: true,
			},
		}),
	}, nil
}

// IsInvokable indicates that this tool can be invoked
func (t *GenFieldTool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *GenFieldTool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
	// Parse the arguments
	var params struct {
		Entity struct {
			EntityName string `json:"entityName"`
			Name       string `json:"name"`
			Note       string `json:"note"`
		} `json:"entity"`
		Attributes []struct {
			AttributeName string              `json:"attributeName"`
			ComponentType string              `json:"componentType"`
			FieldName     string              `json:"fieldName"`
			FieldType     string              `json:"fieldType"`
			Placeholder   string              `json:"placeholder"`
			Required      bool                `json:"required"`
			Options       []map[string]string `json:"options,omitempty"`
		} `json:"attributes"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}
	fmt.Printf("Parsed arguments: entity=%+v, attributes count=%d\n", params.Entity, len(params.Attributes))

	if len(params.Attributes) == 0 {
		return "", fmt.Errorf("attributes cannot be empty")
	}

	// Generate field configurations
	entityConfig := map[string]interface{}{
		"entity": map[string]interface{}{
			"entityName": params.Entity.EntityName,
			"name":       params.Entity.Name,
			"note":       params.Entity.Note,
		},
		"attributes": make([]map[string]interface{}, 0),
	}
	for _, attribute := range params.Attributes {
		attrMap := map[string]interface{}{
			"attributeName": attribute.AttributeName,
			"componentType": attribute.ComponentType,
			"fieldName":     attribute.FieldName,
			"fieldType":     attribute.FieldType,
			"placeholder":   attribute.Placeholder,
			"required":      attribute.Required,
		}
		if attribute.ComponentType == "select" {
			attrMap["options"] = []map[string]string{
				{"label": "Option 1", "value": "1"},
				{"label": "Option 2", "value": "2"},
			}
		}
		entityConfig["attributes"] = append(entityConfig["attributes"].([]map[string]interface{}), attrMap)
	}

	result, err := json.Marshal(entityConfig)
	if err != nil {
		return "", fmt.Errorf("failed to generate field configurations: %w", err)
	}
	fmt.Printf("Generated field configurations: %s\n", string(result))

	// Save to cache
	userReq, ok := ctx.Value(config.StateKey).(*api.ChatRequest)
	if ok {
		cacheKey := cache.CacheKey(userReq.ConversationID)
		fmt.Printf("Saving to cache with key: %s\n", cacheKey)
		fmt.Printf("Cache content: %+v\n", entityConfig)

		jsonCur, _ := json.Marshal(entityConfig)
		entityName, _ := json.Marshal(params.Entity)
		attributes, _ := json.Marshal(params.Attributes)
		moduleCache := cache.NewEntityCacheData(string(entityName), string(attributes), string(jsonCur))
		cache.EntityCacheInstance.Set(cacheKey, moduleCache, cache.DefaultCacheExpiration)
	}

	return string(result), nil
}
