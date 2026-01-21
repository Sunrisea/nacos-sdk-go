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

package rpc_request

import (
	"github.com/a2aproject/a2a-go/a2a"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
)

// AIRequest is the base request for all AI module requests
type AIRequest struct {
	*Request
	Module string `json:"module"`
}

func NewAIRequest() *AIRequest {
	request := Request{
		Headers: make(map[string]string, 8),
	}
	return &AIRequest{
		Request: &request,
		Module:  constant.LABEL_MODULE_AI,
	}
}

// ========== MCP Server Requests ==========

// McpRequest is the base request for MCP operations
type McpRequest struct {
	*AIRequest
	NamespaceID string `json:"namespaceId,omitempty"`
	McpID       string `json:"mcpId,omitempty"`
	McpName     string `json:"mcpName,omitempty"`
}

func NewMcpRequest() *McpRequest {
	return &McpRequest{
		AIRequest: NewAIRequest(),
	}
}

// QueryMcpServerRequest is the request for querying MCP server details
type QueryMcpServerRequest struct {
	*McpRequest
	Version string `json:"version,omitempty"`
}

func NewQueryMcpServerRequest(namespaceID, mcpName, version string) *QueryMcpServerRequest {
	req := &QueryMcpServerRequest{
		McpRequest: NewMcpRequest(),
		Version:    version,
	}
	req.NamespaceID = namespaceID
	req.McpName = mcpName
	return req
}

func (r *QueryMcpServerRequest) GetRequestType() string {
	return constant.QUERY_MCP_SERVER_REQUEST_NAME
}

// ReleaseMcpServerRequest is the request for releasing/publishing MCP server
type ReleaseMcpServerRequest struct {
	*McpRequest
	ServerSpecification   *model.McpServerBasicInfo   `json:"serverSpecification,omitempty"`
	ToolSpecification     *model.McpToolSpecification `json:"toolSpecification,omitempty"`
	EndpointSpecification *model.McpEndpointSpec      `json:"endpointSpecification,omitempty"`
}

func NewReleaseMcpServerRequest(namespaceID string, serverSpec *model.McpServerBasicInfo, toolSpec *model.McpToolSpecification, endpointSpec *model.McpEndpointSpec) *ReleaseMcpServerRequest {
	req := &ReleaseMcpServerRequest{
		McpRequest:            NewMcpRequest(),
		ServerSpecification:   serverSpec,
		ToolSpecification:     toolSpec,
		EndpointSpecification: endpointSpec,
	}
	req.NamespaceID = namespaceID
	return req
}

func (r *ReleaseMcpServerRequest) GetRequestType() string {
	return constant.RELEASE_MCP_SERVER_REQUEST_NAME
}

// McpServerEndpointRequest is the request for registering/deregistering MCP server endpoint
type McpServerEndpointRequest struct {
	*McpRequest
	Address string `json:"address,omitempty"`
	Port    int    `json:"port,omitempty"`
	Version string `json:"version,omitempty"`
	Type    string `json:"type,omitempty"` // REGISTER_ENDPOINT or DE_REGISTER_ENDPOINT
}

func NewMcpServerEndpointRequest(namespaceID, mcpName, address string, port int, version, opType string) *McpServerEndpointRequest {
	req := &McpServerEndpointRequest{
		McpRequest: NewMcpRequest(),
		Address:    address,
		Port:       port,
		Version:    version,
		Type:       opType,
	}
	req.NamespaceID = namespaceID
	req.McpName = mcpName
	return req
}

func (r *McpServerEndpointRequest) GetRequestType() string {
	return constant.MCP_SERVER_ENDPOINT_REQUEST_NAME
}

// ========== Agent Card Requests ==========

// AgentRequest is the base request for Agent operations
type AgentRequest struct {
	*AIRequest
	NamespaceID string `json:"namespaceId,omitempty"`
	AgentName   string `json:"agentName,omitempty"`
}

func NewAgentRequest() *AgentRequest {
	return &AgentRequest{
		AIRequest: NewAIRequest(),
	}
}

// QueryAgentCardRequest is the request for querying agent card details
type QueryAgentCardRequest struct {
	*AgentRequest
	Version          string `json:"version,omitempty"`
	RegistrationType string `json:"registrationType,omitempty"`
}

func NewQueryAgentCardRequest(namespaceID, agentName, version, registrationType string) *QueryAgentCardRequest {
	req := &QueryAgentCardRequest{
		AgentRequest:     NewAgentRequest(),
		Version:          version,
		RegistrationType: registrationType,
	}
	req.NamespaceID = namespaceID
	req.AgentName = agentName
	return req
}

func (r *QueryAgentCardRequest) GetRequestType() string {
	return constant.QUERY_AGENT_CARD_REQUEST_NAME
}

// ReleaseAgentCardRequest is the request for releasing/publishing agent card
type ReleaseAgentCardRequest struct {
	*AgentRequest
	AgentCard        *a2a.AgentCard `json:"agentCard,omitempty"`
	RegistrationType string         `json:"registrationType,omitempty"`
	SetAsLatest      bool           `json:"setAsLatest,omitempty"`
}

func NewReleaseAgentCardRequest(namespaceID, agentName string, agentCard *a2a.AgentCard, registrationType string, setAsLatest bool) *ReleaseAgentCardRequest {
	req := &ReleaseAgentCardRequest{
		AgentRequest:     NewAgentRequest(),
		AgentCard:        agentCard,
		RegistrationType: registrationType,
		SetAsLatest:      setAsLatest,
	}
	req.NamespaceID = namespaceID
	req.AgentName = agentName
	return req
}

func (r *ReleaseAgentCardRequest) GetRequestType() string {
	return constant.RELEASE_AGENT_CARD_REQUEST_NAME
}

// AgentEndpointRequest is the request for registering/deregistering agent endpoint
type AgentEndpointRequest struct {
	*AgentRequest
	Endpoint *model.AgentEndpoint `json:"endpoint,omitempty"`
	Type     string               `json:"type,omitempty"` // REGISTER_ENDPOINT or DE_REGISTER_ENDPOINT
}

func NewAgentEndpointRequest(namespaceID, agentName string, endpoint *model.AgentEndpoint, opType string) *AgentEndpointRequest {
	req := &AgentEndpointRequest{
		AgentRequest: NewAgentRequest(),
		Endpoint:     endpoint,
		Type:         opType,
	}
	req.NamespaceID = namespaceID
	req.AgentName = agentName
	return req
}

func (r *AgentEndpointRequest) GetRequestType() string {
	return constant.AGENT_ENDPOINT_REQUEST_NAME
}
