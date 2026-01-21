/*
 * Copyright 1999-2020 Alibaba Group Holding Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ai_cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
)

// McpServerQueryFunc is the function type for querying MCP server
type McpServerQueryFunc func(mcpName, version string) (*model.McpServerDetailInfo, error)

// McpServerCacheHolder manages the cache and periodic updates for MCP servers
type McpServerCacheHolder struct {
	mu               sync.RWMutex
	cache            map[string]*model.McpServerDetailInfo
	updateTasks      map[string]context.CancelFunc
	subscribeManager *McpSubscribeManager
	queryFunc        McpServerQueryFunc
	namespaceID      string
}

// NewMcpServerCacheHolder creates a new McpServerCacheHolder
func NewMcpServerCacheHolder(subscribeManager *McpSubscribeManager, queryFunc McpServerQueryFunc, namespaceID string) *McpServerCacheHolder {
	return &McpServerCacheHolder{
		cache:            make(map[string]*model.McpServerDetailInfo),
		updateTasks:      make(map[string]context.CancelFunc),
		subscribeManager: subscribeManager,
		queryFunc:        queryFunc,
		namespaceID:      namespaceID,
	}
}

// buildKey builds the cache key
func (h *McpServerCacheHolder) buildKey(mcpName, version string) string {
	if version == "" {
		version = "latest"
	}
	return mcpName + "::" + version
}

// GetMcpServer gets MCP server info from cache
func (h *McpServerCacheHolder) GetMcpServer(mcpName, version string) *model.McpServerDetailInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()

	key := h.buildKey(mcpName, version)
	return h.cache[key]
}

// SetMcpServer sets MCP server info to cache
func (h *McpServerCacheHolder) SetMcpServer(mcpName, version string, info *model.McpServerDetailInfo) {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := h.buildKey(mcpName, version)
	h.cache[key] = info
}

// RemoveMcpServer removes MCP server info from cache
func (h *McpServerCacheHolder) RemoveMcpServer(mcpName, version string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := h.buildKey(mcpName, version)
	delete(h.cache, key)
}

// StartUpdateTask starts a background task to periodically update the cache
func (h *McpServerCacheHolder) StartUpdateTask(mcpName, version string) {
	h.mu.Lock()
	key := h.buildKey(mcpName, version)

	// Check if task already exists
	if _, exists := h.updateTasks[key]; exists {
		h.mu.Unlock()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	h.updateTasks[key] = cancel
	h.mu.Unlock()

	go h.updateLoop(ctx, mcpName, version)
}

// StopUpdateTask stops the background update task
func (h *McpServerCacheHolder) StopUpdateTask(mcpName, version string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := h.buildKey(mcpName, version)
	if cancel, exists := h.updateTasks[key]; exists {
		cancel()
		delete(h.updateTasks, key)
	}
}

// updateLoop is the main update loop
func (h *McpServerCacheHolder) updateLoop(ctx context.Context, mcpName, version string) {
	ticker := time.NewTicker(time.Duration(constant.AI_SUBSCRIPTION_UPDATE_INTERVAL) * time.Second)
	defer ticker.Stop()

	for {
		// Execute update first, then wait (same as Python SDK)
		h.updateMcpServer(mcpName, version)

		select {
		case <-ctx.Done():
			logger.Infof("MCP server update task stopped, mcpName=%s, version=%s", mcpName, version)
			return
		case <-ticker.C:
			// Continue to next iteration
		}
	}
}

// updateMcpServer fetches the latest MCP server info and notifies subscribers if changed
func (h *McpServerCacheHolder) updateMcpServer(mcpName, version string) {
	if h.queryFunc == nil {
		return
	}

	newInfo, err := h.queryFunc(mcpName, version)
	if err != nil {
		logger.Errorf("Failed to update MCP server info, mcpName=%s, version=%s, error=%v", mcpName, version, err)
		return
	}

	if newInfo == nil {
		return
	}

	// Check if changed
	oldInfo := h.GetMcpServer(mcpName, version)
	if h.isEqual(oldInfo, newInfo) {
		return
	}

	// Update cache
	h.SetMcpServer(mcpName, version, newInfo)

	// Notify subscribers
	mcpID := ""
	if newInfo != nil {
		mcpID = newInfo.ID
	}
	h.subscribeManager.NotifySubscribers(mcpID, h.namespaceID, mcpName, version, *newInfo)
	logger.Infof("MCP server info changed, mcpName=%s, version=%s", mcpName, version)
}

// isEqual compares two McpServerDetailInfo for equality
func (h *McpServerCacheHolder) isEqual(a, b *model.McpServerDetailInfo) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Compare by JSON serialization for simplicity
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}

// Close stops all update tasks
func (h *McpServerCacheHolder) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for key, cancel := range h.updateTasks {
		cancel()
		delete(h.updateTasks, key)
	}
}
