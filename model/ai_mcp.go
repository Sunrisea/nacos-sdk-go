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

package model

import "github.com/nacos-group/nacos-sdk-go/v2/common/constant"

// McpCapability represents the capability of an MCP server
type McpCapability string

const (
	McpCapabilityTool     McpCapability = "TOOL"
	McpCapabilityPrompt   McpCapability = "PROMPT"
	McpCapabilityResource McpCapability = "RESOURCE"
)

// Input represents an input parameter definition for MCP tools
type Input struct {
	Description  string   `json:"description,omitempty"`
	IsRequired   bool     `json:"isRequired,omitempty"`
	Format       string   `json:"format,omitempty"`
	Value        string   `json:"value,omitempty"`
	IsSecret     bool     `json:"isSecret,omitempty"`
	DefaultValue string   `json:"defaultValue,omitempty"`
	Choices      []string `json:"choices,omitempty"`
}

// KeyValueInput represents a key-value pair input with variable support
type KeyValueInput struct {
	Input
	Name      string           `json:"name,omitempty"`
	Variables map[string]Input `json:"variables,omitempty"`
}

// Repository represents repository information for MCP server source code
type Repository struct {
	URL       string `json:"url,omitempty"`
	Source    string `json:"source,omitempty"`
	ID        string `json:"id,omitempty"`
	Subfolder string `json:"subfolder,omitempty"`
}

// Argument represents an argument for MCP tool execution
type Argument struct {
	Type       string `json:"type,omitempty"` // "positional" or "named"
	ValueHint  string `json:"valueHint,omitempty"`
	Name       string `json:"name,omitempty"`
	IsRepeated bool   `json:"isRepeated,omitempty"`
}

// Package represents package definition for MCP server deployment
type Package struct {
	RegistryType         string          `json:"registryType,omitempty"`
	RegistryBaseURL      string          `json:"registryBaseUrl,omitempty"`
	Identifier           string          `json:"identifier,omitempty"`
	Version              string          `json:"version,omitempty"`
	FileSha256           string          `json:"fileSha256,omitempty"`
	RuntimeHint          string          `json:"runtimeHint,omitempty"`
	RuntimeArguments     []Argument      `json:"runtimeArguments,omitempty"`
	PackageArguments     []Argument      `json:"packageArguments,omitempty"`
	EnvironmentVariables []KeyValueInput `json:"environmentVariables,omitempty"`
}

// ServerVersionDetail represents version details for MCP server
type ServerVersionDetail struct {
	Version     string `json:"version,omitempty"`
	ReleaseDate string `json:"releaseDate,omitempty"`
	IsLatest    bool   `json:"isLatest,omitempty"`
}

// McpServiceRef represents a reference to an MCP service registered in Nacos
type McpServiceRef struct {
	NamespaceID       string `json:"namespaceId,omitempty"`
	GroupName         string `json:"groupName,omitempty"`
	ServiceName       string `json:"serviceName,omitempty"`
	TransportProtocol string `json:"transportProtocol,omitempty"`
}

// FrontEndpointConfig represents frontend endpoint configuration
type FrontEndpointConfig struct {
	Type         string          `json:"type,omitempty"`
	Protocol     string          `json:"protocol,omitempty"`
	EndpointType string          `json:"endpointType,omitempty"`
	EndpointData interface{}     `json:"endpointData,omitempty"`
	Path         string          `json:"path,omitempty"`
	Headers      []KeyValueInput `json:"headers,omitempty"`
}

// McpServerRemoteServiceConfig represents configuration for MCP server remote service access
type McpServerRemoteServiceConfig struct {
	ServiceRef              *McpServiceRef        `json:"serviceRef,omitempty"`
	ExportPath              string                `json:"exportPath,omitempty"`
	FrontEndpointConfigList []FrontEndpointConfig `json:"frontEndpointConfigList,omitempty"`
}

// McpServerBasicInfo represents basic information about an MCP server
type McpServerBasicInfo struct {
	ID                 string                        `json:"id,omitempty"`
	Name               string                        `json:"name,omitempty"`
	Protocol           string                        `json:"protocol,omitempty"`
	FrontProtocol      string                        `json:"frontProtocol,omitempty"`
	Description        string                        `json:"description,omitempty"`
	Repository         *Repository                   `json:"repository,omitempty"`
	Packages           []Package                     `json:"packages,omitempty"`
	VersionDetail      *ServerVersionDetail          `json:"versionDetail,omitempty"`
	Version            string                        `json:"version,omitempty"`
	RemoteServerConfig *McpServerRemoteServiceConfig `json:"remoteServerConfig,omitempty"`
	LocalServerConfig  map[string]interface{}        `json:"localServerConfig,omitempty"`
	Enabled            bool                          `json:"enabled,omitempty"`
	Status             string                        `json:"status,omitempty"`
	Capabilities       []McpCapability               `json:"capabilities,omitempty"`
}

// NewMcpServerBasicInfo creates a new McpServerBasicInfo with default values
func NewMcpServerBasicInfo() *McpServerBasicInfo {
	return &McpServerBasicInfo{
		Enabled: true,
		Status:  constant.MCP_STATUS_ACTIVE,
	}
}

// McpEndpointInfo represents endpoint information for accessing MCP server
type McpEndpointInfo struct {
	Protocol string          `json:"protocol,omitempty"`
	Address  string          `json:"address,omitempty"`
	Port     int             `json:"port,omitempty"`
	Path     string          `json:"path,omitempty"`
	Headers  []KeyValueInput `json:"headers,omitempty"`
}

// McpTool represents definition of a tool provided by MCP server
type McpTool struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
}

// McpToolMeta represents metadata for MCP tool configuration
type McpToolMeta struct {
	InvokeContext map[string]string      `json:"invokeContext,omitempty"`
	Enabled       bool                   `json:"enabled,omitempty"`
	Templates     map[string]interface{} `json:"templates,omitempty"`
}

// SecuritySchema represents security schema definition for MCP tool access
type SecuritySchema struct {
	ID                string `json:"id,omitempty"`
	Type              string `json:"type,omitempty"`
	Schema            string `json:"schema,omitempty"`
	In                string `json:"in,omitempty"`
	Name              string `json:"name,omitempty"`
	DefaultCredential string `json:"defaultCredential,omitempty"`
}

// EncryptObject represents encrypted data object
type EncryptObject struct {
	Data        string            `json:"data,omitempty"`
	EncryptInfo map[string]string `json:"encryptInfo,omitempty"`
}

// McpToolSpecification represents complete specification of tools provided by MCP server
type McpToolSpecification struct {
	SpecificationType string                 `json:"specificationType,omitempty"`
	EncryptData       *EncryptObject         `json:"encryptData,omitempty"`
	Tools             []McpTool              `json:"tools,omitempty"`
	ToolsMeta         map[string]McpToolMeta `json:"toolsMeta,omitempty"`
	SecuritySchema    []SecuritySchema       `json:"securitySchema,omitempty"`
}

// McpServerDetailInfo represents detailed information about an MCP server
type McpServerDetailInfo struct {
	McpServerBasicInfo
	BackendEndpoints  []McpEndpointInfo     `json:"backendEndpoints,omitempty"`
	FrontendEndpoints []McpEndpointInfo     `json:"frontendEndpoints,omitempty"`
	ToolSpec          *McpToolSpecification `json:"toolSpec,omitempty"`
	AllVersions       []ServerVersionDetail `json:"allVersions,omitempty"`
	NamespaceID       string                `json:"namespaceId,omitempty"`
}

// McpEndpointSpec represents specification for MCP server endpoint configuration
type McpEndpointSpec struct {
	Type string            `json:"type,omitempty"` // DIRECT or REF
	Data map[string]string `json:"data,omitempty"`
}
