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

package redo

import (
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
	"github.com/nacos-group/nacos-sdk-go/v2/common/redo"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
)

const (
	AIModule = "ai"
)

// AIClientProxy defines the interface for AI proxy operations needed by redo service
type AIClientProxy interface {
	RegisterMcpServerEndpoint(mcpName, address string, port int, version string) error
	DeregisterMcpServerEndpoint(mcpName, address string, port int) error
	RegisterAgentEndpoint(agentName string, endpoint *model.AgentEndpoint) error
	DeregisterAgentEndpoint(agentName string, endpoint *model.AgentEndpoint) error
	IsServerHealthy() bool
}

// AIGrpcRedoService handles redo operations for AI module
type AIGrpcRedoService struct {
	*redo.AbstractRedoService
	proxy AIClientProxy
}

// NewAIGrpcRedoService creates a new AI redo service
func NewAIGrpcRedoService(proxy AIClientProxy) *AIGrpcRedoService {
	service := &AIGrpcRedoService{
		AbstractRedoService: redo.NewAbstractRedoService(AIModule),
		proxy:               proxy,
	}
	// Set self as the redo task
	service.SetRedoTask(service)
	return service
}

// RedoTask implements redo.RedoTask interface
func (s *AIGrpcRedoService) RedoTask() error {
	// Redo MCP server endpoints
	if err := s.redoForMcpServerEndpoints(); err != nil {
		logger.Errorf("[%s] redo for mcp server endpoints failed: %v", AIModule, err)
	}

	// Redo Agent endpoints
	if err := s.redoForAgentEndpoints(); err != nil {
		logger.Errorf("[%s] redo for agent endpoints failed: %v", AIModule, err)
	}

	return nil
}

// ========== MCP Server Endpoint Redo Operations ==========

// redoForMcpServerEndpoints processes all MCP server endpoint redo data
func (s *AIGrpcRedoService) redoForMcpServerEndpoints() error {
	redoDataList := s.FindRedoDataByType(McpServerEndpointRedoDataType)
	for _, data := range redoDataList {
		mcpEndpoint, ok := data.GetData().(*McpServerEndpoint)
		if !ok {
			continue
		}

		if err := s.redoForMcpEndpoint(data, mcpEndpoint); err != nil {
			logger.Errorf("[%s] redo for mcp endpoint %s failed: %v", AIModule, mcpEndpoint.McpName, err)
		}
	}
	return nil
}

// redoForMcpEndpoint handles a single MCP endpoint redo operation
func (s *AIGrpcRedoService) redoForMcpEndpoint(redoData *redo.RedoData, endpoint *McpServerEndpoint) error {
	if redoData == nil || endpoint == nil {
		return nil
	}

	redoType := redoData.GetRedoType()
	mcpName := endpoint.McpName

	switch redoType {
	case redo.RedoTypeRegister:
		if s.proxy != nil && !s.proxy.IsServerHealthy() {
			logger.Warnf("[%s] server is not healthy, skip register mcp endpoint", AIModule)
			return nil
		}
		return s.doRegisterMcpServerEndpoint(mcpName, endpoint.Address, endpoint.Port, endpoint.Version)

	case redo.RedoTypeUnregister:
		if s.proxy != nil && !s.proxy.IsServerHealthy() {
			logger.Warnf("[%s] server is not healthy, skip deregister mcp endpoint", AIModule)
			return nil
		}
		return s.doDeregisterMcpServerEndpoint(mcpName, endpoint.Address, endpoint.Port)

	case redo.RedoTypeRemove:
		s.removeMcpServerEndpointForRedo(mcpName)
		return nil

	default:
		return nil
	}
}

// doRegisterMcpServerEndpoint performs the actual registration
func (s *AIGrpcRedoService) doRegisterMcpServerEndpoint(mcpName, address string, port int, version string) error {
	if s.proxy == nil {
		return fmt.Errorf("proxy is not set")
	}

	err := s.proxy.RegisterMcpServerEndpoint(mcpName, address, port, version)
	if err != nil {
		return err
	}

	// Mark as registered
	s.McpServerEndpointRegistered(mcpName)
	return nil
}

// doDeregisterMcpServerEndpoint performs the actual deregistration
func (s *AIGrpcRedoService) doDeregisterMcpServerEndpoint(mcpName, address string, port int) error {
	if s.proxy == nil {
		return fmt.Errorf("proxy is not set")
	}

	err := s.proxy.DeregisterMcpServerEndpoint(mcpName, address, port)
	if err != nil {
		return err
	}

	// Mark as deregistered
	s.McpServerEndpointDeregistered(mcpName)
	return nil
}

// CacheMcpServerEndpointForRedo caches MCP server endpoint data for redo
func (s *AIGrpcRedoService) CacheMcpServerEndpointForRedo(mcpName, address string, port int, version string) {
	redoData := NewMcpServerEndpointRedoData(mcpName, address, port, version)
	s.CacheRedoData(McpServerEndpointRedoDataType, mcpName, redoData)
	logger.Debugf("[%s] cached mcp server endpoint for redo: mcpName=%s, address=%s, port=%d", AIModule, mcpName, address, port)
}

// McpServerEndpointRegistered marks MCP server endpoint as registered
func (s *AIGrpcRedoService) McpServerEndpointRegistered(mcpName string) {
	s.DataRegistered(McpServerEndpointRedoDataType, mcpName)
	logger.Debugf("[%s] mcp server endpoint registered: mcpName=%s", AIModule, mcpName)
}

// McpServerEndpointDeregister marks MCP server endpoint for deregistration
func (s *AIGrpcRedoService) McpServerEndpointDeregister(mcpName string) {
	s.DataDeregister(McpServerEndpointRedoDataType, mcpName)
	logger.Debugf("[%s] mcp server endpoint marked for deregister: mcpName=%s", AIModule, mcpName)
}

// McpServerEndpointDeregistered marks MCP server endpoint as deregistered
func (s *AIGrpcRedoService) McpServerEndpointDeregistered(mcpName string) {
	s.DataDeregistered(McpServerEndpointRedoDataType, mcpName)
	logger.Debugf("[%s] mcp server endpoint deregistered: mcpName=%s", AIModule, mcpName)
}

// IsMcpServerEndpointRegistered checks if MCP server endpoint is registered
func (s *AIGrpcRedoService) IsMcpServerEndpointRegistered(mcpName string) bool {
	return s.IsDataRegistered(McpServerEndpointRedoDataType, mcpName)
}

// removeMcpServerEndpointForRedo removes MCP server endpoint from redo cache
func (s *AIGrpcRedoService) removeMcpServerEndpointForRedo(mcpName string) {
	s.RemoveRedoData(McpServerEndpointRedoDataType, mcpName)
	logger.Debugf("[%s] removed mcp server endpoint from redo cache: mcpName=%s", AIModule, mcpName)
}

// ========== Agent Endpoint Redo Operations ==========

// redoForAgentEndpoints processes all Agent endpoint redo data
func (s *AIGrpcRedoService) redoForAgentEndpoints() error {
	redoDataList := s.FindRedoDataByType(AgentEndpointRedoDataType)
	for _, data := range redoDataList {
		agentData, ok := data.GetData().(*AgentEndpointRedoData)
		if !ok {
			continue
		}

		if err := s.redoForAgentEndpoint(data, agentData); err != nil {
			logger.Errorf("[%s] redo for agent endpoint %s failed: %v", AIModule, agentData.AgentName, err)
		}
	}
	return nil
}

// redoForAgentEndpoint handles a single Agent endpoint redo operation
func (s *AIGrpcRedoService) redoForAgentEndpoint(redoData *redo.RedoData, agentData *AgentEndpointRedoData) error {
	if redoData == nil || agentData == nil {
		return nil
	}

	redoType := redoData.GetRedoType()
	agentName := agentData.AgentName
	endpoint := agentData.Endpoint

	switch redoType {
	case redo.RedoTypeRegister:
		if s.proxy != nil && !s.proxy.IsServerHealthy() {
			logger.Warnf("[%s] server is not healthy, skip register agent endpoint", AIModule)
			return nil
		}
		return s.doRegisterAgentEndpoint(agentName, endpoint)

	case redo.RedoTypeUnregister:
		if s.proxy != nil && !s.proxy.IsServerHealthy() {
			logger.Warnf("[%s] server is not healthy, skip deregister agent endpoint", AIModule)
			return nil
		}
		return s.doDeregisterAgentEndpoint(agentName, endpoint)

	case redo.RedoTypeRemove:
		s.removeAgentEndpointForRedo(agentName)
		return nil

	default:
		return nil
	}
}

// doRegisterAgentEndpoint performs the actual registration
func (s *AIGrpcRedoService) doRegisterAgentEndpoint(agentName string, endpoint *model.AgentEndpoint) error {
	if s.proxy == nil {
		return fmt.Errorf("proxy is not set")
	}

	err := s.proxy.RegisterAgentEndpoint(agentName, endpoint)
	if err != nil {
		return err
	}

	// Mark as registered
	s.AgentEndpointRegistered(agentName)
	return nil
}

// doDeregisterAgentEndpoint performs the actual deregistration
func (s *AIGrpcRedoService) doDeregisterAgentEndpoint(agentName string, endpoint *model.AgentEndpoint) error {
	if s.proxy == nil {
		return fmt.Errorf("proxy is not set")
	}

	err := s.proxy.DeregisterAgentEndpoint(agentName, endpoint)
	if err != nil {
		return err
	}

	// Mark as deregistered
	s.AgentEndpointDeregistered(agentName)
	return nil
}

// CacheAgentEndpointForRedo caches Agent endpoint data for redo
func (s *AIGrpcRedoService) CacheAgentEndpointForRedo(agentName string, endpoint *model.AgentEndpoint) {
	redoData := NewAgentEndpointRedoData(agentName, endpoint)
	s.CacheRedoData(AgentEndpointRedoDataType, agentName, redoData)
	logger.Debugf("[%s] cached agent endpoint for redo: agentName=%s, address=%s, port=%d", AIModule, agentName, endpoint.Address, endpoint.Port)
}

// AgentEndpointRegistered marks Agent endpoint as registered
func (s *AIGrpcRedoService) AgentEndpointRegistered(agentName string) {
	s.DataRegistered(AgentEndpointRedoDataType, agentName)
	logger.Debugf("[%s] agent endpoint registered: agentName=%s", AIModule, agentName)
}

// AgentEndpointDeregister marks Agent endpoint for deregistration
func (s *AIGrpcRedoService) AgentEndpointDeregister(agentName string) {
	s.DataDeregister(AgentEndpointRedoDataType, agentName)
	logger.Debugf("[%s] agent endpoint marked for deregister: agentName=%s", AIModule, agentName)
}

// AgentEndpointDeregistered marks Agent endpoint as deregistered
func (s *AIGrpcRedoService) AgentEndpointDeregistered(agentName string) {
	s.DataDeregistered(AgentEndpointRedoDataType, agentName)
	logger.Debugf("[%s] agent endpoint deregistered: agentName=%s", AIModule, agentName)
}

// IsAgentEndpointRegistered checks if Agent endpoint is registered
func (s *AIGrpcRedoService) IsAgentEndpointRegistered(agentName string) bool {
	return s.IsDataRegistered(AgentEndpointRedoDataType, agentName)
}

// removeAgentEndpointForRedo removes Agent endpoint from redo cache
func (s *AIGrpcRedoService) removeAgentEndpointForRedo(agentName string) {
	s.RemoveRedoData(AgentEndpointRedoDataType, agentName)
	logger.Debugf("[%s] removed agent endpoint from redo cache: agentName=%s", AIModule, agentName)
}
