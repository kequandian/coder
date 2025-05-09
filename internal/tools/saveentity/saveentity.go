package saveentity

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

// SaveEntityTool is a tool for saving entity configuration
type SaveEntityTool struct{}

// NewSaveEntityTool creates a new save entity tool
func NewSaveEntityTool() (*SaveEntityTool, error) {
	return &SaveEntityTool{}, nil
}

// Info returns information about the tool
func (t *SaveEntityTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "saveEntity",
		Desc: "Save the configuration of an entity",
	}, nil
}

// APIResponse is the standardized API response structure
type APIResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// IsInvokable indicates that this tool can be invoked
func (t *SaveEntityTool) IsInvokable() bool {
	return true
}

// InvokableRun runs the tool
func (t *SaveEntityTool) InvokableRun(ctx context.Context, args string, _ ...tool.Option) (string, error) {
	// Get state
	userReq, ok := ctx.Value(config.StateKey).(*api.ChatRequest)
	if !ok {
		return "", fmt.Errorf("state not found in context")
	}
	log.Printf("Processing LocalTool calls in message: %v", userReq)

	// Get cache
	cacheKey := cache.CacheKey(userReq.ConversationID)
	info, ok := cache.EntityCacheInstance.Get(cacheKey)
	if !ok {
		return "", fmt.Errorf("failed to get cache")
	}
	infoCache := info.(*cache.EntityCacheData)

	// Create the request payload
	// Parse entity name info
	var entityNameInfo map[string]interface{}
	if err := json.Unmarshal([]byte(infoCache.EntityName), &entityNameInfo); err != nil {
		return "", fmt.Errorf("failed to parse entity name info: %w", err)
	}

	// Parse attributes
	var attributes []map[string]interface{}
	if err := json.Unmarshal([]byte(infoCache.Attributes), &attributes); err != nil {
		return "", fmt.Errorf("failed to parse attributes: %w", err)
	}

	// Convert attributes to createFields format

	createFields := make([]map[string]interface{}, 0)
	for _, attr := range attributes {
		field := map[string]interface{}{
			"label": attr["fieldName"],
			"field": attr["attributeName"],
			"type":  attr["componentType"],
			"props": map[string]interface{}{
				"placeholder": attr["placeholder"],
			},
			"options": attr["options"],
		}
		if attr["required"] == true {
			field["rules"] = []map[string]interface{}{{"type": "required"}}
		}
		createFields = append(createFields, field)
	}

	tableFields := make([]map[string]interface{}, 0)
	for _, attr := range attributes {
		field := map[string]interface{}{
			"label": attr["fieldName"],
			"field": attr["attributeName"],
			// "type":  attr["componentType"],
			// "props": map[string]interface{}{
			// 	"placeholder": attr["placeholder"],
			// },
		}
		// if attr["required"] == true {
		// 	field["rules"] = []map[string]interface{}{{"type": "required"}}
		// }
		tableFields = append(tableFields, field)
	}

	// Create the request payload with new structure
	// Generate viewConfig fields
	viewFields := make([]map[string]interface{}, 0)
	for _, attr := range attributes {
		field := map[string]interface{}{
			"field": attr["attributeName"],
			"label": attr["fieldName"],
			"type":  "plain",
		}
		if attr["componentType"] == "select" || attr["componentType"] == "radio" {
			field["options"] = map[string]interface{}{
				"map": attr["options"],
				"chy": attr["options"],
			}
		}
		viewFields = append(viewFields, field)
	}

	payload := map[string]interface{}{
		"entityName": infoCache.EntityName,
		"entityConfig": map[string]interface{}{
			"pageName": map[string]interface{}{
				"table": "字段模板",
				"new":   "新增字段模板",
				"edit":  "更改字段模板",
				"name":  "",
			},
			"createFields": createFields,
			"updateFields": createFields,
			"tableFields":  tableFields,
			"viewConfig":   viewFields,
			"map":          map[string]interface{}{},
			"listAPI":      fmt.Sprintf("/api/adm/data/services/%s", entityNameInfo["entityName"]),
			"createAPI":    fmt.Sprintf("/api/adm/data/services/%s", entityNameInfo["entityName"]),
			"getAPI":       fmt.Sprintf("/api/adm/data/services/%s/[id]", entityNameInfo["entityName"]),
			"updateAPI":    fmt.Sprintf("/api/adm/data/services/%s/[id]", entityNameInfo["entityName"]),
			"deleteAPI":    fmt.Sprintf("/api/adm/data/services/%s/(id)", entityNameInfo["entityName"]),
			"columns":      1,
			"layout": map[string]string{
				"table": "Content",
				"form":  "TitleContent",
			},
			"tableActions": []map[string]interface{}{
				{
					"title": "添加",
					"type":  "path",
					"options": map[string]interface{}{
						"style": "primary",
						"path":  "checkfiles/checkfiles-add",
					},
				},
			},
			"tableOperation": []map[string]interface{}{
				{
					"title": "详情",
					"type":  "path",
					"options": map[string]interface{}{
						"outside": true,
						"path":    "checkfiles/checkfiles-view",
					},
				},
				{
					"title": "编辑",
					"type":  "path",
					"options": map[string]interface{}{
						"outside": true,
						"path":    "checkfiles/checkfiles-edit",
					},
				},
				{
					"title": "删除",
					"type":  "delete",
				},
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to create request payload: %w", err)
	}
	log.Printf("payload: %s", string(jsonPayload))

	// 1. 将结构体序列化为 JSON 字节
	// entityJSON, err := json.Marshal(infoCache.EntityName)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to marshal entity data: %v", err)
	// }

	entityJSON := []byte(infoCache.EntityName)

	// 2. 创建 Reader 作为 Body
	body := bytes.NewReader(entityJSON)

	// 3. 创建请求并设置 JSON 头
	log.Printf("Sending entity data to %s/api/adm/cfg/entities with payload: %s", app.ConfigClient.BaseURL, string(entityJSON))
	reqURL1 := fmt.Sprintf("%s/api/adm/cfg/entities", app.ConfigClient.BaseURL)
	req1, err := http.NewRequestWithContext(ctx, "POST", reqURL1, body)
	req1.Header.Set("Content-Type", "application/json") // 必须设置 Content-Type!

	if err != nil {
		return "", fmt.Errorf("failed to create entity request: %w", err)
	}
	req1.Header.Set("Content-Type", "application/json")
	resp1, err := app.ConfigClient.Client.Do(req1)
	if err != nil {
		return "", fmt.Errorf("failed to send entity request: %w", err)
	}
	defer resp1.Body.Close()
	body1, err := io.ReadAll(resp1.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read entity response: %w", err)
	}

	// 发送到字段添加API
	log.Printf("Sending field data to %s/api/adm/cfg/attribute/%s/list  data:%s", app.ConfigClient.BaseURL, entityNameInfo["entityName"], infoCache.Attributes)
	reqURL2 := fmt.Sprintf("%s/api/adm/cfg/attribute/%s/list", app.ConfigClient.BaseURL, entityNameInfo["entityName"])
	req2, err := http.NewRequestWithContext(ctx, "POST", reqURL2, bytes.NewBuffer([]byte(infoCache.Attributes)))
	if err != nil {
		return "", fmt.Errorf("failed to create field request: %w", err)
	}
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := app.ConfigClient.Client.Do(req2)
	if err != nil {
		return "", fmt.Errorf("failed to send field request: %w", err)
	}
	defer resp2.Body.Close()
	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read field response: %w", err)
	}

	log.Printf("filedS: %s", body2)

	// 发送到dynamicForm API
	dynamicFormPayload := map[string]interface{}{
		"moduleName": entityNameInfo["name"],
		"moduleCode": entityNameInfo["entityName"],
		"configJson": payload["entityConfig"],
	}
	jsonDynamicForm, err := json.Marshal(dynamicFormPayload)
	if err != nil {
		return "", fmt.Errorf("failed to create dynamicForm payload: %w", err)
	}
	reqURL3 := fmt.Sprintf("%s/dynamicForm/config", app.ConfigClient.BaseURL)
	req3, err := http.NewRequestWithContext(ctx, "POST", reqURL3, bytes.NewBuffer(jsonDynamicForm))
	if err != nil {
		return "", fmt.Errorf("failed to create dynamicForm request: %w", err)
	}
	req3.Header.Set("Content-Type", "application/json")
	resp3, err := app.ConfigClient.Client.Do(req3)
	if err != nil {
		return "", fmt.Errorf("failed to send dynamicForm request: %w", err)
	}
	defer resp3.Body.Close()
	body3, err := io.ReadAll(resp3.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read dynamicForm response: %w", err)
	}
	log.Printf("filedS: %s", body3)

	// Store in cache
	configJsonStr, err := json.Marshal(dynamicFormPayload["configJson"])
	if err != nil {
		return "", fmt.Errorf("failed to marshal configJson: %w", err)
	}
	moduleCache := cache.NewModuleCacheData(
		fmt.Sprint(dynamicFormPayload["moduleName"]),
		fmt.Sprint(dynamicFormPayload["moduleCode"]),
		string(configJsonStr),
		string(configJsonStr),
	)
	cache.ModuleCacheInstance.Set(cacheKey, moduleCache, cache.DefaultCacheExpiration)

	fmt.Println("Cached module", dynamicFormPayload["moduleName"], dynamicFormPayload["moduleCode"], "with key", cacheKey)
	return string(body1), nil
}
