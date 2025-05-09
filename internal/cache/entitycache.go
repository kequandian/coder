package cache

import (
	"coder/api"
	"coder/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// EntityCacheData represents the cached entity configuration data
type EntityCacheData struct {
	EntityName string    // Entity name
	Attributes string    // JSON string of entity attributes
	Config     string    // JSON string of entity configuration
	CachedAt   time.Time // When the data was cached
}

// NewEntityCacheData creates a new entity cache entry
func NewEntityCacheData(entityName, attributes, config string) *EntityCacheData {
	return &EntityCacheData{
		EntityName: entityName,
		Attributes: attributes,
		Config:     config,
		CachedAt:   time.Now(),
	}
}

func DecodeEntityFromCtx(ctx context.Context) (*EntityCacheData, map[string]interface{}, map[string]interface{}, *api.ChatRequest, error) {
	userReq, ok := ctx.Value(config.StateKey).(*api.ChatRequest)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("state not found in context")
	}
	log.Printf("Processing entity config in message: %v", userReq)

	// Get cache
	cacheKey := CacheKey(userReq.ConversationID)
	info, ok := EntityCacheInstance.Get(cacheKey)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("failed to get cache")
	}

	infoCache := info.(*EntityCacheData)
	attributes := make(map[string]interface{})
	err := json.Unmarshal([]byte(infoCache.Attributes), &attributes)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to unmarshal attributes: %w", err)
	}
	config := make(map[string]interface{})
	err = json.Unmarshal([]byte(infoCache.Config), &config)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return infoCache, attributes, config, userReq, nil
}

// Global cache instance for entities
var EntityCacheInstance = New()
