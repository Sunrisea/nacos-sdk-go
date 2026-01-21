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

package rpc

import "errors"

// Errors related to ability negotiation
var (
	ErrAbilityNegotiationTimeout = errors.New("ability negotiation timeout")
	ErrServerNotSupported        = errors.New("server version too low, feature not supported")
)

// AbilityMode represents the mode of an ability (server or client side)
type AbilityMode int

const (
	// AbilityModeServer indicates the ability is for server side
	AbilityModeServer AbilityMode = iota
	// AbilityModeSDKClient indicates the ability is for SDK client side
	AbilityModeSDKClient
)

// AbilityStatus represents the status of an ability
type AbilityStatus int

const (
	// AbilityStatusUnknown indicates the ability status is unknown
	AbilityStatusUnknown AbilityStatus = iota
	// AbilityStatusSupported indicates the ability is supported
	AbilityStatusSupported
	// AbilityStatusNotSupported indicates the ability is not supported
	AbilityStatusNotSupported
)

// AbilityKey represents a capability that can be negotiated between client and server
type AbilityKey struct {
	KeyName     string
	Description string
	Mode        AbilityMode
}

// Server-side abilities
var (
	// ServerPersistentInstanceByGrpc indicates server supports persistent instance by gRPC
	ServerPersistentInstanceByGrpc = AbilityKey{
		KeyName:     "supportPersistentInstanceByGrpc",
		Description: "Server whether support persistent instance by gRPC",
		Mode:        AbilityModeServer,
	}

	// ServerFuzzyWatch indicates server supports fuzzy watch
	ServerFuzzyWatch = AbilityKey{
		KeyName:     "fuzzyWatch",
		Description: "Server whether support fuzzy watch",
		Mode:        AbilityModeServer,
	}

	// ServerDistributedLock indicates server supports distributed lock
	ServerDistributedLock = AbilityKey{
		KeyName:     "lock",
		Description: "Server whether support distributed lock",
		Mode:        AbilityModeServer,
	}

	// ServerMcpRegistry indicates server supports MCP server registry
	ServerMcpRegistry = AbilityKey{
		KeyName:     "mcp",
		Description: "Server whether support release mcp server and register endpoint for mcp server",
		Mode:        AbilityModeServer,
	}

	// ServerAgentRegistry indicates server supports Agent registry
	ServerAgentRegistry = AbilityKey{
		KeyName:     "agent",
		Description: "Server whether support release agent server and register endpoint for agent server",
		Mode:        AbilityModeServer,
	}
)

// Client-side abilities
var (
	// SDKMcpRegistry indicates SDK client supports MCP registry
	SDKMcpRegistry = AbilityKey{
		KeyName:     "mcp",
		Description: "Client whether support release mcp server and register endpoint for mcp server",
		Mode:        AbilityModeSDKClient,
	}

	// SDKAgentRegistry indicates SDK client supports Agent registry
	SDKAgentRegistry = AbilityKey{
		KeyName:     "agent",
		Description: "Client whether support release agent server and register endpoint for agent server",
		Mode:        AbilityModeSDKClient,
	}
)

// SDKAbilityTable contains the abilities that the SDK client supports
var SDKAbilityTable = map[string]bool{
	SDKMcpRegistry.KeyName:   true,
	SDKAgentRegistry.KeyName: true,
}
