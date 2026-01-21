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
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// McpSubscribeManager manages MCP server subscriptions
type McpSubscribeManager struct {
	mu          sync.RWMutex
	subscribers map[string][]vo.McpServerChangeCallback
}

// NewMcpSubscribeManager creates a new McpSubscribeManager
func NewMcpSubscribeManager() *McpSubscribeManager {
	return &McpSubscribeManager{
		subscribers: make(map[string][]vo.McpServerChangeCallback),
	}
}

// buildKey builds the subscription key
func (m *McpSubscribeManager) buildKey(mcpName, version string) string {
	if version == "" {
		version = "latest"
	}
	return mcpName + "::" + version
}

// RegisterSubscriber registers a subscriber callback
func (m *McpSubscribeManager) RegisterSubscriber(mcpName, version string, callback vo.McpServerChangeCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.buildKey(mcpName, version)
	if _, exists := m.subscribers[key]; !exists {
		m.subscribers[key] = make([]vo.McpServerChangeCallback, 0)
	}
	m.subscribers[key] = append(m.subscribers[key], callback)
}

// DeregisterSubscriber removes all subscriber callbacks for the given MCP server
// Note: Go doesn't support direct function comparison, so this removes all callbacks for the key
func (m *McpSubscribeManager) DeregisterSubscriber(mcpName, version string, callback vo.McpServerChangeCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.buildKey(mcpName, version)
	delete(m.subscribers, key)
}

// IsSubscribed checks if there are subscribers for the given MCP server
func (m *McpSubscribeManager) IsSubscribed(mcpName, version string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := m.buildKey(mcpName, version)
	callbacks, exists := m.subscribers[key]
	return exists && len(callbacks) > 0
}

// GetSubscribers returns all subscribers for the given MCP server
func (m *McpSubscribeManager) GetSubscribers(mcpName, version string) []vo.McpServerChangeCallback {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := m.buildKey(mcpName, version)
	callbacks, exists := m.subscribers[key]
	if !exists {
		return nil
	}

	// Return a copy to avoid race conditions
	result := make([]vo.McpServerChangeCallback, len(callbacks))
	copy(result, callbacks)
	return result
}

// NotifySubscribers notifies all subscribers with the updated MCP server info
func (m *McpSubscribeManager) NotifySubscribers(mcpID, namespaceID, mcpName, version string, mcpServer model.McpServerDetailInfo) {
	callbacks := m.GetSubscribers(mcpName, version)
	for _, callback := range callbacks {
		if callback != nil {
			go callback(mcpID, namespaceID, mcpName, mcpServer)
		}
	}
}

// GetAllSubscribedKeys returns all subscribed keys
func (m *McpSubscribeManager) GetAllSubscribedKeys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.subscribers))
	for key := range m.subscribers {
		keys = append(keys, key)
	}
	return keys
}
