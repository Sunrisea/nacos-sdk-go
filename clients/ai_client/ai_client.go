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
	"context"
	"errors"
	"sync"

	"github.com/a2aproject/a2a-go/a2a"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/ai_client/ai_cache"
	ai_redo "github.com/nacos-group/nacos-sdk-go/v2/clients/ai_client/redo"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/nacos_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/common/http_agent"
	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
	"github.com/nacos-group/nacos-sdk-go/v2/common/nacos_server"
	"github.com/nacos-group/nacos-sdk-go/v2/common/remote/rpc"
	"github.com/nacos-group/nacos-sdk-go/v2/common/security"
	"github.com/nacos-group/nacos-sdk-go/v2/inner/uuid"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// AIClient implements IAIClient interface
type AIClient struct {
	ctx    context.Context
	cancel context.CancelFunc
	nacos_client.INacosClient

	mutex       sync.Mutex
	isClosed    bool
	namespaceID string

	// RPC client
	rpcClient rpc.IRpcClient

	// Proxy for RPC communication
	aiProxy *AIProxy

	// Redo service for disconnect recovery
	redoService *ai_redo.AIGrpcRedoService

	// Subscription and cache management
	mcpSubscribeManager   *ai_cache.McpSubscribeManager
	agentSubscribeManager *ai_cache.AgentSubscribeManager
	mcpCacheHolder        *ai_cache.McpServerCacheHolder
	agentCacheHolder      *ai_cache.AgentCacheHolder
}

// NewAIClient creates a new AI client
func NewAIClient(nc nacos_client.INacosClient) (*AIClient, error) {
	return NewAIClientWithRamCredentialProvider(nc, nil)
}

// NewAIClientWithRamCredentialProvider creates a new AI client with RAM credential provider
func NewAIClientWithRamCredentialProvider(nc nacos_client.INacosClient, provider security.RamCredentialProvider) (*AIClient, error) {
	clientConfig, err := nc.GetClientConfig()
	if err != nil {
		return nil, err
	}

	serverConfig, err := nc.GetServerConfig()
	if err != nil {
		return nil, err
	}

	httpAgent, err := nc.GetHttpAgent()
	if err != nil {
		return nil, err
	}

	if err = initLogger(clientConfig); err != nil {
		return nil, err
	}

	namespaceID := clientConfig.NamespaceId
	if namespaceID == "" {
		namespaceID = constant.DEFAULT_NAMESPACE_ID
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &AIClient{
		ctx:                   ctx,
		cancel:                cancel,
		INacosClient:          nc,
		namespaceID:           namespaceID,
		mcpSubscribeManager:   ai_cache.NewMcpSubscribeManager(),
		agentSubscribeManager: ai_cache.NewAgentSubscribeManager(),
	}

	// Initialize RPC client
	if err = client.initRpcClient(ctx, clientConfig, serverConfig, httpAgent, provider); err != nil {
		cancel()
		return nil, err
	}

	return client, nil
}

// initRpcClient initializes the RPC client for AI module
func (c *AIClient) initRpcClient(ctx context.Context, clientCfg constant.ClientConfig, serverCfgs []constant.ServerConfig,
	httpAgent http_agent.IHttpAgent, provider security.RamCredentialProvider) error {

	uid, err := uuid.NewV4()
	if err != nil {
		return err
	}

	aiHeader := map[string][]string{
		"Client-Version": {constant.CLIENT_VERSION},
		"User-Agent":     {constant.CLIENT_VERSION},
		"RequestId":      {uid.String()},
		"Request-Module": {"AI"},
	}

	nacosServer, err := nacos_server.NewNacosServerWithRamCredentialProvider(ctx, serverCfgs, clientCfg, httpAgent, clientCfg.TimeoutMs, clientCfg.Endpoint, aiHeader, provider)
	if err != nil {
		return err
	}

	labels := map[string]string{
		constant.LABEL_SOURCE: constant.LABEL_SOURCE_SDK,
		constant.LABEL_MODULE: "ai",
	}

	iRpcClient, err := rpc.CreateClient(ctx, uid.String(), rpc.GRPC, labels, nacosServer, &clientCfg.TLSCfg, clientCfg.AppConnLabels)
	if err != nil {
		return err
	}

	c.rpcClient = iRpcClient

	rpcClient := c.rpcClient.GetRpcClient()
	rpcClient.Start()

	// Initialize proxy and redo service
	c.aiProxy = NewAIProxy(rpcClient, nacosServer, c.namespaceID, clientCfg.TimeoutMs)

	c.redoService = ai_redo.NewAIGrpcRedoService(c.aiProxy)
	c.redoService.Start()
	rpcClient.RegisterConnectionListener(c.redoService)

	// Initialize cache holders with query functions
	c.mcpCacheHolder = ai_cache.NewMcpServerCacheHolder(
		c.mcpSubscribeManager,
		c.aiProxy.QueryMcpServer,
		c.namespaceID,
	)
	c.agentCacheHolder = ai_cache.NewAgentCacheHolder(
		c.agentSubscribeManager,
		c.aiProxy.QueryAgentCard,
		c.namespaceID,
	)

	return nil
}

func initLogger(clientConfig constant.ClientConfig) error {
	return logger.InitLogger(logger.BuildLoggerConfig(clientConfig))
}

// ========== MCP Server Operations ==========

// GetMcpServer retrieves MCP server information
func (c *AIClient) GetMcpServer(param vo.GetMcpServerParam) (model.McpServerDetailInfo, error) {
	if err := c.checkClosed(); err != nil {
		return model.McpServerDetailInfo{}, err
	}

	if param.McpName == "" {
		return model.McpServerDetailInfo{}, errors.New("mcpName is required")
	}

	result, err := c.aiProxy.QueryMcpServer(param.McpName, param.Version)
	if err != nil {
		return model.McpServerDetailInfo{}, err
	}
	if result == nil {
		return model.McpServerDetailInfo{}, nil
	}
	return *result, nil
}

// ReleaseMcpServer releases/publishes an MCP server
func (c *AIClient) ReleaseMcpServer(param vo.ReleaseMcpServerParam) (string, error) {
	if err := c.checkClosed(); err != nil {
		return "", err
	}

	if param.ServerSpec == nil {
		return "", errors.New("serverSpec is required")
	}
	if param.ServerSpec.Name == "" {
		return "", errors.New("serverSpec.name is required")
	}
	if param.ToolSpec == nil {
		return "", errors.New("toolSpec is required")
	}

	// Auto-fill namespaceId for REF type endpoint spec
	if param.McpEndpointSpec != nil && param.McpEndpointSpec.Type == constant.MCP_ENDPOINT_TYPE_REF {
		if param.McpEndpointSpec.Data == nil {
			param.McpEndpointSpec.Data = make(map[string]string)
		}
		if existingNs, ok := param.McpEndpointSpec.Data["namespaceId"]; !ok || existingNs == "" {
			// Auto-fill with client's namespaceId
			param.McpEndpointSpec.Data["namespaceId"] = c.namespaceID
		} else if existingNs != c.namespaceID {
			// Validate that provided namespaceId matches client's namespaceId
			return "", errors.New("mcpEndpointSpec.data.namespaceId does not match client's namespaceId")
		}
	}

	return c.aiProxy.ReleaseMcpServer(param.ServerSpec, param.ToolSpec, param.McpEndpointSpec)
}

// RegisterMcpServerEndpoint registers an MCP server endpoint
func (c *AIClient) RegisterMcpServerEndpoint(param vo.RegisterMcpServerEndpointParam) error {
	if err := c.checkClosed(); err != nil {
		return err
	}

	if param.McpName == "" {
		return errors.New("mcpName is required")
	}
	if param.Address == "" {
		return errors.New("address is required")
	}
	if param.Port <= 0 {
		return errors.New("port must be positive")
	}

	// Cache for redo before registering
	if c.redoService != nil {
		c.redoService.CacheMcpServerEndpointForRedo(param.McpName, param.Address, param.Port, param.Version)
	}

	err := c.aiProxy.RegisterMcpServerEndpoint(param.McpName, param.Address, param.Port, param.Version)
	if err != nil {
		return err
	}

	// Mark as registered
	if c.redoService != nil {
		c.redoService.McpServerEndpointRegistered(param.McpName)
	}

	return nil
}

// DeregisterMcpServerEndpoint deregisters an MCP server endpoint
func (c *AIClient) DeregisterMcpServerEndpoint(param vo.DeregisterMcpServerEndpointParam) error {
	if err := c.checkClosed(); err != nil {
		return err
	}

	if param.McpName == "" {
		return errors.New("mcpName is required")
	}
	if param.Address == "" {
		return errors.New("address is required")
	}
	if param.Port <= 0 {
		return errors.New("port must be positive")
	}

	// Mark for deregistration
	if c.redoService != nil {
		c.redoService.McpServerEndpointDeregister(param.McpName)
	}

	err := c.aiProxy.DeregisterMcpServerEndpoint(param.McpName, param.Address, param.Port)
	if err != nil {
		return err
	}

	// Mark as deregistered
	if c.redoService != nil {
		c.redoService.McpServerEndpointDeregistered(param.McpName)
	}

	return nil
}

// SubscribeMcpServer subscribes to MCP server changes
func (c *AIClient) SubscribeMcpServer(param vo.SubscribeMcpServerParam) (model.McpServerDetailInfo, error) {
	if err := c.checkClosed(); err != nil {
		return model.McpServerDetailInfo{}, err
	}

	if param.McpName == "" {
		return model.McpServerDetailInfo{}, errors.New("mcpName is required")
	}
	if param.SubscribeCallback == nil {
		return model.McpServerDetailInfo{}, errors.New("subscribeCallback is required")
	}

	// Register the callback
	c.mcpSubscribeManager.RegisterSubscriber(param.McpName, param.Version, param.SubscribeCallback)

	// Try to get from cache first
	cached := c.mcpCacheHolder.GetMcpServer(param.McpName, param.Version)
	if cached != nil {
		// Start update task if not already running
		c.mcpCacheHolder.StartUpdateTask(param.McpName, param.Version)
		return *cached, nil
	}

	// Query from server
	result, err := c.aiProxy.QueryMcpServer(param.McpName, param.Version)
	if err != nil {
		return model.McpServerDetailInfo{}, err
	}

	if result != nil {
		// Cache the result
		c.mcpCacheHolder.SetMcpServer(param.McpName, param.Version, result)
		// Start update task
		c.mcpCacheHolder.StartUpdateTask(param.McpName, param.Version)
		return *result, nil
	}

	return model.McpServerDetailInfo{}, nil
}

// UnsubscribeMcpServer cancels subscription to MCP server changes
func (c *AIClient) UnsubscribeMcpServer(param vo.SubscribeMcpServerParam) error {
	if err := c.checkClosed(); err != nil {
		return err
	}

	if param.McpName == "" {
		return errors.New("mcpName is required")
	}

	c.mcpSubscribeManager.DeregisterSubscriber(param.McpName, param.Version, param.SubscribeCallback)

	// Stop update task if no more subscribers
	if !c.mcpSubscribeManager.IsSubscribed(param.McpName, param.Version) {
		c.mcpCacheHolder.StopUpdateTask(param.McpName, param.Version)
		c.mcpCacheHolder.RemoveMcpServer(param.McpName, param.Version)
	}

	return nil
}

// ========== Agent Card Operations ==========

// GetAgentCard retrieves Agent Card information
func (c *AIClient) GetAgentCard(param vo.GetAgentCardParam) (model.AgentCardDetailInfo, error) {
	if err := c.checkClosed(); err != nil {
		return model.AgentCardDetailInfo{}, err
	}

	if param.AgentName == "" {
		return model.AgentCardDetailInfo{}, errors.New("agentName is required")
	}

	result, err := c.aiProxy.QueryAgentCard(param.AgentName, param.Version, param.RegistrationType)
	if err != nil {
		return model.AgentCardDetailInfo{}, err
	}
	if result == nil {
		return model.AgentCardDetailInfo{}, nil
	}
	return *result, nil
}

// ReleaseAgentCard releases/publishes an Agent Card
func (c *AIClient) ReleaseAgentCard(param vo.ReleaseAgentCardParam) error {
	if err := c.checkClosed(); err != nil {
		return err
	}

	if param.AgentCard == nil {
		return errors.New("agentCard is required")
	}
	if param.AgentCard.Name == "" {
		return errors.New("agentCard.name is required")
	}
	if param.AgentCard.Version == "" {
		return errors.New("agentCard.version is required")
	}
	if param.AgentCard.ProtocolVersion == "" {
		return errors.New("agentCard.protocolVersion is required")
	}

	// Auto-fill preferredTransport with default value if not set (same as Python SDK)
	if param.AgentCard.PreferredTransport == "" {
		param.AgentCard.PreferredTransport = a2a.TransportProtocolJSONRPC
	}

	return c.aiProxy.ReleaseAgentCard(param.AgentCard, param.RegistrationType, param.SetAsLatest)
}

// RegisterAgentEndpoint registers an Agent endpoint
func (c *AIClient) RegisterAgentEndpoint(param vo.RegisterAgentEndpointParam) error {
	if err := c.checkClosed(); err != nil {
		return err
	}

	if param.AgentName == "" {
		return errors.New("agentName is required")
	}
	if param.Address == "" {
		return errors.New("address is required")
	}
	if param.Port <= 0 {
		return errors.New("port must be positive")
	}

	endpoint := &model.AgentEndpoint{
		Transport:  param.Transport,
		Address:    param.Address,
		Port:       param.Port,
		Path:       param.Path,
		SupportTLS: param.SupportTLS,
		Version:    param.Version,
	}

	if endpoint.Transport == "" {
		endpoint.Transport = constant.A2A_ENDPOINT_DEFAULT_TRANSPORT
	}

	// Cache for redo before registering
	if c.redoService != nil {
		c.redoService.CacheAgentEndpointForRedo(param.AgentName, endpoint)
	}

	err := c.aiProxy.RegisterAgentEndpoint(param.AgentName, endpoint)
	if err != nil {
		return err
	}

	// Mark as registered
	if c.redoService != nil {
		c.redoService.AgentEndpointRegistered(param.AgentName)
	}

	return nil
}

// DeregisterAgentEndpoint deregisters an Agent endpoint
func (c *AIClient) DeregisterAgentEndpoint(param vo.DeregisterAgentEndpointParam) error {
	if err := c.checkClosed(); err != nil {
		return err
	}

	if param.AgentName == "" {
		return errors.New("agentName is required")
	}
	if param.Address == "" {
		return errors.New("address is required")
	}
	if param.Port <= 0 {
		return errors.New("port must be positive")
	}

	endpoint := &model.AgentEndpoint{
		Address: param.Address,
		Port:    param.Port,
		Version: param.Version,
	}

	// Mark for deregistration
	if c.redoService != nil {
		c.redoService.AgentEndpointDeregister(param.AgentName)
	}

	err := c.aiProxy.DeregisterAgentEndpoint(param.AgentName, endpoint)
	if err != nil {
		return err
	}

	// Mark as deregistered
	if c.redoService != nil {
		c.redoService.AgentEndpointDeregistered(param.AgentName)
	}

	return nil
}

// SubscribeAgentCard subscribes to Agent Card changes
func (c *AIClient) SubscribeAgentCard(param vo.SubscribeAgentCardParam) (model.AgentCardDetailInfo, error) {
	if err := c.checkClosed(); err != nil {
		return model.AgentCardDetailInfo{}, err
	}

	if param.AgentName == "" {
		return model.AgentCardDetailInfo{}, errors.New("agentName is required")
	}
	if param.SubscribeCallback == nil {
		return model.AgentCardDetailInfo{}, errors.New("subscribeCallback is required")
	}

	// Register the callback
	c.agentSubscribeManager.RegisterSubscriber(param.AgentName, param.Version, param.SubscribeCallback)

	// Try to get from cache first
	cached := c.agentCacheHolder.GetAgentCard(param.AgentName, param.Version)
	if cached != nil {
		// Start update task if not already running
		c.agentCacheHolder.StartUpdateTask(param.AgentName, param.Version, "")
		return *cached, nil
	}

	// Query from server
	result, err := c.aiProxy.QueryAgentCard(param.AgentName, param.Version, "")
	if err != nil {
		return model.AgentCardDetailInfo{}, err
	}

	if result != nil {
		// Cache the result
		c.agentCacheHolder.SetAgentCard(param.AgentName, param.Version, result)
		// Start update task
		c.agentCacheHolder.StartUpdateTask(param.AgentName, param.Version, "")
		return *result, nil
	}

	return model.AgentCardDetailInfo{}, nil
}

// UnsubscribeAgentCard cancels subscription to Agent Card changes
func (c *AIClient) UnsubscribeAgentCard(param vo.SubscribeAgentCardParam) error {
	if err := c.checkClosed(); err != nil {
		return err
	}

	if param.AgentName == "" {
		return errors.New("agentName is required")
	}

	c.agentSubscribeManager.DeregisterSubscriber(param.AgentName, param.Version, param.SubscribeCallback)

	// Stop update task if no more subscribers
	if !c.agentSubscribeManager.IsSubscribed(param.AgentName, param.Version) {
		c.agentCacheHolder.StopUpdateTask(param.AgentName, param.Version)
		c.agentCacheHolder.RemoveAgentCard(param.AgentName, param.Version)
	}

	return nil
}

// ========== Lifecycle Management ==========

// CloseClient closes the AI client
func (c *AIClient) CloseClient() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isClosed {
		return
	}

	c.isClosed = true
	c.cancel()

	// Stop redo service
	if c.redoService != nil {
		c.redoService.Stop()
	}

	// Close cache holders
	if c.mcpCacheHolder != nil {
		c.mcpCacheHolder.Close()
	}
	if c.agentCacheHolder != nil {
		c.agentCacheHolder.Close()
	}

	logger.Infof("AI client closed")
}

// checkClosed checks if the client is closed
func (c *AIClient) checkClosed() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isClosed {
		return errors.New("AI client is closed")
	}
	return nil
}
