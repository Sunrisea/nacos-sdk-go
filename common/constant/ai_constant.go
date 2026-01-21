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

package constant

// AI Module Constants

const (
	// LABEL_MODULE_AI is the label for AI module
	LABEL_MODULE_AI = "ai"

	// MCP endpoint operation types
	REGISTER_ENDPOINT    = "registerEndpoint"
	DE_REGISTER_ENDPOINT = "deregisterEndpoint"

	// MCP endpoint types
	MCP_ENDPOINT_TYPE_DIRECT = "DIRECT" // Direct address and port
	MCP_ENDPOINT_TYPE_REF    = "REF"    // Reference to Nacos service

	// MCP status
	MCP_STATUS_ACTIVE     = "active"
	MCP_STATUS_DEPRECATED = "deprecated"

	// A2A endpoint types
	A2A_ENDPOINT_TYPE_URL     = "URL"     // Direct URL
	A2A_ENDPOINT_TYPE_SERVICE = "SERVICE" // Reference to Nacos service

	// A2A default transport
	A2A_ENDPOINT_DEFAULT_TRANSPORT = "JSONRPC"

	// AI request type names
	QUERY_MCP_SERVER_REQUEST_NAME    = "QueryMcpServerRequest"
	RELEASE_MCP_SERVER_REQUEST_NAME  = "ReleaseMcpServerRequest"
	MCP_SERVER_ENDPOINT_REQUEST_NAME = "McpServerEndpointRequest"
	QUERY_AGENT_CARD_REQUEST_NAME    = "QueryAgentCardRequest"
	RELEASE_AGENT_CARD_REQUEST_NAME  = "ReleaseAgentCardRequest"
	AGENT_ENDPOINT_REQUEST_NAME      = "AgentEndpointRequest"

	// Default update interval for subscription cache (in seconds)
	AI_SUBSCRIPTION_UPDATE_INTERVAL = 10
)
