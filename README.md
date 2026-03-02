# Nacos-sdk-go [中文](./README_CN.md) #

[![Build Status](https://travis-ci.org/nacos-group/nacos-sdk-go.svg?branch=master)](https://travis-ci.org/nacos-group/nacos-sdk-go) [![Go Report Card](https://goreportcard.com/badge/github.com/nacos-group/nacos-sdk-go)](https://goreportcard.com/report/github.com/nacos-group/nacos-sdk-go) ![license](https://img.shields.io/badge/license-Apache--2.0-green.svg)

---

## Nacos-sdk-go

Nacos-sdk-go for Go client allows you to access Nacos service. It supports service discovery, dynamic configuration, **AI service management (MCP Server & A2A Agent)**, and **Maintainer operations (admin/control plane)**.

## What's New in 3.x

- **AI Module**: Support for MCP (Model Context Protocol) Server and A2A (Agent-to-Agent) Agent management
- **MCP Server**: Register, discover, and subscribe to MCP servers for AI tool integration
- **A2A Agent**: Manage AI agent cards with skills, endpoints, and version control
- **Maintainer Client**: Admin/control plane client for server management via `/v3/admin/` HTTP APIs, covering core operations, config management, naming management, and AI management
- **Redo Mechanism**: Enhanced connection recovery with automatic data re-synchronization

## Requirements

Supported Go version over 1.24

Supported Nacos version over 3.x

## Installation

Use `go get` to install SDK：

```sh
$ go get -u github.com/nacos-group/nacos-sdk-go/v2
```

## Quick Examples

* ClientConfig

```go

constant.ClientConfig {
	TimeoutMs   uint64 // timeout for requesting Nacos server, default value is 10000ms
	NamespaceId string // the namespaceId of Nacos
	Endpoint    string // the endpoint for ACM. https://help.aliyun.com/document_detail/130146.html
	RegionId    string // the regionId for ACM & KMS
	AccessKey   string // the AccessKey for ACM & KMS
	SecretKey   string // the SecretKey for ACM & KMS
	OpenKMS     bool   // it's to open KMS, default is false. https://help.aliyun.com/product/28933.html
	// , to enable encrypt/decrypt, DataId should be start with "cipher-"
	CacheDir             string // the directory for persist nacos service info,default value is current path
	UpdateThreadNum      int    // the number of goroutine for update nacos service info,default value is 20
	NotLoadCacheAtStart  bool   // not to load persistent nacos service info in CacheDir at start time
	UpdateCacheWhenEmpty bool   // update cache when get empty service instance from server
	Username             string // the username for nacos auth
	Password             string // the password for nacos auth
	LogDir               string // the directory for log, default is current path
	RotateTime           string // the rotate time for log, eg: 30m, 1h, 24h, default is 24h
	MaxAge               int64  // the max age of a log file, default value is 3
	LogLevel             string // the level of log, it's must be debug,info,warn,error, default value is info
}

```

* ServerConfig

```go

constant.ServerConfig{
    Scheme      string // the nacos server scheme, default=http, this is not required in 2.0+
    ContextPath string // the nacos server contextpath, default=/nacos, this is not required in 2.0+
    IpAddr      string // the nacos server address 
    Port        uint64 // nacos server port
    GrpcPort    uint64 // nacos server grpc port, default=server port + 1000, this is not required
}

```

<b>Note: We can config multiple ServerConfig, the client will rotate request the servers</b>

### Create Client

```go

// create clientConfig
clientConfig := constant.ClientConfig{
    NamespaceId:         "e525eafa-f7d7-4029-83d9-008937f9d468", // we can create multiple clients with different namespaceId to support multiple namespace. When namespace is public, fill in the blank string here.
    TimeoutMs:           5000,
    NotLoadCacheAtStart: true,
    LogDir:              "/tmp/nacos/log",
    CacheDir:            "/tmp/nacos/cache",
    LogLevel:            "debug",
}

// Another way of create clientConfig
clientConfig := *constant.NewClientConfig(
    constant.WithNamespaceId("e525eafa-f7d7-4029-83d9-008937f9d468"), // When namespace is public, fill in the blank string here.
    constant.WithTimeoutMs(5000),
    constant.WithNotLoadCacheAtStart(true),
    constant.WithLogDir("/tmp/nacos/log"),
    constant.WithCacheDir("/tmp/nacos/cache"),
    constant.WithLogLevel("debug"),
)

// At least one ServerConfig
serverConfigs := []constant.ServerConfig{
    *constant.NewServerConfig(
        "console1.nacos.io",
        8848,
        constant.WithScheme("http"),
        constant.WithContextPath("/nacos"),
    ),
}

// Create naming client for service discovery
namingClient, err := clients.NewNamingClient(
    vo.NacosClientParam{
        ClientConfig:  &clientConfig,
        ServerConfigs: serverConfigs,
    },
)

// Create config client for dynamic configuration
configClient, err := clients.NewConfigClient(
    vo.NacosClientParam{
        ClientConfig:  &clientConfig,
        ServerConfigs: serverConfigs,
    },
)

// Create AI client for MCP Server and Agent management (New in 3.x)
aiClient, err := clients.NewAIClient(
    vo.NacosClientParam{
        ClientConfig:  &clientConfig,
        ServerConfigs: serverConfigs,
    },
)
defer aiClient.CloseClient()

// Create maintainer clients for admin/control plane operations (New in 3.x)
namingMaintainerClient, err := clients.NewNamingMaintainerClient(
    vo.NacosClientParam{
        ClientConfig:  &clientConfig,
        ServerConfigs: serverConfigs,
    },
)
defer namingMaintainerClient.CloseClient()

configMaintainerClient, err := clients.NewConfigMaintainerClient(
    vo.NacosClientParam{
        ClientConfig:  &clientConfig,
        ServerConfigs: serverConfigs,
    },
)
defer configMaintainerClient.CloseClient()

aiMaintainerClient, err := clients.NewAiMaintainerClient(
    vo.NacosClientParam{
        ClientConfig:  &clientConfig,
        ServerConfigs: serverConfigs,
    },
)
defer aiMaintainerClient.CloseClient()

```

### Service Discovery

* Register instance: RegisterInstance

```go

success, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
    Ip:          "10.0.0.11",
    Port:        8848,
    ServiceName: "demo.go",
    Weight:      10,
    Enable:      true,
    Healthy:     true,
    Ephemeral:   true,
    Metadata:    map[string]string{"idc":"shanghai"},
    ClusterName: "cluster-a", // default value is DEFAULT
    GroupName:   "group-a",   // default value is DEFAULT_GROUP
})
   
```

* Deregister instance: DeregisterInstance

```go

success, err := namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
    Ip:          "10.0.0.11",
    Port:        8848,
    ServiceName: "demo.go",
    Ephemeral:   true,
    Cluster:     "cluster-a", // default value is DEFAULT
    GroupName:   "group-a",   // default value is DEFAULT_GROUP
})

```

* Get service: GetService

```go

services, err := namingClient.GetService(vo.GetServiceParam{
    ServiceName: "demo.go",
    Clusters:    []string{"cluster-a"}, // default value is DEFAULT
    GroupName:   "group-a",             // default value is DEFAULT_GROUP
})

```

* Get all instances: SelectAllInstances

```go

// SelectAllInstance return all instances, include healthy=false, enable=false, weight<=0
instances, err := namingClient.SelectAllInstances(vo.SelectAllInstancesParam{
    ServiceName: "demo.go",
    GroupName:   "group-a",             // default value is DEFAULT_GROUP
    Clusters:    []string{"cluster-a"}, // default value is DEFAULT
})

```

* Get instances: SelectInstances

```go

// SelectInstances only return the instances of healthy=${HealthyOnly}, enable=true and weight>0
instances, err := namingClient.SelectInstances(vo.SelectInstancesParam{
    ServiceName: "demo.go",
    GroupName:   "group-a",             // default value is DEFAULT_GROUP
    Clusters:    []string{"cluster-a"}, // default value is DEFAULT
    HealthyOnly: true,
})

```

* Get one healthy instance (WRR): SelectOneHealthyInstance

```go

// SelectOneHealthyInstance return one instance by WRR strategy for load balance
// And the instance should be health=true, enable=true and weight>0
instance, err := namingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
    ServiceName: "demo.go",
    GroupName:   "group-a",             // default value is DEFAULT_GROUP
    Clusters:    []string{"cluster-a"}, // default value is DEFAULT
})

```

* Listen service change event: Subscribe

```go

// Subscribe key = serviceName+groupName+cluster
// Note: We can add multiple SubscribeCallback with the same key.
err := namingClient.Subscribe(vo.SubscribeParam{
    ServiceName: "demo.go",
    GroupName:   "group-a",             // default value is DEFAULT_GROUP
    Clusters:    []string{"cluster-a"}, // default value is DEFAULT
    SubscribeCallback: func(services []model.Instance, err error) {
        log.Printf("\n\n callback return services:%s \n\n", utils.ToJsonString(services))
    },
})

```

* Cancel listen of service change event: Unsubscribe

```go

err := namingClient.Unsubscribe(vo.SubscribeParam{
    ServiceName: "demo.go",
    GroupName:   "group-a",             // default value is DEFAULT_GROUP
    Clusters:    []string{"cluster-a"}, // default value is DEFAULT
    SubscribeCallback: func(services []model.Instance, err error) {
        log.Printf("\n\n callback return services:%s \n\n", utils.ToJsonString(services))
    },
})

```

* Get all services name: GetAllServicesInfo

```go

serviceInfos, err := namingClient.GetAllServicesInfo(vo.GetAllServiceInfoParam{
    NameSpace: "0e83cc81-9d8c-4bb8-a28a-ff703187543f",
    PageNo:   1,
    PageSize: 10,
})

```

### Dynamic Configuration

* Publish config: PublishConfig

```go

success, err := configClient.PublishConfig(vo.ConfigParam{
    DataId:  "dataId",
    Group:   "group",
    Content: "hello world!",
})

```

* Delete config: DeleteConfig

```go

success, err = configClient.DeleteConfig(vo.ConfigParam{
    DataId: "dataId",
    Group:  "group",
})

```

* Get config info: GetConfig

```go

content, err := configClient.GetConfig(vo.ConfigParam{
    DataId: "dataId",
    Group:  "group",
})

```

* Listen config change event: ListenConfig

```go

err := configClient.ListenConfig(vo.ConfigParam{
    DataId: "dataId",
    Group:  "group",
    OnChange: func(namespace, group, dataId, data string) {
        fmt.Println("group:" + group + ", dataId:" + dataId + ", data:" + data)
    },
})

```

* Cancel the listening of config change event: CancelListenConfig

```go

err := configClient.CancelListenConfig(vo.ConfigParam{
    DataId: "dataId",
    Group:  "group",
})

```

* Search config: SearchConfig

```go

configPage, err := configClient.SearchConfig(vo.SearchConfigParam{
    Search:   "blur",
    DataId:   "",
    Group:    "",
    PageNo:   1,
    PageSize: 10,
})

```

### AI Service Management (New in 3.x)

The AI module provides management capabilities for MCP (Model Context Protocol) Servers and A2A (Agent-to-Agent) Agents.

#### MCP Server Operations

* Release (publish) MCP Server: ReleaseMcpServer

```go

serverSpec := model.NewMcpServerBasicInfo()
serverSpec.Name = "demo-mcp-server"
serverSpec.Description = "Demo MCP Server"
serverSpec.Protocol = "mcp-sse"
serverSpec.FrontProtocol = "mcp-sse"
serverSpec.VersionDetail = &model.ServerVersionDetail{
    Version: "v1.0.0",
}

toolSpec := &model.McpToolSpecification{
    Tools: []model.McpTool{
        {
            Name:        "search",
            Description: "Search tool for finding information",
        },
    },
}

endpointSpec := &model.McpEndpointSpec{
    Type: constant.MCP_ENDPOINT_TYPE_REF,
    Data: map[string]string{
        "groupName":   "DEFAULT_GROUP",
        "serviceName": "demo-mcp-service",
    },
}

mcpId, err := aiClient.ReleaseMcpServer(vo.ReleaseMcpServerParam{
    ServerSpec:      serverSpec,
    ToolSpec:        toolSpec,
    McpEndpointSpec: endpointSpec,
})

```

* Get MCP Server: GetMcpServer

```go

mcpServer, err := aiClient.GetMcpServer(vo.GetMcpServerParam{
    McpName: "demo-mcp-server",
    Version: "v1.0.0",
})

```

* Register MCP Server endpoint: RegisterMcpServerEndpoint

```go

err := aiClient.RegisterMcpServerEndpoint(vo.RegisterMcpServerEndpointParam{
    McpName: "demo-mcp-server",
    Address: "192.168.1.100",
    Port:    8080,
    Version: "v1.0.0",
})

```

* Deregister MCP Server endpoint: DeregisterMcpServerEndpoint

```go

err := aiClient.DeregisterMcpServerEndpoint(vo.DeregisterMcpServerEndpointParam{
    McpName: "demo-mcp-server",
    Address: "192.168.1.100",
    Port:    8080,
})

```

* Subscribe to MCP Server changes: SubscribeMcpServer

```go

mcpServer, err := aiClient.SubscribeMcpServer(vo.SubscribeMcpServerParam{
    McpName: "demo-mcp-server",
    Version: "v1.0.0",
    SubscribeCallback: func(mcpId, namespaceId, mcpName string, mcpServer model.McpServerDetailInfo) {
        fmt.Printf("MCP Server changed: %s, endpoints: %d\n", mcpName, len(mcpServer.BackendEndpoints))
    },
})

```

#### A2A Agent Operations

* Release (publish) Agent Card: ReleaseAgentCard

```go

agentCard := &a2a.AgentCard{
    Name:            "demo-agent",
    Version:         "v1.0.0",
    ProtocolVersion: "0.2.1",
    Description:     "Demo AI Agent",
    URL:             "https://demo-agent.example.com",
    Skills: []a2a.AgentSkill{
        {
            ID:          "search",
            Name:        "Search Skill",
            Description: "Can search for information",
        },
    },
}

err := aiClient.ReleaseAgentCard(vo.ReleaseAgentCardParam{
    AgentCard:        agentCard,
    RegistrationType: constant.A2A_ENDPOINT_TYPE_SERVICE,
    SetAsLatest:      true,
})

```

* Get Agent Card: GetAgentCard

```go

agentCardInfo, err := aiClient.GetAgentCard(vo.GetAgentCardParam{
    AgentName:        "demo-agent",
    Version:          "v1.0.0",
    RegistrationType: constant.A2A_ENDPOINT_TYPE_URL,
})

```

* Register Agent endpoint: RegisterAgentEndpoint

```go

err := aiClient.RegisterAgentEndpoint(vo.RegisterAgentEndpointParam{
    AgentName:  "demo-agent",
    Address:    "192.168.1.100",
    Port:       9090,
    Path:       "/api/v1/agent",
    Transport:  constant.A2A_ENDPOINT_DEFAULT_TRANSPORT,
    SupportTLS: true,
    Version:    "v1.0.0",
})

```

* Deregister Agent endpoint: DeregisterAgentEndpoint

```go

err := aiClient.DeregisterAgentEndpoint(vo.DeregisterAgentEndpointParam{
    AgentName: "demo-agent",
    Address:   "192.168.1.100",
    Port:      9090,
    Version:   "v1.0.0",
})

```

* Subscribe to Agent Card changes: SubscribeAgentCard

```go

agentCardInfo, err := aiClient.SubscribeAgentCard(vo.SubscribeAgentCardParam{
    AgentName: "demo-agent",
    Version:   "v1.0.0",
    SubscribeCallback: func(agentName string, agentCard model.AgentCardDetailInfo) {
        fmt.Printf("Agent Card changed: %s, latestVersion: %v\n", agentName, agentCard.LatestVersion)
    },
})

```

### Maintainer Operations (New in 3.x)

The Maintainer module provides admin/control plane operations for Nacos server management via HTTP Admin APIs. It offers three specialized clients that all inherit core management capabilities:

- **NamingMaintainerClient**: Core + Naming admin operations (service/instance/cluster management)
- **ConfigMaintainerClient**: Core + Config admin operations (config CRUD/history/beta/clone)
- **AiMaintainerClient**: Core + AI admin operations (MCP Server & A2A Agent management)

#### Core Operations (available in all maintainer clients)

```go

// Server health checks
state, err := client.GetServerState()
alive, err := client.Liveness()
ready, err := client.Readiness()

// Cluster management
nodes, err := client.ListClusterNodes("", "")
metrics, err := client.GetClusterLoaderMetrics()
conns, err := client.GetCurrentClients()

// Namespace CRUD
ok, err := client.CreateNamespace(vo.CreateNamespaceParam{
    NamespaceId:   "my-namespace",
    NamespaceName: "My Namespace",
    NamespaceDesc: "Description",
})
ns, err := client.GetNamespace("my-namespace")
ok, err = client.UpdateNamespace(vo.UpdateNamespaceParam{
    NamespaceId:   "my-namespace",
    NamespaceName: "Updated Name",
})
ok, err = client.DeleteNamespace("my-namespace")

```

#### Config Maintainer Operations

```go

// Config CRUD
ok, err := configMaintainerClient.PublishConfig(vo.MaintainerPublishConfigParam{
    DataId:  "app.properties",
    Group:   "DEFAULT_GROUP",
    Content: "server.port=8080",
    Type:    "properties",
})

cfg, err := configMaintainerClient.GetConfig(vo.MaintainerConfigParam{
    DataId: "app.properties",
    Group:  "DEFAULT_GROUP",
})

// Search configs
result, err := configMaintainerClient.SearchConfig(vo.MaintainerSearchConfigParam{
    Search: "blur", DataId: "app*", PageNo: 1, PageSize: 10,
})

// Config history
history, err := configMaintainerClient.ListConfigHistory(vo.ListConfigHistoryParam{
    DataId: "app.properties", Group: "DEFAULT_GROUP", PageNo: 1, PageSize: 10,
})

// Beta config
ok, err = configMaintainerClient.PublishBetaConfig(vo.PublishBetaConfigParam{
    DataId: "app.properties", Group: "DEFAULT_GROUP",
    Content: "beta.enabled=true", BetaIps: "10.0.0.1,10.0.0.2",
})

```

#### Naming Maintainer Operations

```go

// Service CRUD
result, err := namingMaintainerClient.CreateService(vo.MaintainerServiceParam{
    NamespaceId: "my-namespace", GroupName: "DEFAULT_GROUP",
    ServiceName: "my-service",   Ephemeral: false,
    ProtectThreshold: 0.5,
})

// Instance management
result, err = namingMaintainerClient.RegisterInstance(vo.MaintainerInstanceParam{
    NamespaceId: "my-namespace", GroupName: "DEFAULT_GROUP",
    ServiceName: "my-service",   Ip: "10.0.0.1", Port: 8080,
    Weight: 1.0, Healthy: true, Enabled: true,
})

instances, err := namingMaintainerClient.ListInstances(vo.ListInstancesParam{
    NamespaceId: "my-namespace", GroupName: "DEFAULT_GROUP",
    ServiceName: "my-service",
})

// Client info
clientList, err := namingMaintainerClient.GetClientList()
clientDetail, err := namingMaintainerClient.GetClientDetail(clientList[0])

// Metrics
metrics, err := namingMaintainerClient.GetMetrics(false)

```

#### AI Maintainer Operations

```go

// MCP Server CRUD
result, err := aiMaintainerClient.CreateMcpServer(vo.CreateMcpServerParam{
    McpName: "my-mcp-server",
    ServerSpec: map[string]interface{}{
        "name": "my-mcp-server", "protocol": "HTTP",
        "versionDetail": map[string]interface{}{"version": "1.0.0"},
    },
    EndpointSpec: map[string]interface{}{
        "type": "DIRECT",
        "data": map[string]string{"address": "127.0.0.1", "port": "9090"},
    },
})

mcpServers, err := aiMaintainerClient.ListMcpServer(vo.ListMcpServerParam{
    PageNo: 1, PageSize: 10,
})

// A2A Agent CRUD
ok, err := aiMaintainerClient.RegisterAgent(vo.RegisterAgentParam{
    AgentName: "my-agent",
    AgentCard: map[string]interface{}{
        "name": "my-agent", "version": "1.0.0",
        "protocolVersion": "0.2.0",
        "url": "http://localhost:8080/a2a",
        "preferredTransport": "httpPost",
    },
})

agents, err := aiMaintainerClient.ListAgentCards(vo.ListAgentCardsParam{
    PageNo: 1, PageSize: 10,
})

```

## Example

We can run example to learn how to use nacos go client.

* [Config Example](./example/config)
* [Naming Example](./example/service)
* [AI Example](./example/ai)
* [Maintainer Example](./example/maintainer)

## Documentation

You can view the open-api documentation from the [Nacos open-api website](https://nacos.io/en-us/docs/open-api.html).

You can view the full documentation from the [Nacos website](https://nacos.io/en-us/docs/what-is-nacos.html).

## Contributing

Contributors are welcomed to join Nacos-sdk-go project. Please check [CONTRIBUTING.md](./CONTRIBUTING.md) about how to contribute to this project.

## Contact

* Join us from DingDing Group (23191211).
* [Gitter](https://gitter.im/alibaba/nacos): Nacos's IM tool for community messaging, collaboration and discovery.
* [Twitter](https://twitter.com/nacos2): Follow along for latest nacos news on Twitter.
* [Weibo](https://weibo.com/u/6574374908): Follow along for latest nacos news on Weibo (Twitter of China version).
* [Nacos SegmentFault](https://segmentfault.com/t/nacos): Get the latest notice and prompt help from SegmentFault.
* Email Group:
    * users-nacos@googlegroups.com: Nacos usage general discussion.
    * dev-nacos@googlegroups.com: Nacos developer discussion (APIs, feature design, etc).
    * commits-nacos@googlegroups.com: Commits notice, very high frequency.
