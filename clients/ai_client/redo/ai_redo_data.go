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
	"github.com/nacos-group/nacos-sdk-go/v2/common/redo"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
)

// RedoData type constants
const (
	McpServerEndpointRedoDataType = "McpServerEndpointRedoData"
	AgentEndpointRedoDataType     = "AgentEndpointRedoData"
)

// McpServerEndpoint holds MCP server endpoint information for redo.
// This struct is stored in RedoData.data field to preserve all necessary information.
type McpServerEndpoint struct {
	McpName string
	Address string
	Port    int
	Version string
}

// NewMcpServerEndpointRedoData creates a new redo.RedoData for MCP server endpoint
func NewMcpServerEndpointRedoData(mcpName, address string, port int, version string) *redo.RedoData {
	endpoint := &McpServerEndpoint{
		McpName: mcpName,
		Address: address,
		Port:    port,
		Version: version,
	}
	return redo.NewRedoData(endpoint)
}

// AgentEndpointRedoData holds Agent endpoint information for redo.
// This struct is stored in RedoData.data field to preserve all necessary information.
type AgentEndpointRedoData struct {
	AgentName string
	Endpoint  *model.AgentEndpoint
}

// NewAgentEndpointRedoData creates a new redo.RedoData for Agent endpoint
func NewAgentEndpointRedoData(agentName string, endpoint *model.AgentEndpoint) *redo.RedoData {
	data := &AgentEndpointRedoData{
		AgentName: agentName,
		Endpoint:  endpoint,
	}
	return redo.NewRedoData(data)
}
