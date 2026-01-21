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

// AgentSubscribeManager manages Agent card subscriptions
type AgentSubscribeManager struct {
	mu          sync.RWMutex
	subscribers map[string][]vo.AgentCardChangeCallback
}

// NewAgentSubscribeManager creates a new AgentSubscribeManager
func NewAgentSubscribeManager() *AgentSubscribeManager {
	return &AgentSubscribeManager{
		subscribers: make(map[string][]vo.AgentCardChangeCallback),
	}
}

// buildKey builds the subscription key
func (m *AgentSubscribeManager) buildKey(agentName, version string) string {
	if version == "" {
		version = "latest"
	}
	return agentName + "::" + version
}

// RegisterSubscriber registers a subscriber callback
func (m *AgentSubscribeManager) RegisterSubscriber(agentName, version string, callback vo.AgentCardChangeCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.buildKey(agentName, version)
	if _, exists := m.subscribers[key]; !exists {
		m.subscribers[key] = make([]vo.AgentCardChangeCallback, 0)
	}
	m.subscribers[key] = append(m.subscribers[key], callback)
}

// DeregisterSubscriber removes a subscriber callback
func (m *AgentSubscribeManager) DeregisterSubscriber(agentName, version string, callback vo.AgentCardChangeCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.buildKey(agentName, version)
	callbacks, exists := m.subscribers[key]
	if !exists {
		return
	}

	// Note: Go doesn't support direct function comparison, so we remove all callbacks for now
	delete(m.subscribers, key)
	_ = callbacks // suppress unused warning
}

// IsSubscribed checks if there are subscribers for the given Agent
func (m *AgentSubscribeManager) IsSubscribed(agentName, version string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := m.buildKey(agentName, version)
	callbacks, exists := m.subscribers[key]
	return exists && len(callbacks) > 0
}

// GetSubscribers returns all subscribers for the given Agent
func (m *AgentSubscribeManager) GetSubscribers(agentName, version string) []vo.AgentCardChangeCallback {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := m.buildKey(agentName, version)
	callbacks, exists := m.subscribers[key]
	if !exists {
		return nil
	}

	// Return a copy to avoid race conditions
	result := make([]vo.AgentCardChangeCallback, len(callbacks))
	copy(result, callbacks)
	return result
}

// NotifySubscribers notifies all subscribers with the updated Agent card info
func (m *AgentSubscribeManager) NotifySubscribers(agentName, version string, agentCard model.AgentCardDetailInfo) {
	callbacks := m.GetSubscribers(agentName, version)
	for _, callback := range callbacks {
		if callback != nil {
			go callback(agentName, agentCard)
		}
	}
}

// GetAllSubscribedKeys returns all subscribed keys
func (m *AgentSubscribeManager) GetAllSubscribedKeys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.subscribers))
	for key := range m.subscribers {
		keys = append(keys, key)
	}
	return keys
}
