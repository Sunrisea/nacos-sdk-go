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

// AgentCardQueryFunc is the function type for querying Agent card
type AgentCardQueryFunc func(agentName, version, registrationType string) (*model.AgentCardDetailInfo, error)

// AgentCacheHolder manages the cache and periodic updates for Agent cards
type AgentCacheHolder struct {
	mu               sync.RWMutex
	cache            map[string]*model.AgentCardDetailInfo
	updateTasks      map[string]context.CancelFunc
	subscribeManager *AgentSubscribeManager
	queryFunc        AgentCardQueryFunc
	namespaceID      string
}

// NewAgentCacheHolder creates a new AgentCacheHolder
func NewAgentCacheHolder(subscribeManager *AgentSubscribeManager, queryFunc AgentCardQueryFunc, namespaceID string) *AgentCacheHolder {
	return &AgentCacheHolder{
		cache:            make(map[string]*model.AgentCardDetailInfo),
		updateTasks:      make(map[string]context.CancelFunc),
		subscribeManager: subscribeManager,
		queryFunc:        queryFunc,
		namespaceID:      namespaceID,
	}
}

// buildKey builds the cache key
func (h *AgentCacheHolder) buildKey(agentName, version string) string {
	if version == "" {
		version = "latest"
	}
	return agentName + "::" + version
}

// GetAgentCard gets Agent card info from cache
func (h *AgentCacheHolder) GetAgentCard(agentName, version string) *model.AgentCardDetailInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()

	key := h.buildKey(agentName, version)
	return h.cache[key]
}

// SetAgentCard sets Agent card info to cache
func (h *AgentCacheHolder) SetAgentCard(agentName, version string, info *model.AgentCardDetailInfo) {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := h.buildKey(agentName, version)
	h.cache[key] = info
}

// RemoveAgentCard removes Agent card info from cache
func (h *AgentCacheHolder) RemoveAgentCard(agentName, version string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := h.buildKey(agentName, version)
	delete(h.cache, key)
}

// StartUpdateTask starts a background task to periodically update the cache
func (h *AgentCacheHolder) StartUpdateTask(agentName, version, registrationType string) {
	h.mu.Lock()
	key := h.buildKey(agentName, version)

	// Check if task already exists
	if _, exists := h.updateTasks[key]; exists {
		h.mu.Unlock()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	h.updateTasks[key] = cancel
	h.mu.Unlock()

	go h.updateLoop(ctx, agentName, version, registrationType)
}

// StopUpdateTask stops the background update task
func (h *AgentCacheHolder) StopUpdateTask(agentName, version string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := h.buildKey(agentName, version)
	if cancel, exists := h.updateTasks[key]; exists {
		cancel()
		delete(h.updateTasks, key)
	}
}

// updateLoop is the main update loop
func (h *AgentCacheHolder) updateLoop(ctx context.Context, agentName, version, registrationType string) {
	ticker := time.NewTicker(time.Duration(constant.AI_SUBSCRIPTION_UPDATE_INTERVAL) * time.Second)
	defer ticker.Stop()

	for {
		// Execute update first, then wait (same as Python SDK)
		h.updateAgentCard(agentName, version, registrationType)

		select {
		case <-ctx.Done():
			logger.Infof("Agent card update task stopped, agentName=%s, version=%s", agentName, version)
			return
		case <-ticker.C:
			// Continue to next iteration
		}
	}
}

// updateAgentCard fetches the latest Agent card info and notifies subscribers if changed
func (h *AgentCacheHolder) updateAgentCard(agentName, version, registrationType string) {
	if h.queryFunc == nil {
		return
	}

	newInfo, err := h.queryFunc(agentName, version, registrationType)
	if err != nil {
		logger.Errorf("Failed to update Agent card info, agentName=%s, version=%s, error=%v", agentName, version, err)
		return
	}

	if newInfo == nil {
		return
	}

	// Check if changed
	oldInfo := h.GetAgentCard(agentName, version)

	if h.isEqual(oldInfo, newInfo) {
		return
	}

	// Update cache
	h.SetAgentCard(agentName, version, newInfo)

	// Notify subscribers
	h.subscribeManager.NotifySubscribers(agentName, version, *newInfo)
	logger.Infof("Agent card info changed, agentName=%s, version=%s", agentName, version)
}

// isEqual compares two AgentCardDetailInfo for equality
func (h *AgentCacheHolder) isEqual(a, b *model.AgentCardDetailInfo) bool {
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
func (h *AgentCacheHolder) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for key, cancel := range h.updateTasks {
		cancel()
		delete(h.updateTasks, key)
	}
}
