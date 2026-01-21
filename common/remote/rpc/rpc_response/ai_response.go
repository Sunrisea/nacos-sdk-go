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

package rpc_response

import "github.com/nacos-group/nacos-sdk-go/v2/model"

// ========== MCP Server Responses ==========

// QueryMcpServerResponse is the response for MCP server query requests
type QueryMcpServerResponse struct {
	*Response
	McpServerDetailInfo *model.McpServerDetailInfo `json:"mcpServerDetailInfo,omitempty"`
}

func (r *QueryMcpServerResponse) GetResponseType() string {
	return "QueryMcpServerResponse"
}

// ReleaseMcpServerResponse is the response for MCP server release/publish requests
type ReleaseMcpServerResponse struct {
	*Response
	McpID string `json:"mcpId,omitempty"`
}

func (r *ReleaseMcpServerResponse) GetResponseType() string {
	return "ReleaseMcpServerResponse"
}

// McpServerEndpointResponse is the response for MCP server endpoint registration/deregistration requests
type McpServerEndpointResponse struct {
	*Response
	Type string `json:"type,omitempty"` // register or deregister
}

func (r *McpServerEndpointResponse) GetResponseType() string {
	return "McpServerEndpointResponse"
}

// ========== Agent Card Responses ==========

// QueryAgentCardResponse is the response for agent card query requests
type QueryAgentCardResponse struct {
	*Response
	AgentCardDetailInfo *model.AgentCardDetailInfo `json:"agentCardDetailInfo,omitempty"`
}

func (r *QueryAgentCardResponse) GetResponseType() string {
	return "QueryAgentCardResponse"
}

// ReleaseAgentCardResponse is the response for agent card release requests
type ReleaseAgentCardResponse struct {
	*Response
}

func (r *ReleaseAgentCardResponse) GetResponseType() string {
	return "ReleaseAgentCardResponse"
}

// AgentEndpointResponse is the response for agent endpoint registration/deregistration requests
type AgentEndpointResponse struct {
	*Response
	Type string `json:"type,omitempty"` // register or deregister
}

func (r *AgentEndpointResponse) GetResponseType() string {
	return "AgentEndpointResponse"
}
