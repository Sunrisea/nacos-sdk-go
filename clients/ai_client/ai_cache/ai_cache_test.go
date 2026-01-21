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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func TestMcpSubscribeManager(t *testing.T) {
	manager := NewMcpSubscribeManager()

	callback1 := func(mcpId, namespaceId, mcpName string, mcpServer model.McpServerDetailInfo) {}
	callback2 := func(mcpId, namespaceId, mcpName string, mcpServer model.McpServerDetailInfo) {}

	// Test register subscriber
	manager.RegisterSubscriber("test-mcp", "v1", callback1)
	assert.True(t, manager.IsSubscribed("test-mcp", "v1"))
	assert.False(t, manager.IsSubscribed("test-mcp", "v2"))

	// Test multiple callbacks
	manager.RegisterSubscriber("test-mcp", "v1", callback2)
	subscribers := manager.GetSubscribers("test-mcp", "v1")
	assert.Equal(t, 2, len(subscribers))

	// Test deregister (removes all callbacks for the key)
	manager.DeregisterSubscriber("test-mcp", "v1", callback1)
	assert.False(t, manager.IsSubscribed("test-mcp", "v1"))

	// Test get all subscribed keys
	manager.RegisterSubscriber("mcp-1", "v1", callback1)
	manager.RegisterSubscriber("mcp-2", "v2", callback1)
	keys := manager.GetAllSubscribedKeys()
	assert.Equal(t, 2, len(keys))
}

func TestAgentSubscribeManager(t *testing.T) {
	manager := NewAgentSubscribeManager()

	callback1 := func(agentName string, agentCard model.AgentCardDetailInfo) {}
	callback2 := func(agentName string, agentCard model.AgentCardDetailInfo) {}

	// Test register subscriber
	manager.RegisterSubscriber("test-agent", "v1", callback1)
	assert.True(t, manager.IsSubscribed("test-agent", "v1"))
	assert.False(t, manager.IsSubscribed("test-agent", "v2"))

	// Test multiple callbacks
	manager.RegisterSubscriber("test-agent", "v1", callback2)
	subscribers := manager.GetSubscribers("test-agent", "v1")
	assert.Equal(t, 2, len(subscribers))

	// Test deregister (removes all callbacks for the key)
	manager.DeregisterSubscriber("test-agent", "v1", callback1)
	assert.False(t, manager.IsSubscribed("test-agent", "v1"))
}

func TestMcpServerCacheHolder(t *testing.T) {
	subscribeManager := NewMcpSubscribeManager()
	queryFunc := func(mcpName, version string) (*model.McpServerDetailInfo, error) {
		return &model.McpServerDetailInfo{
			McpServerBasicInfo: model.McpServerBasicInfo{
				Name:    mcpName,
				Version: version,
			},
		}, nil
	}

	cacheHolder := NewMcpServerCacheHolder(subscribeManager, queryFunc, "test-ns")

	// Test set and get
	mcpServer := &model.McpServerDetailInfo{
		McpServerBasicInfo: model.McpServerBasicInfo{
			Name:    "test-mcp",
			Version: "v1",
		},
	}
	cacheHolder.SetMcpServer("test-mcp", "v1", mcpServer)

	cached := cacheHolder.GetMcpServer("test-mcp", "v1")
	assert.NotNil(t, cached)
	assert.Equal(t, "test-mcp", cached.Name)

	// Test remove
	cacheHolder.RemoveMcpServer("test-mcp", "v1")
	cached = cacheHolder.GetMcpServer("test-mcp", "v1")
	assert.Nil(t, cached)

	// Test close
	cacheHolder.Close()
}

func TestAgentCacheHolder(t *testing.T) {
	subscribeManager := NewAgentSubscribeManager()
	queryFunc := func(agentName, version, registrationType string) (*model.AgentCardDetailInfo, error) {
		return &model.AgentCardDetailInfo{
			LatestVersion: true,
		}, nil
	}

	cacheHolder := NewAgentCacheHolder(subscribeManager, queryFunc, "test-ns")

	// Test set and get
	agentCard := &model.AgentCardDetailInfo{
		LatestVersion: true,
	}
	cacheHolder.SetAgentCard("test-agent", "v1", agentCard)

	cached := cacheHolder.GetAgentCard("test-agent", "v1")
	assert.NotNil(t, cached)
	assert.True(t, cached.LatestVersion)

	// Test remove
	cacheHolder.RemoveAgentCard("test-agent", "v1")
	cached = cacheHolder.GetAgentCard("test-agent", "v1")
	assert.Nil(t, cached)

	// Test close
	cacheHolder.Close()
}

func TestBuildKey(t *testing.T) {
	manager := NewMcpSubscribeManager()

	// Register with empty version should use "latest"
	callback := func(mcpId, namespaceId, mcpName string, mcpServer model.McpServerDetailInfo) {}
	manager.RegisterSubscriber("test-mcp", "", callback)

	// Should be subscribed with "latest" version
	assert.True(t, manager.IsSubscribed("test-mcp", ""))
	assert.True(t, manager.IsSubscribed("test-mcp", "latest"))
}

func TestSubscribeCallbackTypes(t *testing.T) {
	// Test that callback types are correctly defined
	var mcpCallback vo.McpServerChangeCallback = func(mcpId, namespaceId, mcpName string, mcpServer model.McpServerDetailInfo) {
		// Callback implementation
	}
	assert.NotNil(t, mcpCallback)

	var agentCallback vo.AgentCardChangeCallback = func(agentName string, agentCard model.AgentCardDetailInfo) {
		// Callback implementation
	}
	assert.NotNil(t, agentCallback)
}
