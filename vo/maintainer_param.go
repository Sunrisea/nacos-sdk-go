/*
 * Copyright 1999-2025 Alibaba Group Holding Ltd.
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

import "github.com/nacos-group/nacos-sdk-go/v2/model"

// ======================== Core Params ========================

// CreateNamespaceParam represents parameters for creating a namespace.
type CreateNamespaceParam struct {
	NamespaceId   string `param:"namespaceId"`
	NamespaceName string `param:"namespaceName"`
	NamespaceDesc string `param:"namespaceDesc"`
}

// UpdateNamespaceParam represents parameters for updating a namespace.
type UpdateNamespaceParam struct {
	NamespaceId   string `param:"namespaceId"`
	NamespaceName string `param:"namespaceName"`
	NamespaceDesc string `param:"namespaceDesc"`
}

// ======================== Naming - Service Params ========================

// MaintainerServiceParam represents parameters for service CRUD operations.
type MaintainerServiceParam struct {
	NamespaceId      string            `param:"namespaceId"`
	GroupName        string            `param:"groupName"`
	ServiceName      string            `param:"serviceName"`
	Ephemeral        bool              `param:"ephemeral"`
	ProtectThreshold float64           `param:"protectThreshold"`
	Metadata         map[string]string `param:"metadata"`
	Selector         string            `param:"selector"`
}

// ListServicesParam represents parameters for listing services.
type ListServicesParam struct {
	NamespaceId        string `param:"namespaceId"`
	GroupNameParam     string `param:"groupNameParam"`
	ServiceNameParam   string `param:"serviceNameParam"`
	IgnoreEmptyService bool   `param:"ignoreEmptyService"`
	PageNo             int    `param:"pageNo"`
	PageSize           int    `param:"pageSize"`
}

// ListServicesDetailParam represents parameters for listing services with detail.
type ListServicesDetailParam struct {
	NamespaceId      string `param:"namespaceId"`
	GroupNameParam   string `param:"groupNameParam"`
	ServiceNameParam string `param:"serviceNameParam"`
	PageNo           int    `param:"pageNo"`
	PageSize         int    `param:"pageSize"`
}

// GetSubscribersParam represents parameters for getting service subscribers.
type GetSubscribersParam struct {
	NamespaceId string `param:"namespaceId"`
	GroupName   string `param:"groupName"`
	ServiceName string `param:"serviceName"`
	PageNo      int    `param:"pageNo"`
	PageSize    int    `param:"pageSize"`
	Aggregation bool   `param:"aggregation"`
}

// ======================== Naming - Instance Params ========================

// MaintainerInstanceParam represents parameters for instance CRUD operations.
type MaintainerInstanceParam struct {
	NamespaceId string            `param:"namespaceId"`
	GroupName   string            `param:"groupName"`
	ServiceName string            `param:"serviceName"`
	Ip          string            `param:"ip"`
	Port        int               `param:"port"`
	ClusterName string            `param:"clusterName"`
	Weight      float64           `param:"weight"`
	Healthy     bool              `param:"healthy"`
	Enabled     bool              `param:"enabled"`
	Ephemeral   bool              `param:"ephemeral"`
	Metadata    map[string]string `param:"metadata"`
}

// BatchInstanceMetadataParam represents parameters for batch instance metadata operations.
type BatchInstanceMetadataParam struct {
	NamespaceId     string            `param:"namespaceId"`
	GroupName       string            `param:"groupName"`
	ServiceName     string            `param:"serviceName"`
	ConsistencyType string            `param:"consistencyType"`
	Instances       string            `param:"instances"`
	Metadata        map[string]string `param:"metadata"`
}

// ListInstancesParam represents parameters for listing instances.
type ListInstancesParam struct {
	NamespaceId string `param:"namespaceId"`
	GroupName   string `param:"groupName"`
	ServiceName string `param:"serviceName"`
	ClusterName string `param:"clusterName"`
	HealthyOnly bool   `param:"healthyOnly"`
}

// GetInstanceDetailParam represents parameters for getting instance detail.
type GetInstanceDetailParam struct {
	NamespaceId string `param:"namespaceId"`
	GroupName   string `param:"groupName"`
	ServiceName string `param:"serviceName"`
	Ip          string `param:"ip"`
	Port        int    `param:"port"`
	ClusterName string `param:"clusterName"`
}

// ======================== Naming - Client Params ========================

// GetServiceClientParam represents parameters for querying clients of a service.
type GetServiceClientParam struct {
	NamespaceId string `param:"namespaceId"`
	GroupName   string `param:"groupName"`
	ServiceName string `param:"serviceName"`
	Ip          string `param:"ip"`
	Port        int    `param:"port"`
}

// ======================== Naming - Health Params ========================

// UpdateInstanceHealthParam represents parameters for updating instance health status.
type UpdateInstanceHealthParam struct {
	NamespaceId string `param:"namespaceId"`
	GroupName   string `param:"groupName"`
	ServiceName string `param:"serviceName"`
	Ip          string `param:"ip"`
	Port        int    `param:"port"`
	ClusterName string `param:"clusterName"`
	Healthy     bool   `param:"healthy"`
}

// UpdateClusterParam represents parameters for updating cluster metadata.
type UpdateClusterParam struct {
	NamespaceId          string            `param:"namespaceId"`
	GroupName            string            `param:"groupName"`
	ServiceName          string            `param:"serviceName"`
	ClusterName          string            `param:"clusterName"`
	HealthChecker        string            `param:"healthChecker"`
	CheckPort            int               `param:"checkPort"`
	UseInstancePort4Check bool             `param:"useInstancePort4Check"`
	Metadata             map[string]string `param:"metadata"`
}

// ======================== Config Params ========================

// MaintainerConfigParam represents parameters for config get/delete operations.
type MaintainerConfigParam struct {
	DataId      string `param:"dataId"`
	Group       string `param:"groupName"`
	NamespaceId string `param:"namespaceId"`
}

// MaintainerPublishConfigParam represents parameters for publishing a config.
type MaintainerPublishConfigParam struct {
	DataId           string `param:"dataId"`
	Group            string `param:"groupName"`
	NamespaceId      string `param:"namespaceId"`
	Content          string `param:"content"`
	AppName          string `param:"appName"`
	SrcUser          string `param:"srcUser"`
	ConfigTags       string `param:"configTags"`
	Desc             string `param:"desc"`
	Type             string `param:"type"`
	EncryptedDataKey string `param:"encryptedDataKey"`
}

// UpdateConfigMetadataParam represents parameters for updating config metadata.
type UpdateConfigMetadataParam struct {
	DataId      string `param:"dataId"`
	Group       string `param:"groupName"`
	NamespaceId string `param:"namespaceId"`
	Desc        string `param:"desc"`
	ConfigTags  string `param:"configTags"`
}

// MaintainerSearchConfigParam represents parameters for searching/listing configs via admin API.
type MaintainerSearchConfigParam struct {
	DataId       string `param:"dataId"`
	Group        string `param:"groupName"`
	NamespaceId  string `param:"namespaceId"`
	Search       string `param:"search"`
	ConfigDetail string `param:"configDetail"`
	Type         string `param:"type"`
	ConfigTags   string `param:"configTags"`
	AppName      string `param:"appName"`
	PageNo       int    `param:"pageNo"`
	PageSize     int    `param:"pageSize"`
}

// GetListenersParam represents parameters for getting config listeners.
type GetListenersParam struct {
	DataId      string `param:"dataId"`
	Group       string `param:"groupName"`
	NamespaceId string `param:"namespaceId"`
	Aggregation bool   `param:"aggregation"`
}

// GetSubClientConfigParam represents parameters for getting subscribed client configs by IP.
type GetSubClientConfigParam struct {
	Ip          string `param:"ip"`
	All         bool   `param:"all"`
	NamespaceId string `param:"namespaceId"`
	Aggregation bool   `param:"aggregation"`
}

// CloneConfigParam represents parameters for cloning configs (JSON body).
type CloneConfigParam struct {
	NamespaceId string              `json:"namespaceId"`
	CloneInfos  []model.ConfigCloneInfo `json:"ids"`
	SrcUser     string              `json:"srcUser"`
	Policy      string              `json:"policy"`
}

// PublishBetaConfigParam represents parameters for publishing a beta config.
type PublishBetaConfigParam struct {
	DataId      string `param:"dataId"`
	Group       string `param:"groupName"`
	NamespaceId string `param:"namespaceId"`
	Content     string `param:"content"`
	AppName     string `param:"appName"`
	SrcUser     string `param:"srcUser"`
	ConfigTags  string `param:"configTags"`
	Desc        string `param:"desc"`
	Type        string `param:"type"`
	BetaIps     string `param:"betaIps"`
}

// StopBetaParam represents parameters for stopping beta config.
type StopBetaParam struct {
	DataId      string `param:"dataId"`
	Group       string `param:"groupName"`
	NamespaceId string `param:"namespaceId"`
}

// QueryBetaParam is an alias for StopBetaParam.
type QueryBetaParam = StopBetaParam

// ListConfigHistoryParam represents parameters for listing config history.
type ListConfigHistoryParam struct {
	DataId      string `param:"dataId"`
	Group       string `param:"groupName"`
	NamespaceId string `param:"namespaceId"`
	PageNo      int    `param:"pageNo"`
	PageSize    int    `param:"pageSize"`
}

// GetConfigHistoryParam represents parameters for getting a specific config history entry.
type GetConfigHistoryParam struct {
	DataId      string `param:"dataId"`
	Group       string `param:"groupName"`
	NamespaceId string `param:"namespaceId"`
	Nid         int64  `param:"nid"`
}

// ======================== AI - MCP Params ========================

// ListMcpServerParam represents parameters for listing MCP servers.
type ListMcpServerParam struct {
	NamespaceId string `param:"namespaceId"`
	McpName     string `param:"mcpName"`
	PageNo      int    `param:"pageNo"`
	PageSize    int    `param:"pageSize"`
}

// SearchMcpServerParam is an alias for ListMcpServerParam.
type SearchMcpServerParam = ListMcpServerParam

// GetMcpServerDetailParam represents parameters for getting MCP server detail.
type GetMcpServerDetailParam struct {
	NamespaceId string `param:"namespaceId"`
	McpName     string `param:"mcpName"`
	McpId       string `param:"mcpId"`
	Version     string `param:"version"`
}

// CreateMcpServerParam represents parameters for creating an MCP server (JSON body).
type CreateMcpServerParam struct {
	NamespaceId  string      `json:"namespaceId"`
	McpName      string      `json:"mcpName"`
	ServerSpec   interface{} `json:"serverSpec"`
	ToolSpec     interface{} `json:"toolSpec,omitempty"`
	EndpointSpec interface{} `json:"endpointSpec,omitempty"`
}

// UpdateMcpServerParam represents parameters for updating an MCP server (JSON body).
type UpdateMcpServerParam struct {
	NamespaceId      string      `json:"namespaceId"`
	McpName          string      `json:"mcpName"`
	IsLatest         bool        `json:"isLatest"`
	ServerSpec       interface{} `json:"serverSpec"`
	ToolSpec         interface{} `json:"toolSpec,omitempty"`
	EndpointSpec     interface{} `json:"endpointSpec,omitempty"`
	OverrideExisting bool        `json:"overrideExisting"`
}

// DeleteMcpServerParam represents parameters for deleting an MCP server.
type DeleteMcpServerParam struct {
	NamespaceId string `param:"namespaceId"`
	McpName     string `param:"mcpName"`
	McpId       string `param:"mcpId"`
	Version     string `param:"version"`
}

// ======================== AI - A2A Params ========================

// RegisterAgentParam represents parameters for registering an agent.
type RegisterAgentParam struct {
	AgentCard        interface{} `json:"agentCard"`
	NamespaceId      string      `param:"namespaceId"`
	AgentName        string      `param:"agentName"`
	RegistrationType string      `param:"registrationType"`
}

// GetMaintainerAgentCardParam represents parameters for getting an agent card.
type GetMaintainerAgentCardParam struct {
	AgentName        string `param:"agentName"`
	NamespaceId      string `param:"namespaceId"`
	RegistrationType string `param:"registrationType"`
	Version          string `param:"version"`
}

// UpdateAgentCardParam represents parameters for updating an agent card.
type UpdateAgentCardParam struct {
	AgentCard        interface{} `json:"agentCard"`
	NamespaceId      string      `param:"namespaceId"`
	AgentName        string      `param:"agentName"`
	SetAsLatest      bool        `param:"setAsLatest"`
	RegistrationType string      `param:"registrationType"`
}

// DeleteAgentParam represents parameters for deleting an agent.
type DeleteAgentParam struct {
	AgentName   string `param:"agentName"`
	NamespaceId string `param:"namespaceId"`
	Version     string `param:"version"`
}

// ListAgentVersionParam represents parameters for listing all versions of an agent.
type ListAgentVersionParam struct {
	AgentName   string `param:"agentName"`
	NamespaceId string `param:"namespaceId"`
}

// SearchAgentParam represents parameters for searching agents by name.
type SearchAgentParam struct {
	AgentName   string `param:"agentName"`
	NamespaceId string `param:"namespaceId"`
	PageNo      int    `param:"pageNo"`
	PageSize    int    `param:"pageSize"`
}

// ListAgentCardsParam represents parameters for listing agent cards.
type ListAgentCardsParam struct {
	NamespaceId string `param:"namespaceId"`
	PageNo      int    `param:"pageNo"`
	PageSize    int    `param:"pageSize"`
}
