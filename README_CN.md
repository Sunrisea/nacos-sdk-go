# Nacos-sdk-go [English](./README.md) #

[![Build Status](https://travis-ci.org/nacos-group/nacos-sdk-go.svg?branch=master)](https://travis-ci.org/nacos-group/nacos-sdk-go) [![Go Report Card](https://goreportcard.com/badge/github.com/nacos-group/nacos-sdk-go)](https://goreportcard.com/report/github.com/nacos-group/nacos-sdk-go) ![license](https://img.shields.io/badge/license-Apache--2.0-green.svg)

---

## Nacos-sdk-go

Nacos-sdk-go是Nacos的Go语言客户端，它实现了服务发现、动态配置和 **AI服务管理（MCP Server 和 A2A Agent）** 的功能。

## 3.x 版本新特性

- **AI 模块**: 支持 MCP (Model Context Protocol) Server 和 A2A (Agent-to-Agent) Agent 管理
- **MCP Server**: 注册、发现和订阅 MCP 服务器，用于 AI 工具集成
- **A2A Agent**: 管理 AI Agent 卡片，包含技能、端点和版本控制
- **Redo 机制**: 增强的连接恢复机制，支持自动数据重新同步

## 使用限制

支持 Go >= v1.24 版本

支持 Nacos >= 3.x 版本

## 安装

使用`go get`安装SDK：

```sh
$ go get -u github.com/nacos-group/nacos-sdk-go/v2
```

## 快速使用

* ClientConfig

```go
constant.ClientConfig{
	TimeoutMs            uint64 // 请求Nacos服务端的超时时间，默认是10000ms
	NamespaceId          string // ACM的命名空间Id
	Endpoint             string // 当使用ACM时，需要该配置. https://help.aliyun.com/document_detail/130146.html
	RegionId             string // ACM&KMS的regionId，用于配置中心的鉴权
	AccessKey            string // ACM&KMS的AccessKey，用于配置中心的鉴权
	SecretKey            string // ACM&KMS的SecretKey，用于配置中心的鉴权
	OpenKMS              bool   // 是否开启kms，默认不开启，kms可以参考文档 https://help.aliyun.com/product/28933.html
	                            // 同时DataId必须以"cipher-"作为前缀才会启动加解密逻辑
	CacheDir             string // 缓存service信息的目录，默认是当前运行目录
	UpdateThreadNum      int    // 监听service变化的并发数，默认20
	NotLoadCacheAtStart  bool   // 在启动的时候不读取缓存在CacheDir的service信息
	UpdateCacheWhenEmpty bool   // 当service返回的实例列表为空时，不更新缓存，用于推空保护
	Username             string // Nacos服务端的API鉴权Username
	Password             string // Nacos服务端的API鉴权Password
	LogDir               string // 日志存储路径
	RotateTime           string // 日志轮转周期，比如：30m, 1h, 24h, 默认是24h
	MaxAge               int64  // 日志最大文件数，默认3
	LogLevel             string // 日志默认级别，值必须是：debug,info,warn,error，默认值是info
}
```

* ServerConfig

```go
constant.ServerConfig{
	ContextPath string // Nacos的ContextPath，默认/nacos，在2.0+中不需要设置
	IpAddr      string // Nacos的服务地址
	Port        uint64 // Nacos的服务端口
	Scheme      string // Nacos的服务地址前缀，默认http，在2.0+中不需要设置
	GrpcPort    uint64 // Nacos的 grpc 服务端口, 默认为 服务端口+1000, 不是必填
}
```

<b>Note：我们可以配置多个ServerConfig，客户端会对这些服务端做轮询请求</b>

### 创建客户端

```go
// 创建clientConfig
clientConfig := constant.ClientConfig{
	NamespaceId:         "e525eafa-f7d7-4029-83d9-008937f9d468", // 如果需要支持多namespace，我们可以创建多个client,它们有不同的NamespaceId。当namespace是public时，此处填空字符串。
	TimeoutMs:           5000,
	NotLoadCacheAtStart: true,
	LogDir:              "/tmp/nacos/log",
	CacheDir:            "/tmp/nacos/cache",
	LogLevel:            "debug",
}

// 创建clientConfig的另一种方式
clientConfig := *constant.NewClientConfig(
    constant.WithNamespaceId("e525eafa-f7d7-4029-83d9-008937f9d468"), // 当namespace是public时，此处填空字符串。
    constant.WithTimeoutMs(5000),
    constant.WithNotLoadCacheAtStart(true),
    constant.WithLogDir("/tmp/nacos/log"),
    constant.WithCacheDir("/tmp/nacos/cache"),
    constant.WithLogLevel("debug"),
)

// 至少一个ServerConfig
serverConfigs := []constant.ServerConfig{
    *constant.NewServerConfig(
        "console1.nacos.io",
        8848,
        constant.WithScheme("http"),
        constant.WithContextPath("/nacos"),
    ),
}

// 创建服务发现客户端
namingClient, err := clients.NewNamingClient(
    vo.NacosClientParam{
        ClientConfig:  &clientConfig,
        ServerConfigs: serverConfigs,
    },
)

// 创建动态配置客户端
configClient, err := clients.NewConfigClient(
    vo.NacosClientParam{
        ClientConfig:  &clientConfig,
        ServerConfigs: serverConfigs,
    },
)

// 创建AI客户端，用于MCP Server和Agent管理（3.x新增）
aiClient, err := clients.NewAIClient(
    vo.NacosClientParam{
        ClientConfig:  &clientConfig,
        ServerConfigs: serverConfigs,
    },
)
defer aiClient.CloseClient()
```

### 服务发现

* 注册实例：RegisterInstance

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
    ClusterName: "cluster-a", // 默认值DEFAULT
    GroupName:   "group-a",   // 默认值DEFAULT_GROUP
})

```

* 注销实例：DeregisterInstance

```go

success, err := namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
    Ip:          "10.0.0.11",
    Port:        8848,
    ServiceName: "demo.go",
    Ephemeral:   true,
    Cluster:     "cluster-a", // 默认值DEFAULT
    GroupName:   "group-a",   // 默认值DEFAULT_GROUP
})

```

* 获取服务信息：GetService

```go

services, err := namingClient.GetService(vo.GetServiceParam{
    ServiceName: "demo.go",
    Clusters:    []string{"cluster-a"}, // 默认值DEFAULT
    GroupName:   "group-a",             // 默认值DEFAULT_GROUP
})

```

* 获取所有的实例列表：SelectAllInstances

```go

// SelectAllInstance可以返回全部实例列表,包括healthy=false,enable=false,weight<=0
instances, err := namingClient.SelectAllInstances(vo.SelectAllInstancesParam{
    ServiceName: "demo.go",
    GroupName:   "group-a",             // 默认值DEFAULT_GROUP
    Clusters:    []string{"cluster-a"}, // 默认值DEFAULT
})

```

* 获取实例列表：SelectInstances

```go

// SelectInstances 只返回满足这些条件的实例列表：healthy=${HealthyOnly},enable=true 和weight>0
instances, err := namingClient.SelectInstances(vo.SelectInstancesParam{
    ServiceName: "demo.go",
    GroupName:   "group-a",             // 默认值DEFAULT_GROUP
    Clusters:    []string{"cluster-a"}, // 默认值DEFAULT
    HealthyOnly: true,
})

```

* 获取一个健康的实例（加权随机轮询）：SelectOneHealthyInstance

```go

// SelectOneHealthyInstance将会按加权随机轮询的负载均衡策略返回一个健康的实例
// 实例必须满足的条件：health=true,enable=true and weight>0
instance, err := namingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
    ServiceName: "demo.go",
    GroupName:   "group-a",             // 默认值DEFAULT_GROUP
    Clusters:    []string{"cluster-a"}, // 默认值DEFAULT
})

```

* 监听服务变化：Subscribe

```go

// Subscribe key=serviceName+groupName+cluster
// 注意:我们可以在相同的key添加多个SubscribeCallback.
err := namingClient.Subscribe(vo.SubscribeParam{
    ServiceName: "demo.go",
    GroupName:   "group-a",             // 默认值DEFAULT_GROUP
    Clusters:    []string{"cluster-a"}, // 默认值DEFAULT
    SubscribeCallback: func(services []model.Instance, err error) {
        log.Printf("\n\n callback return services:%s \n\n", utils.ToJsonString(services))
    },
})

```

* 取消服务监听：Unsubscribe

```go

err := namingClient.Unsubscribe(vo.SubscribeParam{
    ServiceName: "demo.go",
    GroupName:   "group-a",             // 默认值DEFAULT_GROUP
    Clusters:    []string{"cluster-a"}, // 默认值DEFAULT
    SubscribeCallback: func(services []model.Instance, err error) {
        log.Printf("\n\n callback return services:%s \n\n", utils.ToJsonString(services))
    },
})

```

* 获取服务名列表：GetAllServicesInfo

```go

serviceInfos, err := namingClient.GetAllServicesInfo(vo.GetAllServiceInfoParam{
    NameSpace: "0e83cc81-9d8c-4bb8-a28a-ff703187543f",
    PageNo:   1,
    PageSize: 10,
})

```

### 动态配置

* 发布配置：PublishConfig

```go

success, err := configClient.PublishConfig(vo.ConfigParam{
    DataId:  "dataId",
    Group:   "group",
    Content: "hello world!",
})

```

* 删除配置：DeleteConfig

```go

success, err = configClient.DeleteConfig(vo.ConfigParam{
    DataId: "dataId",
    Group:  "group",
})

```

* 获取配置：GetConfig

```go

content, err := configClient.GetConfig(vo.ConfigParam{
    DataId: "dataId",
    Group:  "group",
})

```

* 监听配置变化：ListenConfig

```go

err := configClient.ListenConfig(vo.ConfigParam{
    DataId: "dataId",
    Group:  "group",
    OnChange: func(namespace, group, dataId, data string) {
        fmt.Println("group:" + group + ", dataId:" + dataId + ", data:" + data)
    },
})

```

* 取消配置监听：CancelListenConfig

```go

err := configClient.CancelListenConfig(vo.ConfigParam{
    DataId: "dataId",
    Group:  "group",
})

```

* 搜索配置：SearchConfig

```go

configPage, err := configClient.SearchConfig(vo.SearchConfigParam{
    Search:   "blur",
    DataId:   "",
    Group:    "",
    PageNo:   1,
    PageSize: 10,
})

```

### AI 服务管理（3.x 新增）

AI 模块提供了 MCP (Model Context Protocol) Server 和 A2A (Agent-to-Agent) Agent 的管理能力。

#### MCP Server 操作

* 发布 MCP Server：ReleaseMcpServer

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
            Description: "搜索工具，用于查找信息",
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

* 获取 MCP Server：GetMcpServer

```go

mcpServer, err := aiClient.GetMcpServer(vo.GetMcpServerParam{
    McpName: "demo-mcp-server",
    Version: "v1.0.0",
})

```

* 注册 MCP Server 端点：RegisterMcpServerEndpoint

```go

err := aiClient.RegisterMcpServerEndpoint(vo.RegisterMcpServerEndpointParam{
    McpName: "demo-mcp-server",
    Address: "192.168.1.100",
    Port:    8080,
    Version: "v1.0.0",
})

```

* 注销 MCP Server 端点：DeregisterMcpServerEndpoint

```go

err := aiClient.DeregisterMcpServerEndpoint(vo.DeregisterMcpServerEndpointParam{
    McpName: "demo-mcp-server",
    Address: "192.168.1.100",
    Port:    8080,
})

```

* 订阅 MCP Server 变化：SubscribeMcpServer

```go

mcpServer, err := aiClient.SubscribeMcpServer(vo.SubscribeMcpServerParam{
    McpName: "demo-mcp-server",
    Version: "v1.0.0",
    SubscribeCallback: func(mcpId, namespaceId, mcpName string, mcpServer model.McpServerDetailInfo) {
        fmt.Printf("MCP Server 变化: %s, 端点数量: %d\n", mcpName, len(mcpServer.BackendEndpoints))
    },
})

```

#### A2A Agent 操作

* 发布 Agent Card：ReleaseAgentCard

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
            Name:        "搜索技能",
            Description: "可以搜索信息",
        },
    },
}

err := aiClient.ReleaseAgentCard(vo.ReleaseAgentCardParam{
    AgentCard:        agentCard,
    RegistrationType: constant.A2A_ENDPOINT_TYPE_SERVICE,
    SetAsLatest:      true,
})

```

* 获取 Agent Card：GetAgentCard

```go

agentCardInfo, err := aiClient.GetAgentCard(vo.GetAgentCardParam{
    AgentName:        "demo-agent",
    Version:          "v1.0.0",
    RegistrationType: constant.A2A_ENDPOINT_TYPE_URL,
})

```

* 注册 Agent 端点：RegisterAgentEndpoint

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

* 注销 Agent 端点：DeregisterAgentEndpoint

```go

err := aiClient.DeregisterAgentEndpoint(vo.DeregisterAgentEndpointParam{
    AgentName: "demo-agent",
    Address:   "192.168.1.100",
    Port:      9090,
    Version:   "v1.0.0",
})

```

* 订阅 Agent Card 变化：SubscribeAgentCard

```go

agentCardInfo, err := aiClient.SubscribeAgentCard(vo.SubscribeAgentCardParam{
    AgentName: "demo-agent",
    Version:   "v1.0.0",
    SubscribeCallback: func(agentName string, agentCard model.AgentCardDetailInfo) {
        fmt.Printf("Agent Card 变化: %s, 最新版本: %v\n", agentName, agentCard.LatestVersion)
    },
})

```

## 示例

我们能从示例中学习如何使用Nacos go客户端：

* [动态配置示例](./example/config)
* [服务发现示例](./example/service)
* [AI 示例](./example/ai)

## 文档

Nacos open-api相关信息可以查看文档 [Nacos open-api website](https://nacos.io/en-us/docs/open-api.html).

Nacos产品了解可以查看 [Nacos website](https://nacos.io/en-us/docs/what-is-nacos.html).

## 贡献代码

我们非常欢迎大家为Nacos-sdk-go贡献代码. 贡献前请查看[CONTRIBUTING.md](./CONTRIBUTING.md)

## 联系我们

* 加入Nacos-sdk-go钉钉群(23191211).
* [Gitter](https://gitter.im/alibaba/nacos): Nacos即时聊天工具.
* [Twitter](https://twitter.com/nacos2): 在Twitter上关注Nacos的最新动态.
* [Weibo](https://weibo.com/u/6574374908): 在微博上关注Nacos的最新动态.
* [Nacos SegmentFault](https://segmentfault.com/t/nacos): SegmentFault可以获得最新的推送和帮助.
* Email Group:
     * users-nacos@googlegroups.com: Nacos用户讨论组.
     * dev-nacos@googlegroups.com: Nacos开发者讨论组 (APIs, feature design, etc).
     * commits-nacos@googlegroups.com: Nacos commit提醒.
