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

// DefaultCacheExpiration is the default expiration time for cached module data
const DefaultCacheExpiration = 1 * time.Hour

// ModuleCacheData represents the cached module data
type ModuleCacheData struct {
	ModuleName string    // Module name
	ModuleCode string    // Module code
	Support    string    // Module configuration
	Cur        string    // Module configuration
	CachedAt   time.Time // When the data was cached
}

// NewModuleCacheData creates a new module cache entry
func NewModuleCacheData(moduleName, moduleCode, support, cur string) *ModuleCacheData {
	return &ModuleCacheData{
		ModuleName: moduleName,
		ModuleCode: moduleCode,
		Support:    support,
		Cur:        cur,
		CachedAt:   time.Now(),
	}
}

func DecodeModuleFromCtx(ctx context.Context) (*ModuleCacheData, map[string]interface{}, map[string]interface{}, *api.ChatRequest, error) {
	// In a real implementation, we would edit the action in the module
	userReq, ok := ctx.Value(config.StateKey).(*api.ChatRequest)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("state not found in context")
	}
	log.Printf("Processing LocalTool calls in message: %v", userReq)

	// Get cache
	cacheKey := CacheKey(userReq.ConversationID)
	info, ok := ModuleCacheInstance.Get(cacheKey)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("failed to get cache")
	}

	infoCache := info.(*ModuleCacheData)
	mapCur := make(map[string]interface{})
	err := json.Unmarshal([]byte(infoCache.Cur), &mapCur)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to unmarshal cur: %w", err)
	}
	mapSupport := make(map[string]interface{})
	err = json.Unmarshal([]byte(infoCache.Support), &mapSupport)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to unmarshal support: %w", err)
	}
	return infoCache, mapCur, mapSupport, userReq, nil
}

// Global cache instance for modules
var ModuleCacheInstance = New()

// CacheKey generates a cache key for a module
func CacheKey(sessionID string) string {
	return sessionID
}
