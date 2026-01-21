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

package ai_client

import (
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

//go:generate mockgen -destination ../../mock/mock_ai_client_interface.go -package mock -source=./ai_client_interface.go

// IAIClient is the interface for AI service operations
type IAIClient interface {
	// ========== MCP Server Operations ==========

	// GetMcpServer retrieves MCP server information by name and version
	// mcpName  required - Name of the MCP server to query
	// version  optional - Version of the MCP server to query
	GetMcpServer(param vo.GetMcpServerParam) (model.McpServerDetailInfo, error)

	// ReleaseMcpServer releases/publishes a new MCP server to the registry
	// serverSpec  required - Basic information specification for the MCP server
	// toolSpec    required - Tool specification defining the tools provided by MCP server
	// mcpEndpointSpec optional - Endpoint specification for MCP server network configuration
	// Returns the MCP server ID assigned by the server
	ReleaseMcpServer(param vo.ReleaseMcpServerParam) (string, error)

	// RegisterMcpServerEndpoint registers an MCP server endpoint
	// mcpName  required - Name of the MCP server
	// address  required - IP address or hostname of the MCP server endpoint
	// port     required - Port number of the MCP server endpoint
	// version  optional - Version of the MCP server
	RegisterMcpServerEndpoint(param vo.RegisterMcpServerEndpointParam) error

	// DeregisterMcpServerEndpoint deregisters an MCP server endpoint
	// mcpName  required - Name of the MCP server
	// address  required - IP address or hostname of the MCP server endpoint
	// port     required - Port number of the MCP server endpoint
	DeregisterMcpServerEndpoint(param vo.DeregisterMcpServerEndpointParam) error

	// SubscribeMcpServer subscribes to MCP server changes
	// mcpName  required - Name of the MCP server to subscribe to
	// version  optional - Version of the MCP server to subscribe to
	// subscribeCallback required - Callback function to handle MCP server changes
	// Returns the initial MCP server detail info
	SubscribeMcpServer(param vo.SubscribeMcpServerParam) (model.McpServerDetailInfo, error)

	// UnsubscribeMcpServer cancels subscription to MCP server changes
	// mcpName  required - Name of the MCP server to unsubscribe from
	// version  optional - Version of the MCP server
	UnsubscribeMcpServer(param vo.SubscribeMcpServerParam) error

	// ========== Agent Card Operations ==========

	// GetAgentCard retrieves Agent Card information by name and version
	// agentName  required - Name of agent card
	// version    optional - Target version, if null or empty, get latest version
	// registrationType optional - URL or SERVICE
	GetAgentCard(param vo.GetAgentCardParam) (model.AgentCardDetailInfo, error)

	// ReleaseAgentCard releases/publishes a new Agent Card to the registry
	// agentCard        required - Agent card information
	// registrationType optional - URL or SERVICE, default is SERVICE
	// setAsLatest      optional - Whether to set as latest version
	ReleaseAgentCard(param vo.ReleaseAgentCardParam) error

	// RegisterAgentEndpoint registers an Agent endpoint
	// agentName  required - Name of the agent
	// address    required - IP address or hostname
	// port       required - Port number
	// version    optional - Version of the agent endpoint
	// transport  optional - Transport protocol, default is JSONRPC
	// path       optional - URL path
	// supportTls optional - Whether to support TLS
	RegisterAgentEndpoint(param vo.RegisterAgentEndpointParam) error

	// DeregisterAgentEndpoint deregisters an Agent endpoint
	// agentName  required - Name of the agent
	// address    required - IP address or hostname
	// port       required - Port number
	// version    optional - Version of the agent endpoint
	DeregisterAgentEndpoint(param vo.DeregisterAgentEndpointParam) error

	// SubscribeAgentCard subscribes to Agent Card changes
	// agentName  required - Name of the agent to subscribe to
	// version    optional - Version of the agent to subscribe to
	// subscribeCallback required - Callback function to handle Agent Card changes
	// Returns the initial Agent Card detail info
	SubscribeAgentCard(param vo.SubscribeAgentCardParam) (model.AgentCardDetailInfo, error)

	// UnsubscribeAgentCard cancels subscription to Agent Card changes
	// agentName  required - Name of the agent to unsubscribe from
	// version    optional - Version of the agent
	UnsubscribeAgentCard(param vo.SubscribeAgentCardParam) error

	// ========== Lifecycle Management ==========

	// CloseClient closes the AI client and releases all resources
	CloseClient()
}
