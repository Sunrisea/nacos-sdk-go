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
	"errors"
	"time"

	"github.com/a2aproject/a2a-go/a2a"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
	"github.com/nacos-group/nacos-sdk-go/v2/common/monitor"
	"github.com/nacos-group/nacos-sdk-go/v2/common/nacos_server"
	"github.com/nacos-group/nacos-sdk-go/v2/common/remote/rpc"
	"github.com/nacos-group/nacos-sdk-go/v2/common/remote/rpc/rpc_request"
	"github.com/nacos-group/nacos-sdk-go/v2/common/remote/rpc/rpc_response"
	"github.com/nacos-group/nacos-sdk-go/v2/common/security"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
)

var (
	ErrMcpRegistryNotSupported   = errors.New("server version too low, not support mcp registry feature")
	ErrAgentRegistryNotSupported = errors.New("server version too low, not support agent registry feature")
)

// IAIProxy defines the interface for AI service proxy
type IAIProxy interface {
	// MCP operations
	QueryMcpServer(mcpName, version string) (*model.McpServerDetailInfo, error)
	ReleaseMcpServer(serverSpec *model.McpServerBasicInfo, toolSpec *model.McpToolSpecification, endpointSpec *model.McpEndpointSpec) (string, error)
	RegisterMcpServerEndpoint(mcpName, address string, port int, version string) error
	DeregisterMcpServerEndpoint(mcpName, address string, port int) error

	// Agent operations
	QueryAgentCard(agentName, version, registrationType string) (*model.AgentCardDetailInfo, error)
	ReleaseAgentCard(agentCard *a2a.AgentCard, registrationType string, setAsLatest bool) error
	RegisterAgentEndpoint(agentName string, endpoint *model.AgentEndpoint) error
	DeregisterAgentEndpoint(agentName string, endpoint *model.AgentEndpoint) error
}

// AIProxy implements IAIProxy
type AIProxy struct {
	rpcClient   *rpc.RpcClient
	nacosServer *nacos_server.NacosServer
	namespaceID string
	timeoutMs   uint64
}

// NewAIProxy creates a new AIProxy instance
func NewAIProxy(rpcClient *rpc.RpcClient, nacosServer *nacos_server.NacosServer, namespaceID string, timeoutMs uint64) *AIProxy {
	return &AIProxy{
		rpcClient:   rpcClient,
		nacosServer: nacosServer,
		namespaceID: namespaceID,
		timeoutMs:   timeoutMs,
	}
}

// requestToServer sends a request to the server with security info injected
func (p *AIProxy) requestToServer(request rpc_request.IRequest, resourceName string) (rpc_response.IResponse, error) {
	start := time.Now()
	p.nacosServer.InjectSecurityInfo(request.GetHeaders(), security.BuildAIResource(p.namespaceID, resourceName))
	response, err := p.rpcClient.Request(request, int64(p.timeoutMs))
	monitor.GetAIRequestMonitor(constant.GRPC, request.GetRequestType(), rpc_response.GetGrpcResponseStatusCode(response)).Observe(float64(time.Now().Nanosecond() - start.Nanosecond()))
	return response, err
}

// isAbilitySupportedByServer checks if the server supports a specific ability
func (p *AIProxy) isAbilitySupportedByServer(abilityKey rpc.AbilityKey) bool {
	return p.rpcClient.IsAbilitySupportedByServer(abilityKey)
}

// ========== MCP Server Operations ==========

// QueryMcpServer queries MCP server information
func (p *AIProxy) QueryMcpServer(mcpName, version string) (*model.McpServerDetailInfo, error) {
	if !p.isAbilitySupportedByServer(rpc.ServerMcpRegistry) {
		return nil, ErrMcpRegistryNotSupported
	}

	request := rpc_request.NewQueryMcpServerRequest(p.namespaceID, mcpName, version)
	response, err := p.requestToServer(request, mcpName)
	if err != nil {
		logger.Errorf("QueryMcpServer failed, mcpName=%s, version=%s, error=%v", mcpName, version, err)
		return nil, err
	}

	queryResp, ok := response.(*rpc_response.QueryMcpServerResponse)
	if !ok {
		return nil, errors.New("invalid response type for QueryMcpServerRequest")
	}

	if !queryResp.IsSuccess() {
		return nil, errors.New(queryResp.GetMessage())
	}

	return queryResp.McpServerDetailInfo, nil
}

// ReleaseMcpServer releases/publishes an MCP server
func (p *AIProxy) ReleaseMcpServer(serverSpec *model.McpServerBasicInfo, toolSpec *model.McpToolSpecification, endpointSpec *model.McpEndpointSpec) (string, error) {
	if !p.isAbilitySupportedByServer(rpc.ServerMcpRegistry) {
		return "", ErrMcpRegistryNotSupported
	}

	mcpName := ""
	if serverSpec != nil {
		mcpName = serverSpec.Name
	}

	request := rpc_request.NewReleaseMcpServerRequest(p.namespaceID, serverSpec, toolSpec, endpointSpec)
	response, err := p.requestToServer(request, mcpName)
	if err != nil {
		logger.Errorf("ReleaseMcpServer failed, error=%v", err)
		return "", err
	}

	releaseResp, ok := response.(*rpc_response.ReleaseMcpServerResponse)
	if !ok {
		return "", errors.New("invalid response type for ReleaseMcpServerRequest")
	}

	if !releaseResp.IsSuccess() {
		return "", errors.New(releaseResp.GetMessage())
	}

	return releaseResp.McpID, nil
}

// RegisterMcpServerEndpoint registers an MCP server endpoint
func (p *AIProxy) RegisterMcpServerEndpoint(mcpName, address string, port int, version string) error {
	if !p.isAbilitySupportedByServer(rpc.ServerMcpRegistry) {
		return ErrMcpRegistryNotSupported
	}

	request := rpc_request.NewMcpServerEndpointRequest(p.namespaceID, mcpName, address, port, version, constant.REGISTER_ENDPOINT)
	response, err := p.requestToServer(request, mcpName)
	if err != nil {
		logger.Errorf("RegisterMcpServerEndpoint failed, mcpName=%s, address=%s, port=%d, error=%v", mcpName, address, port, err)
		return err
	}

	endpointResp, ok := response.(*rpc_response.McpServerEndpointResponse)
	if !ok {
		return errors.New("invalid response type for McpServerEndpointRequest")
	}

	if !endpointResp.IsSuccess() {
		return errors.New(endpointResp.GetMessage())
	}

	return nil
}

// DeregisterMcpServerEndpoint deregisters an MCP server endpoint
func (p *AIProxy) DeregisterMcpServerEndpoint(mcpName, address string, port int) error {
	if !p.isAbilitySupportedByServer(rpc.ServerMcpRegistry) {
		return ErrMcpRegistryNotSupported
	}

	request := rpc_request.NewMcpServerEndpointRequest(p.namespaceID, mcpName, address, port, "", constant.DE_REGISTER_ENDPOINT)
	response, err := p.requestToServer(request, mcpName)
	if err != nil {
		logger.Errorf("DeregisterMcpServerEndpoint failed, mcpName=%s, address=%s, port=%d, error=%v", mcpName, address, port, err)
		return err
	}

	endpointResp, ok := response.(*rpc_response.McpServerEndpointResponse)
	if !ok {
		return errors.New("invalid response type for McpServerEndpointRequest")
	}

	if !endpointResp.IsSuccess() {
		return errors.New(endpointResp.GetMessage())
	}

	return nil
}

// ========== Agent Card Operations ==========

// QueryAgentCard queries agent card information
func (p *AIProxy) QueryAgentCard(agentName, version, registrationType string) (*model.AgentCardDetailInfo, error) {
	if !p.isAbilitySupportedByServer(rpc.ServerAgentRegistry) {
		return nil, ErrAgentRegistryNotSupported
	}

	request := rpc_request.NewQueryAgentCardRequest(p.namespaceID, agentName, version, registrationType)
	response, err := p.requestToServer(request, agentName)
	if err != nil {
		logger.Errorf("QueryAgentCard failed, agentName=%s, version=%s, error=%v", agentName, version, err)
		return nil, err
	}

	queryResp, ok := response.(*rpc_response.QueryAgentCardResponse)
	if !ok {
		return nil, errors.New("invalid response type for QueryAgentCardRequest")
	}

	if !queryResp.IsSuccess() {
		return nil, errors.New(queryResp.GetMessage())
	}

	return queryResp.AgentCardDetailInfo, nil
}

// ReleaseAgentCard releases/publishes an agent card
func (p *AIProxy) ReleaseAgentCard(agentCard *a2a.AgentCard, registrationType string, setAsLatest bool) error {
	if !p.isAbilitySupportedByServer(rpc.ServerAgentRegistry) {
		return ErrAgentRegistryNotSupported
	}

	agentName := ""
	if agentCard != nil {
		agentName = agentCard.Name
	}

	request := rpc_request.NewReleaseAgentCardRequest(p.namespaceID, agentName, agentCard, registrationType, setAsLatest)
	response, err := p.requestToServer(request, agentName)
	if err != nil {
		logger.Errorf("ReleaseAgentCard failed, agentName=%s, error=%v", agentName, err)
		return err
	}

	releaseResp, ok := response.(*rpc_response.ReleaseAgentCardResponse)
	if !ok {
		return errors.New("invalid response type for ReleaseAgentCardRequest")
	}

	if !releaseResp.IsSuccess() {
		return errors.New(releaseResp.GetMessage())
	}

	return nil
}

// RegisterAgentEndpoint registers an agent endpoint
func (p *AIProxy) RegisterAgentEndpoint(agentName string, endpoint *model.AgentEndpoint) error {
	if !p.isAbilitySupportedByServer(rpc.ServerAgentRegistry) {
		return ErrAgentRegistryNotSupported
	}

	request := rpc_request.NewAgentEndpointRequest(p.namespaceID, agentName, endpoint, constant.REGISTER_ENDPOINT)
	response, err := p.requestToServer(request, agentName)
	if err != nil {
		logger.Errorf("RegisterAgentEndpoint failed, agentName=%s, error=%v", agentName, err)
		return err
	}

	endpointResp, ok := response.(*rpc_response.AgentEndpointResponse)
	if !ok {
		return errors.New("invalid response type for AgentEndpointRequest")
	}

	if !endpointResp.IsSuccess() {
		return errors.New(endpointResp.GetMessage())
	}

	return nil
}

// DeregisterAgentEndpoint deregisters an agent endpoint
func (p *AIProxy) DeregisterAgentEndpoint(agentName string, endpoint *model.AgentEndpoint) error {
	if !p.isAbilitySupportedByServer(rpc.ServerAgentRegistry) {
		return ErrAgentRegistryNotSupported
	}

	request := rpc_request.NewAgentEndpointRequest(p.namespaceID, agentName, endpoint, constant.DE_REGISTER_ENDPOINT)
	response, err := p.requestToServer(request, agentName)
	if err != nil {
		logger.Errorf("DeregisterAgentEndpoint failed, agentName=%s, error=%v", agentName, err)
		return err
	}

	endpointResp, ok := response.(*rpc_response.AgentEndpointResponse)
	if !ok {
		return errors.New("invalid response type for AgentEndpointRequest")
	}

	if !endpointResp.IsSuccess() {
		return errors.New(endpointResp.GetMessage())
	}

	return nil
}

// IsServerHealthy checks if the server connection is healthy
func (p *AIProxy) IsServerHealthy() bool {
	if p.rpcClient == nil {
		return false
	}
	return p.rpcClient.IsRunning()
}
