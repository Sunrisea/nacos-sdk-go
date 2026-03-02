/*
 * Copyright 1999-2025 Alibaba Group Holding Ltd.
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

package ai

import (
	"github.com/nacos-group/nacos-sdk-go/v2/clients/maintainer_client/core"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// IAiMaintainerClient defines AI module admin operations (MCP + A2A).
// It embeds ICoreMaintainerClient to inherit core admin capabilities.
type IAiMaintainerClient interface {
	core.ICoreMaintainerClient

	// --- MCP Server ---

	// ListMcpServer lists MCP servers with accurate match.
	ListMcpServer(param vo.ListMcpServerParam) (model.Page[model.McpServerBasicInfo], error)

	// SearchMcpServer searches MCP servers with blur match.
	SearchMcpServer(param vo.SearchMcpServerParam) (model.Page[model.McpServerBasicInfo], error)

	// GetMcpServerDetail retrieves detailed MCP server info.
	GetMcpServerDetail(param vo.GetMcpServerDetailParam) (model.McpServerDetailInfo, error)

	// CreateMcpServer creates a new MCP server.
	CreateMcpServer(param vo.CreateMcpServerParam) (string, error)

	// UpdateMcpServer updates an existing MCP server.
	UpdateMcpServer(param vo.UpdateMcpServerParam) (bool, error)

	// DeleteMcpServer deletes an MCP server.
	DeleteMcpServer(param vo.DeleteMcpServerParam) (bool, error)

	// --- A2A Agent ---

	// RegisterAgent registers a new A2A agent.
	RegisterAgent(param vo.RegisterAgentParam) (bool, error)

	// GetAgentCard retrieves an agent card detail.
	GetAgentCard(param vo.GetMaintainerAgentCardParam) (model.AgentCardDetailInfo, error)

	// UpdateAgentCard updates an existing agent card.
	UpdateAgentCard(param vo.UpdateAgentCardParam) (bool, error)

	// DeleteAgent deletes an A2A agent.
	DeleteAgent(param vo.DeleteAgentParam) (bool, error)

	// ListAllVersionOfAgent lists all versions of a specific agent.
	ListAllVersionOfAgent(param vo.ListAgentVersionParam) ([]model.AgentVersionDetail, error)

	// SearchAgentCards searches agent cards by name pattern (blur match).
	SearchAgentCards(param vo.SearchAgentParam) (model.Page[model.AgentCardVersionInfo], error)

	// ListAgentCards lists agent cards (accurate match).
	ListAgentCards(param vo.ListAgentCardsParam) (model.Page[model.AgentCardVersionInfo], error)
}
