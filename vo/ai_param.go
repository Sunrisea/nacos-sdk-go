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

package vo

import (
	"github.com/a2aproject/a2a-go/a2a"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
)

// MCP Server callback types
type McpServerChangeCallback func(mcpId, namespaceId, mcpName string, mcpServer model.McpServerDetailInfo)

// Agent Card callback types
type AgentCardChangeCallback func(agentName string, agentCard model.AgentCardDetailInfo)

// GetMcpServerParam represents parameters for retrieving MCP server information
type GetMcpServerParam struct {
	McpName string `param:"mcpName"` // required - Name of the MCP server to query
	Version string `param:"version"` // optional - Version of the MCP server to query
}

// ReleaseMcpServerParam represents parameters for releasing/publishing MCP server
type ReleaseMcpServerParam struct {
	ServerSpec      *model.McpServerBasicInfo   `param:"serverSpec"`      // required - Basic information specification for the MCP server
	ToolSpec        *model.McpToolSpecification `param:"toolSpec"`        // required - Tool specification defining the tools provided by MCP server
	McpEndpointSpec *model.McpEndpointSpec      `param:"mcpEndpointSpec"` // optional - Endpoint specification for MCP server network configuration
}

// RegisterMcpServerEndpointParam represents parameters for registering MCP server endpoint
type RegisterMcpServerEndpointParam struct {
	McpName string `param:"mcpName"` // required - Name of the MCP server
	Address string `param:"address"` // required - IP address or hostname of the MCP server endpoint
	Port    int    `param:"port"`    // required - Port number of the MCP server endpoint
	Version string `param:"version"` // optional - Version of the MCP server
}

// DeregisterMcpServerEndpointParam represents parameters for deregistering MCP server endpoint
type DeregisterMcpServerEndpointParam struct {
	McpName string `param:"mcpName"` // required - Name of the MCP server
	Address string `param:"address"` // required - IP address or hostname of the MCP server endpoint
	Port    int    `param:"port"`    // required - Port number of the MCP server endpoint
}

// SubscribeMcpServerParam represents parameters for subscribing to MCP server changes
type SubscribeMcpServerParam struct {
	McpName           string                  `param:"mcpName"` // required - Name of the MCP server to subscribe to
	Version           string                  `param:"version"` // optional - Version of the MCP server to subscribe to
	SubscribeCallback McpServerChangeCallback // required - Callback function to handle MCP server changes
}

// GetAgentCardParam represents parameters for retrieving Agent Card information
type GetAgentCardParam struct {
	AgentName        string `param:"agentName"`        // required - Name of agent card
	Version          string `param:"version"`          // optional - Target version, if null or empty, get latest version
	RegistrationType string `param:"registrationType"` // optional - URL or SERVICE, default is empty (use agent card setting in nacos)
}

// ReleaseAgentCardParam represents parameters for releasing/publishing Agent Card
type ReleaseAgentCardParam struct {
	AgentCard        *a2a.AgentCard `param:"agentCard"`        // required - Agent card information
	RegistrationType string         `param:"registrationType"` // optional - URL or SERVICE, default is SERVICE
	SetAsLatest      bool           `param:"setAsLatest"`      // optional - Whether to set as latest version
}

// NewReleaseAgentCardParam creates a new ReleaseAgentCardParam with default values
func NewReleaseAgentCardParam() *ReleaseAgentCardParam {
	return &ReleaseAgentCardParam{
		RegistrationType: constant.A2A_ENDPOINT_TYPE_SERVICE,
		SetAsLatest:      false,
	}
}

// RegisterAgentEndpointParam represents parameters for registering Agent endpoint
type RegisterAgentEndpointParam struct {
	AgentName  string `param:"agentName"`  // required - Name of the agent
	Version    string `param:"version"`    // optional - Version of the agent endpoint
	Address    string `param:"address"`    // required - IP address or hostname
	Port       int    `param:"port"`       // required - Port number
	Transport  string `param:"transport"`  // optional - Transport protocol, default is JSONRPC
	Path       string `param:"path"`       // optional - URL path
	SupportTLS bool   `param:"supportTls"` // optional - Whether to support TLS
}

// NewRegisterAgentEndpointParam creates a new RegisterAgentEndpointParam with default values
func NewRegisterAgentEndpointParam() *RegisterAgentEndpointParam {
	return &RegisterAgentEndpointParam{
		Transport:  constant.A2A_ENDPOINT_DEFAULT_TRANSPORT,
		SupportTLS: false,
	}
}

// DeregisterAgentEndpointParam represents parameters for deregistering Agent endpoint
type DeregisterAgentEndpointParam struct {
	AgentName string `param:"agentName"` // required - Name of the agent
	Version   string `param:"version"`   // optional - Version of the agent endpoint
	Address   string `param:"address"`   // required - IP address or hostname
	Port      int    `param:"port"`      // required - Port number
}

// SubscribeAgentCardParam represents parameters for subscribing to Agent Card changes
type SubscribeAgentCardParam struct {
	AgentName         string                  `param:"agentName"` // required - Name of the agent to subscribe to
	Version           string                  `param:"version"`   // optional - Version of the agent to subscribe to
	SubscribeCallback AgentCardChangeCallback // required - Callback function to handle Agent Card changes
}
