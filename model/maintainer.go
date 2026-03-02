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

package model

// RestResult is the unified response format for Nacos v3 Admin APIs.
type RestResult[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// Page is a generic paginated result.
type Page[T any] struct {
	TotalCount     int `json:"totalCount"`
	PageNumber     int `json:"pageNumber"`
	PagesAvailable int `json:"pagesAvailable"`
	PageItems      []T `json:"pageItems"`
}

// ======================== Core Models ========================

// Namespace represents a Nacos namespace.
type Namespace struct {
	Namespace         string `json:"namespace"`
	NamespaceShowName string `json:"namespaceShowName"`
	NamespaceDesc     string `json:"namespaceDesc"`
	Quota             int    `json:"quota"`
	ConfigCount       int    `json:"configCount"`
	Type              int    `json:"type"`
}

// NacosMember represents a cluster node.
type NacosMember struct {
	Ip         string                 `json:"ip"`
	Port       int                    `json:"port"`
	State      string                 `json:"state"`
	Address    string                 `json:"address"`
	ExtendInfo map[string]interface{} `json:"extendInfo"`
}

// ConnectionInfo represents a client connection to the server.
type ConnectionInfo struct {
	Traced       bool                   `json:"traced"`
	AbilityTable map[string]bool        `json:"abilityTable"`
	MetaInfo     ConnectionMetaInfo     `json:"metaInfo"`
}

// ConnectionMetaInfo represents metadata of a connection.
type ConnectionMetaInfo struct {
	ConnectType    string            `json:"connectType"`
	ClientIp       string            `json:"clientIp"`
	RemoteIp       string            `json:"remoteIp"`
	RemotePort     int               `json:"remotePort"`
	LocalPort      int               `json:"localPort"`
	Version        string            `json:"version"`
	ConnectionId   string            `json:"connectionId"`
	CreateTime     string            `json:"createTime"`
	LastActiveTime int64             `json:"lastActiveTime"`
	AppName        string            `json:"appName"`
	NamespaceId    string            `json:"namespaceId"`
	Labels         map[string]string `json:"labels"`
}

// ServerLoaderMetrics represents cluster loader metrics.
type ServerLoaderMetrics struct {
	Detail      []ServerLoaderMetric `json:"detail"`
	MemberCount int                  `json:"memberCount"`
	MetricsCount int                 `json:"metricsCount"`
	Completed   bool                 `json:"completed"`
	Max         int                  `json:"max"`
	Min         int                  `json:"min"`
	Avg         int                  `json:"avg"`
	Threshold   string               `json:"threshold"`
	Total       int                  `json:"total"`
}

// ServerLoaderMetric represents a single member's loader metric.
type ServerLoaderMetric struct {
	Address     string `json:"address"`
	SdkConCount int    `json:"sdkConCount"`
	ConCount    int    `json:"conCount"`
	Load        string `json:"load"`
	Cpu         string `json:"cpu"`
}

// IdGeneratorInfo represents the health status of an ID generator.
type IdGeneratorInfo struct {
	Resource string                 `json:"resource"`
	Info     map[string]interface{} `json:"info"`
}

// ======================== Naming Models ========================

// ServiceDetailInfo represents detailed information of a service (admin view).
type ServiceDetailInfo struct {
	NamespaceId      string                 `json:"namespaceId"`
	GroupName        string                 `json:"groupName"`
	ServiceName      string                 `json:"serviceName"`
	ProtectThreshold float64                `json:"protectThreshold"`
	Metadata         map[string]string      `json:"metadata"`
	Ephemeral        bool                   `json:"ephemeral"`
	Selector         map[string]interface{} `json:"selector"`
	Clusters         map[string]ClusterInfo `json:"clusterMap"`
}

// ClusterInfo represents cluster metadata in a service.
type ClusterInfo struct {
	ClusterName             string                 `json:"clusterName"`
	HealthChecker           map[string]interface{} `json:"healthChecker"`
	HealthyCheckPort        int                    `json:"healthyCheckPort"`
	UseInstancePortForCheck bool                   `json:"useInstancePortForCheck"`
	Metadata                map[string]string      `json:"metadata"`
	Hosts                   []Instance             `json:"hosts"`
}

// ServiceView represents a summary of a service.
type ServiceView struct {
	Name         string `json:"name"`
	GroupName    string `json:"groupName"`
	ClusterCount int    `json:"clusterCount"`
	IpCount      int    `json:"ipCount"`
	HealthyCount int    `json:"healthyInstanceCount"`
	TriggerFlag  string `json:"triggerFlag"`
}

// SubscriberInfo represents a service subscriber.
type SubscriberInfo struct {
	NamespaceId string `json:"namespaceId"`
	GroupName   string `json:"groupName"`
	ServiceName string `json:"serviceName"`
	Ip          string `json:"ip"`
	Port        int    `json:"port"`
	Agent       string `json:"agent"`
	AppName     string `json:"appName"`
}

// InstanceMetadataBatchResult represents the result of batch metadata operations.
type InstanceMetadataBatchResult struct {
	Updated []string `json:"updated"`
}

// ClientSummaryInfo represents a naming client summary.
type ClientSummaryInfo struct {
	ClientId        string `json:"clientId"`
	IsEphemeral     bool   `json:"ephemeral"`
	LastUpdatedTime int64  `json:"lastUpdatedTime"`
	ClientType      string `json:"clientType"`
	ConnectType     string `json:"connectType"`
	AppName         string `json:"appName"`
	Version         string `json:"version"`
	ClientIp        string `json:"clientIp"`
	ClientPort      int    `json:"clientPort"`
}

// ClientServiceInfo represents a service published or subscribed by a client.
type ClientServiceInfo struct {
	NamespaceId    string              `json:"namespaceId"`
	GroupName      string              `json:"groupName"`
	ServiceName    string              `json:"serviceName"`
	PublisherInfo  *ClientPublisherInfo  `json:"publisherInfo,omitempty"`
	SubscriberInfo *ClientSubscriberInfo `json:"subscriberInfo,omitempty"`
}

// ClientPublisherInfo represents a client that published a service.
type ClientPublisherInfo struct {
	ClientId    string `json:"clientId"`
	Ip          string `json:"ip"`
	Port        int    `json:"port"`
	ClusterName string `json:"clusterName"`
}

// ClientSubscriberInfo represents a client that subscribed to a service.
type ClientSubscriberInfo struct {
	ClientId string `json:"clientId"`
	AppName  string `json:"appName"`
	Agent    string `json:"agent"`
	Address  string `json:"address"`
}

// MetricsInfo represents naming system metrics.
type MetricsInfo struct {
	Status                      string `json:"status"`
	ServiceCount                int    `json:"serviceCount"`
	InstanceCount               int    `json:"instanceCount"`
	SubscribeCount              int    `json:"subscribeCount"`
	ClientCount                 int    `json:"clientCount"`
	ConnectionBasedClientCount  int    `json:"connectionBasedClientCount"`
	EphemeralIpPortClientCount  int    `json:"ephemeralIpPortClientCount"`
	PersistentIpPortClientCount int    `json:"persistentIpPortClientCount"`
	ResponsibleClientCount      int    `json:"responsibleClientCount"`
}

// ======================== Config Models ========================

// ConfigDetailInfo represents detailed config information (admin view).
type ConfigDetailInfo struct {
	Id               string `json:"id"`
	NamespaceId      string `json:"namespaceId"`
	GroupName        string `json:"groupName"`
	DataId           string `json:"dataId"`
	Content          string `json:"content"`
	Md5              string `json:"md5"`
	Type             string `json:"type"`
	AppName          string `json:"appName"`
	Desc             string `json:"desc"`
	ConfigTags       string `json:"configTags"`
	EncryptedDataKey string `json:"encryptedDataKey"`
	CreateTime       int64  `json:"createTime"`
	ModifyTime       int64  `json:"modifyTime"`
	CreateUser       string `json:"createUser"`
	CreateIp         string `json:"createIp"`
}

// ConfigBasicInfo represents config summary information.
type ConfigBasicInfo struct {
	Id          string `json:"id"`
	NamespaceId string `json:"namespaceId"`
	GroupName   string `json:"groupName"`
	DataId      string `json:"dataId"`
	Md5         string `json:"md5"`
	Type        string `json:"type"`
	AppName     string `json:"appName"`
	Desc        string `json:"desc"`
	ConfigTags  string `json:"configTags"`
	CreateTime  int64  `json:"createTime"`
	ModifyTime  int64  `json:"modifyTime"`
}

// ConfigHistoryInfo represents a config history entry.
type ConfigHistoryInfo struct {
	Id          string `json:"id"`
	NamespaceId string `json:"namespaceId"`
	GroupName   string `json:"groupName"`
	DataId      string `json:"dataId"`
	Content     string `json:"content"`
	Md5         string `json:"md5"`
	Type        string `json:"type"`
	AppName     string `json:"appName"`
	Desc        string `json:"desc"`
	CreateTime  int64  `json:"createTime"`
	ModifyTime  int64  `json:"modifyTime"`
	SrcIp       string `json:"srcIp"`
	SrcUser     string `json:"srcUser"`
	OpType      string `json:"opType"`
	PublishType string `json:"publishType"`
}

// ConfigListenerInfo represents config listener query result.
type ConfigListenerInfo struct {
	ListenersGroupkeyStatus map[string]string `json:"listenersGroupkeyStatus"`
}

// ConfigGrayInfo represents gray/beta configuration information.
// Extends ConfigDetailInfo with gray-specific fields.
type ConfigGrayInfo struct {
	Id               string `json:"id"`
	NamespaceId      string `json:"namespaceId"`
	GroupName        string `json:"groupName"`
	DataId           string `json:"dataId"`
	Content          string `json:"content"`
	Md5              string `json:"md5"`
	Type             string `json:"type"`
	AppName          string `json:"appName"`
	Desc             string `json:"desc"`
	EncryptedDataKey string `json:"encryptedDataKey"`
	CreateTime       int64  `json:"createTime"`
	ModifyTime       int64  `json:"modifyTime"`
	CreateUser       string `json:"createUser"`
	CreateIp         string `json:"createIp"`
	GrayName         string `json:"grayName"`
	GrayRule         string `json:"grayRule"`
	ExtInfo          string `json:"extInfo"`
}

// ConfigCloneInfo represents config clone source info.
type ConfigCloneInfo struct {
	CfgId  int64  `json:"cfgId"`
	DataId string `json:"dataId"`
	Group  string `json:"group"`
}

// ======================== AI Models ========================

// AgentVersionDetail represents version detail of an A2A agent.
type AgentVersionDetail struct {
	Version   string `json:"version"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	IsLatest  bool   `json:"isLatest"`
}

// AgentCardVersionInfo represents an A2A agent card with version info.
type AgentCardVersionInfo struct {
	Name                   string               `json:"name"`
	Description            string               `json:"description"`
	Version                string               `json:"version"`
	LatestPublishedVersion string               `json:"latestPublishedVersion"`
	VersionDetails         []AgentVersionDetail `json:"versionDetails"`
	RegistrationType       string               `json:"registrationType"`
}
