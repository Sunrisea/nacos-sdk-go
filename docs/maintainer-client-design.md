# Nacos Go SDK — Maintainer Client 设计方案

## 一、概述

### 1.1 背景

Maintainer Client 是一个管控面客户端，通过调用 Nacos Server 的 `/v3/admin/` 系列 HTTP 接口，
提供运维管理能力（Namespace、集群、服务、实例、配置、AI 等的管理操作）。
Go SDK 需要对标 Java `nacos/maintainer-client` 模块的功能进行实现。

### 1.2 与数据面客户端的区别

| 维度 | 数据面客户端 (Config/Naming/AI) | 管控面客户端 (Maintainer) |
|------|---------------------------------|--------------------------|
| 传输协议 | gRPC（主）+ HTTP（辅） | 纯 HTTP |
| API 路径 | `/v2/...` 或 gRPC | `/v3/admin/...` |
| 使用场景 | 应用运行时读写配置、注册发现 | 运维管理、控制台操作 |
| 连接特征 | 长连接、推送、重连、redo | 短连接、请求-响应 |
| Body 格式 | form-urlencoded / Protobuf | form-urlencoded + JSON body |

### 1.3 对标 Java 接口

| Java 接口 | Go 接口 |
|-----------|---------|
| `CoreMaintainerService` | `ICoreMaintainerClient` |
| `NamingMaintainerService` (含 Service/Instance/NamingClient 子接口) | `INamingMaintainerClient` |
| `ConfigMaintainerService` (含 Beta/History/Ops 子接口) | `IConfigMaintainerClient` |
| `AiMaintainerService` (含 MCP/A2A 子接口) | `IAiMaintainerClient` |
| `AbstractCoreMaintainerService` | `CoreMaintainerClient` (struct，被组合嵌入) |
| `ClientHttpProxy` | `MaintainerHttpProxy` |
| `NacosMaintainerFactory` 等 | `client_factory.go` 工厂函数 |

---

## 二、组件复用分析

### 2.1 完全复用（无需修改）

| 组件 | 路径 | 复用方式 |
|------|------|----------|
| **NacosClient** | `clients/nacos_client/` | 作为 base client 持有 ClientConfig、ServerConfig、HttpAgent |
| **client_factory.setConfig()** | `clients/client_factory.go` | 复用创建 NacosClient 的逻辑 |
| **ClientConfig / ServerConfig** | `common/constant/config.go` | 客户端配置和服务端配置 struct，直接使用 |
| **SecurityProxy** | `common/security/security_proxy.go` | 鉴权代理，支持 Nacos Auth + RAM Auth |
| **NacosAuthClient** | `common/security/nacos_auth_client.go` | 用户名密码登录，获取 accessToken |
| **RamAuthClient** | `common/security/ram_auth_client.go` | RAM 鉴权（AK/SK、ECS Role 等） |
| **ResourceInjector** | `common/security/resource_injector.go` | 签名注入（Config/Naming/AI 三种），复用已有类型 |
| **NacosError** | `common/nacos_error/nacos_error.go` | 统一错误类型 |
| **Logger** | `common/logger/` | 日志 |
| **vo.NacosClientParam** | `vo/client_param.go` | 客户端创建参数，直接复用 |

### 2.2 部分复用（需扩展）

| 组件 | 路径 | 复用方式 | 需要扩展 |
|------|------|----------|----------|
| **NacosServer** | `common/nacos_server/nacos_server.go` | 复用服务列表管理、鉴权注入、重试/failover | 新增 `ReqAdminApi` 方法，支持 JSON body + 灵活的鉴权 resource 构建 |
| **IHttpAgent / HttpAgent** | `common/http_agent/` | 复用 GET、DELETE（query params） | POST/PUT 当前只支持 form-urlencoded，需新增 JSON body 支持 |
| **RequestResource** | `common/security/security_proxy.go` | 复用现有结构和 Build 函数 | 新增 `BuildAdminResource()` 便捷构造方法 |

### 2.3 新增开发

| 组件 | 说明 |
|------|------|
| `MaintainerHttpProxy` | Admin API 专用 HTTP 代理，封装请求构建、JSON body、鉴权、重试 |
| `CoreMaintainerClient` | Core 管理实现（namespace、cluster、loader、health、plugin） |
| `NamingMaintainerClient` | Naming 管理实现（service、instance、client、health） |
| `ConfigMaintainerClient` | Config 管理实现（CRUD、history、beta、listener、ops） |
| `AiMaintainerClient` | AI 管理实现（MCP server、A2A agent） |
| Maintainer VO / Model | 管控面入参和返回值类型定义 |
| Admin API 路径常量 | `/v3/admin/...` 路径常量 |

---

## 三、整体架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                               用户应用                                      │
└─────────────────────────────────────────────────────────────────────────────┘
          │                    │                    │
          ▼                    ▼                    ▼
  NewNamingMaintainer   NewConfigMaintainer   NewAiMaintainer
  Client(param)         Client(param)         Client(param)
          │                    │                    │
          ▼                    ▼                    ▼
┌────────────────┐  ┌────────────────┐  ┌────────────────┐
│ NamingMaintainer│  │ConfigMaintainer│  │  AiMaintainer  │
│    Client      │  │    Client      │  │    Client      │
└───────┬────────┘  └───────┬────────┘  └───────┬────────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │ 嵌入组合
                            ▼
               ┌─────────────────────────┐
               │  CoreMaintainerClient   │  ← 公共 Core 能力实现
               └────────────┬────────────┘
                            │ 持有
                            ▼
               ┌─────────────────────────┐
               │  MaintainerHttpProxy    │  ← Admin HTTP 请求代理
               └────────────┬────────────┘
                            │ 使用
             ┌──────────────┼──────────────────┐
             ▼              ▼                  ▼
     ┌──────────────┐ ┌──────────────┐  ┌───────────────┐
     │ NacosServer  │ │ SecurityProxy│  │   HttpAgent   │  ← 全部复用
     │ (服务列表,    │ │ (鉴权)       │  │   (HTTP)      │
     │  重试,failover)│ │              │  │               │
     └──────┬───────┘ └──────────────┘  └───────────────┘
            │
            ▼
     ┌──────────────────┐
     │  Nacos Server     │
     │  /v3/admin/ APIs  │
     └──────────────────┘
```

---

## 四、需要扩展的现有组件

### 4.1 NacosServer — 新增 `ReqAdminApi` 方法

现有的 `ReqConfigApi` 和 `ReqApi` 分别绑定了 Config 和 Naming 的鉴权 resource 构建和 header 设置。
需要新增一个更通用的 `ReqAdminApi` 方法：

```go
// ReqAdminApi 支持 admin 接口调用，兼容 form-urlencoded 和 JSON body
func (server *NacosServer) ReqAdminApi(
    api string,
    params map[string]string,       // query params 或 form params
    headers map[string]string,      // 自定义 headers
    body []byte,                    // JSON body（可选，nil 表示使用 form-urlencoded）
    method string,
    resource security.RequestResource,
    timeoutMS uint64,
) (string, error)
```

核心逻辑（复用现有模式）：
- 调用 `InjectSecurityInfo(params, resource)` 注入鉴权信息
- 服务列表 + 重试/failover 逻辑与 `ReqConfigApi` 完全相同
- 新增 `callAdminServer` 内部方法，当 `body != nil` 时设置 `Content-Type: application/json`

### 4.2 IHttpAgent — 新增 JSON Body 支持

现有 `Post`/`Put` 只支持 form-urlencoded。需扩展：

```go
// 在 IHttpAgent 接口新增方法
RequestWithBody(method string, path string, header http.Header,
    timeoutMs uint64, params map[string]string, body []byte) (*http.Response, error)
```

或在 `MaintainerHttpProxy` 内部直接使用 `net/http` 构造 JSON 请求，绕过 `IHttpAgent`
（避免修改公共接口影响面过大）。

**推荐方案**：在 `MaintainerHttpProxy` 层处理 JSON body，不修改 `IHttpAgent` 接口，
降低对现有数据面客户端的影响。

### 4.3 RequestResource — 新增 Admin 构建方法

```go
// common/security/security_proxy.go 中新增
func BuildAdminResource(namespace, group, resource string) RequestResource {
    return RequestResource{
        requestType: REQUEST_TYPE_CONFIG,  // admin 接口复用 config 类型的签名逻辑
        namespace:   namespace,
        group:       group,
        resource:    resource,
    }
}
```

> 说明：Admin 接口的鉴权签名逻辑与 Config 类型一致，无需新增 resource type。

---

## 五、新增组件详细设计

### 5.1 Admin API 路径常量

```go
// common/constant/admin_api_path.go

package constant

const (
    // Config Admin
    AdminConfigPath         = "/v3/admin/cs/config"
    AdminConfigHistoryPath  = "/v3/admin/cs/history"
    AdminConfigOpsPath      = "/v3/admin/cs/ops"
    AdminConfigListenerPath = "/v3/admin/cs/listener"

    // Naming Admin
    AdminNamingServicePath  = "/v3/admin/ns/service"
    AdminNamingInstancePath = "/v3/admin/ns/instance"
    AdminNamingClusterPath  = "/v3/admin/ns/cluster"
    AdminNamingHealthPath   = "/v3/admin/ns/health"
    AdminNamingClientPath   = "/v3/admin/ns/client"
    AdminNamingOpsPath      = "/v3/admin/ns/ops"

    // Core Admin
    AdminCoreLoaderPath     = "/v3/admin/core/loader"
    AdminCoreClusterPath    = "/v3/admin/core/cluster"
    AdminCoreOpsPath        = "/v3/admin/core/ops"
    AdminCoreNamespacePath  = "/v3/admin/core/namespace"
    AdminCoreStatePath      = "/v3/admin/core/state"
    AdminCorePluginPath     = "/v3/admin/core/plugin"

    // AI Admin
    AdminAiMcpPath  = "/v3/admin/ai/mcp"
    AdminAiA2aPath  = "/v3/admin/ai/a2a"
)
```

### 5.2 MaintainerHttpProxy — Admin HTTP 代理

```go
// clients/maintainer_client/remote/maintainer_http_proxy.go

package remote

type MaintainerHttpProxy struct {
    nacosServer *nacos_server.NacosServer
    clientCfg   constant.ClientConfig
}

// NewMaintainerHttpProxy 创建 HTTP 代理，复用 NacosServer（含服务列表、鉴权、重试）
func NewMaintainerHttpProxy(ctx context.Context, serverCfgs []constant.ServerConfig,
    clientCfg constant.ClientConfig, httpAgent http_agent.IHttpAgent,
    provider security.RamCredentialProvider) (*MaintainerHttpProxy, error)

// --- 核心请求方法 ---

// ReqApi 发送 form-urlencoded 请求（GET/DELETE/POST/PUT）
func (p *MaintainerHttpProxy) ReqApi(api string, params map[string]string,
    method string, resource security.RequestResource) (string, error)

// ReqApiWithJsonBody 发送 JSON body 请求（POST/PUT）
func (p *MaintainerHttpProxy) ReqApiWithJsonBody(api string, params map[string]string,
    body interface{}, method string, resource security.RequestResource) (string, error)

// --- 响应解析 ---

// ParseResult 解析 Nacos v3 统一返回格式 {"code":0, "message":"...", "data":...}
func ParseResult[T any](response string) (T, error)

// ParseBoolResult 解析返回 bool 的响应
func ParseBoolResult(response string) (bool, error)
```

**复用关系**：内部使用 `NacosServer` 的服务列表、鉴权注入、重试机制。
JSON body 请求在此层构造，使用 `net/http` 直接发送，不改动 `IHttpAgent` 接口。

### 5.3 接口定义

#### 5.3.1 ICoreMaintainerClient

```go
// clients/maintainer_client/core/core_maintainer_interface.go

type ICoreMaintainerClient interface {
    // --- 服务器状态 ---
    GetServerState() (map[string]string, error)
    Liveness() (bool, error)
    Readiness() (bool, error)

    // --- Raft ---
    RaftOps(command, value, groupId string) (string, error)

    // --- ID 生成器 ---
    GetIdGenerators() ([]model.IdGeneratorInfo, error)

    // --- 日志 ---
    UpdateLogLevel(logName, logLevel string) error

    // --- 集群管理 ---
    ListClusterNodes(address, state string) ([]model.NacosMember, error)
    UpdateLookupMode(lookupType string) (bool, error)

    // --- 连接/负载管理 ---
    GetCurrentClients() (map[string]model.ConnectionInfo, error)
    ReloadConnectionCount(count int, redirectAddress string) (string, error)
    SmartReloadCluster(loaderFactor string) (string, error)
    ReloadSingleClient(connectionId, redirectAddress string) (string, error)
    GetClusterLoaderMetrics() (model.ServerLoaderMetrics, error)

    // --- Namespace ---
    GetNamespaceList() ([]model.Namespace, error)
    GetNamespace(namespaceId string) (model.Namespace, error)
    CreateNamespace(param vo.CreateNamespaceParam) (bool, error)
    UpdateNamespace(param vo.UpdateNamespaceParam) (bool, error)
    DeleteNamespace(namespaceId string) (bool, error)
    CheckNamespaceIdExist(namespaceId string) (bool, error)

    // --- 插件管理 ---
    ListPlugins(pluginType string) ([]map[string]interface{}, error)
    GetPluginDetail(pluginType, pluginName string) (map[string]interface{}, error)
    UpdatePluginStatus(pluginType, pluginName string, enabled bool) error
    UpdatePluginConfig(pluginType, pluginName string, config map[string]string) error
    GetPluginAvailability(pluginType, pluginName string) (map[string]bool, error)

    // --- 生命周期 ---
    CloseClient()
}
```

#### 5.3.2 INamingMaintainerClient

```go
// clients/maintainer_client/naming/naming_maintainer_interface.go

type INamingMaintainerClient interface {
    ICoreMaintainerClient

    // --- Service 管理 ---
    CreateService(param vo.MaintainerServiceParam) (string, error)
    UpdateService(param vo.MaintainerServiceParam) (string, error)
    RemoveService(param vo.MaintainerServiceParam) (string, error)
    GetServiceDetail(param vo.MaintainerServiceParam) (model.ServiceDetailInfo, error)
    ListServices(param vo.ListServicesParam) (model.Page[model.ServiceView], error)
    ListServicesWithDetail(param vo.ListServicesDetailParam) (model.Page[model.ServiceDetailInfo], error)
    GetSubscribers(param vo.GetSubscribersParam) (model.Page[model.SubscriberInfo], error)
    ListSelectorTypes() ([]string, error)

    // --- Instance 管理 ---
    RegisterInstance(param vo.MaintainerInstanceParam) (string, error)
    DeregisterInstance(param vo.MaintainerInstanceParam) (string, error)
    UpdateInstance(param vo.MaintainerInstanceParam) (string, error)
    BatchUpdateInstanceMetadata(param vo.BatchInstanceMetadataParam) (model.InstanceMetadataBatchResult, error)
    BatchDeleteInstanceMetadata(param vo.BatchInstanceMetadataParam) (model.InstanceMetadataBatchResult, error)
    PartialUpdateInstance(param vo.MaintainerInstanceParam) (string, error)
    ListInstances(param vo.ListInstancesParam) ([]model.Instance, error)
    GetInstanceDetail(param vo.GetInstanceDetailParam) (model.Instance, error)

    // --- Naming Client 信息查询 ---
    GetClientList() ([]string, error)
    GetClientDetail(clientId string) (model.ClientSummaryInfo, error)
    GetPublishedServiceList(clientId string) ([]model.ClientServiceInfo, error)
    GetSubscribeServiceList(clientId string) ([]model.ClientServiceInfo, error)
    GetPublishedClientList(param vo.GetServiceClientParam) ([]model.ClientPublisherInfo, error)
    GetSubscribeClientList(param vo.GetServiceClientParam) ([]model.ClientSubscriberInfo, error)

    // --- 健康管理 ---
    GetMetrics(onlyStatus bool) (model.MetricsInfo, error)
    UpdateInstanceHealthStatus(param vo.UpdateInstanceHealthParam) (string, error)
    GetHealthCheckers() (map[string]interface{}, error)
    UpdateCluster(param vo.UpdateClusterParam) (string, error)
}
```

#### 5.3.3 IConfigMaintainerClient

```go
// clients/maintainer_client/config/config_maintainer_interface.go

type IConfigMaintainerClient interface {
    ICoreMaintainerClient

    // --- Config CRUD ---
    GetConfig(param vo.MaintainerConfigParam) (model.ConfigDetailInfo, error)
    PublishConfig(param vo.MaintainerPublishConfigParam) (bool, error)
    UpdateConfigMetadata(param vo.UpdateConfigMetadataParam) (bool, error)
    DeleteConfig(param vo.MaintainerConfigParam) (bool, error)
    DeleteConfigs(ids []int64) (bool, error)

    // --- 列表/搜索 ---
    ListConfigs(param vo.SearchConfigParam) (model.Page[model.ConfigBasicInfo], error)
    SearchConfigs(param vo.SearchConfigParam) (model.Page[model.ConfigBasicInfo], error)
    GetConfigListByNamespace(namespaceId string) ([]model.ConfigBasicInfo, error)

    // --- Listener ---
    GetListeners(param vo.GetListenersParam) (model.ConfigListenerInfo, error)
    GetAllSubClientConfigByIp(param vo.GetSubClientConfigParam) (model.ConfigListenerInfo, error)

    // --- Clone ---
    CloneConfig(param vo.CloneConfigParam) (map[string]interface{}, error)

    // --- Beta ---
    PublishBetaConfig(param vo.PublishBetaConfigParam) (bool, error)
    StopBeta(param vo.StopBetaParam) (bool, error)
    QueryBeta(param vo.QueryBetaParam) (model.ConfigDetailInfo, error)

    // --- History ---
    ListConfigHistory(param vo.ListConfigHistoryParam) (model.Page[model.ConfigHistoryInfo], error)
    GetConfigHistoryInfo(param vo.GetConfigHistoryParam) (model.ConfigHistoryInfo, error)
    GetPreviousConfigHistoryInfo(param vo.GetConfigHistoryParam) (model.ConfigHistoryInfo, error)

    // --- Ops ---
    UpdateLocalCacheFromStore() error
}
```

#### 5.3.4 IAiMaintainerClient

```go
// clients/maintainer_client/ai/ai_maintainer_interface.go

type IAiMaintainerClient interface {
    ICoreMaintainerClient

    // --- MCP Server ---
    ListMcpServer(param vo.ListMcpServerParam) (model.Page[model.McpServerBasicInfo], error)
    SearchMcpServer(param vo.SearchMcpServerParam) (model.Page[model.McpServerBasicInfo], error)
    GetMcpServerDetail(param vo.GetMcpServerDetailParam) (model.McpServerDetailInfo, error)
    CreateMcpServer(param vo.CreateMcpServerParam) (string, error)
    UpdateMcpServer(param vo.UpdateMcpServerParam) (bool, error)
    DeleteMcpServer(param vo.DeleteMcpServerParam) (bool, error)

    // --- A2A Agent ---
    RegisterAgent(param vo.RegisterAgentParam) error
    GetAgentCard(param vo.GetMaintainerAgentCardParam) (model.AgentCardInfo, error)
    UpdateAgentCard(param vo.UpdateAgentCardParam) error
    DeleteAgent(param vo.DeleteAgentParam) error
    ListAllVersionOfAgent(param vo.ListAgentVersionParam) ([]model.AgentCardInfo, error)
    SearchAgentCardsByName(param vo.SearchAgentParam) (model.Page[model.AgentCardInfo], error)
    ListAgentCards(param vo.ListAgentCardsParam) (model.Page[model.AgentCardInfo], error)
}
```

### 5.4 实现层 — struct 组合关系

```go
// CoreMaintainerClient 提供 Core 管理能力的共享实现
type CoreMaintainerClient struct {
    ctx    context.Context
    cancel context.CancelFunc
    nacos_client.INacosClient          // 嵌入 base client
    proxy  *remote.MaintainerHttpProxy // HTTP 代理
}

// NamingMaintainerClient 嵌入 CoreMaintainerClient
type NamingMaintainerClient struct {
    *CoreMaintainerClient              // 继承 Core 能力
}

// ConfigMaintainerClient 嵌入 CoreMaintainerClient
type ConfigMaintainerClient struct {
    *CoreMaintainerClient              // 继承 Core 能力
}

// AiMaintainerClient 嵌入 CoreMaintainerClient
type AiMaintainerClient struct {
    *CoreMaintainerClient              // 继承 Core 能力
}
```

### 5.5 工厂方法

在 `clients/client_factory.go` 中新增：

```go
func NewNamingMaintainerClient(param vo.NacosClientParam) (
    naming.INamingMaintainerClient, error) {
    nacosClient, err := setConfig(param)   // 复用现有 setConfig
    if err != nil {
        return nil, err
    }
    return naming.NewNamingMaintainerClient(nacosClient, param.RamCredentialProvider)
}

func NewConfigMaintainerClient(param vo.NacosClientParam) (
    config.IConfigMaintainerClient, error) { ... }

func NewAiMaintainerClient(param vo.NacosClientParam) (
    ai.IAiMaintainerClient, error) { ... }
```

### 5.6 Model 新增类型

```go
// model/maintainer.go

// 通用分页（Go 1.18+ 泛型）
type Page[T any] struct {
    TotalCount     int `json:"totalCount"`
    PageNumber     int `json:"pageNumber"`
    PagesAvailable int `json:"pagesAvailable"`
    PageItems      []T `json:"pageItems"`
}

// Nacos v3 统一响应
type RestResult[T any] struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    T      `json:"data"`
}

// Namespace
type Namespace struct {
    Namespace         string `json:"namespace"`
    NamespaceShowName string `json:"namespaceShowName"`
    NamespaceDesc     string `json:"namespaceDesc"`
    Quota             int    `json:"quota"`
    ConfigCount       int    `json:"configCount"`
    Type              int    `json:"type"`
}

// 集群节点
type NacosMember struct {
    Ip         string            `json:"ip"`
    Port       int               `json:"port"`
    State      string            `json:"state"`
    Address    string            `json:"address"`
    ExtendInfo map[string]string `json:"extendInfo"`
}

// 连接信息
type ConnectionInfo struct {
    ConnectionId string            `json:"connectionId"`
    ClientIp     string            `json:"clientIp"`
    ClientPort   int               `json:"clientPort"`
    ConnectType  string            `json:"connectType"`
    Version      string            `json:"version"`
    Labels       map[string]string `json:"labels"`
    CreateTime   string            `json:"createTime"`
}

// 服务器负载指标
type ServerLoaderMetrics struct {
    Address    string  `json:"address"`
    Metric     string  `json:"metric"`
    Load       float64 `json:"load"`
    SdkCount   int     `json:"sdkCount"`
    Connection int     `json:"connection"`
    Cpu        float64 `json:"cpu"`
}

// ID 生成器
type IdGeneratorInfo struct {
    Resource string `json:"resource"`
    Current  int64  `json:"current"`
}

// 服务详情（管控面）
type ServiceDetailInfo struct {
    Namespace        string                    `json:"namespace"`
    GroupName        string                    `json:"groupName"`
    ServiceName      string                    `json:"serviceName"`
    ProtectThreshold float64                   `json:"protectThreshold"`
    Metadata         map[string]string         `json:"metadata"`
    Ephemeral        bool                      `json:"ephemeral"`
    Selector         map[string]interface{}    `json:"selector"`
    Clusters         map[string]ClusterInfo    `json:"clusters"`
}

// 集群信息
type ClusterInfo struct {
    ClusterName           string                 `json:"clusterName"`
    HealthChecker         map[string]interface{} `json:"healthChecker"`
    HealthyCheckPort      int                    `json:"healthyCheckPort"`
    UseInstancePortForCheck bool                 `json:"useInstancePortForCheck"`
    Metadata              map[string]string      `json:"metadata"`
    Hosts                 []Instance             `json:"hosts"`
}

// 服务摘要
type ServiceView struct {
    Namespace        string  `json:"namespace"`
    GroupName        string  `json:"groupName"`
    ServiceName      string  `json:"serviceName"`
    ProtectThreshold float64 `json:"protectThreshold"`
    ClusterCount     int     `json:"clusterCount"`
    IpCount          int     `json:"ipCount"`
    HealthyCount     int     `json:"healthyCount"`
    Ephemeral        bool    `json:"ephemeral"`
}

// 订阅者
type SubscriberInfo struct {
    Address     string `json:"address"`
    Agent       string `json:"agent"`
    App         string `json:"app"`
    Ip          string `json:"ip"`
    Port        int    `json:"port"`
    NamespaceId string `json:"namespaceId"`
    ServiceName string `json:"serviceName"`
    AddrStr     string `json:"addrStr"`
}

// 批量元数据结果
type InstanceMetadataBatchResult struct {
    Updated []string `json:"updated"`
}

// Naming 客户端信息
type ClientSummaryInfo struct {
    ClientId    string            `json:"clientId"`
    ClientType  string            `json:"clientType"`
    ConnectType string            `json:"connectType"`
    LastUpdated int64             `json:"lastUpdatedTime"`
    Labels      map[string]string `json:"labels"`
}

type ClientServiceInfo struct {
    Namespace   string `json:"namespace"`
    GroupName   string `json:"groupName"`
    ServiceName string `json:"serviceName"`
}

type ClientPublisherInfo struct {
    Ip          string            `json:"ip"`
    Port        int               `json:"port"`
    ClusterName string            `json:"clusterName"`
    Metadata    map[string]string `json:"metadata"`
    ClientId    string            `json:"clientId"`
}

type ClientSubscriberInfo struct {
    Address     string `json:"address"`
    App         string `json:"app"`
    Agent       string `json:"agent"`
    ClientId    string `json:"clientId"`
}

// Config 相关
type ConfigDetailInfo struct {
    DataId      string `json:"dataId"`
    Group       string `json:"group"`
    Tenant      string `json:"tenant"`
    Content     string `json:"content"`
    Md5         string `json:"md5"`
    Type        string `json:"type"`
    AppName     string `json:"appName"`
    Desc        string `json:"desc"`
    CreateTime  int64  `json:"createTime"`
    ModifyTime  int64  `json:"modifyTime"`
    CreateUser  string `json:"createUser"`
    CreateIp    string `json:"createIp"`
}

type ConfigBasicInfo struct {
    DataId  string `json:"dataId"`
    Group   string `json:"group"`
    Tenant  string `json:"tenant"`
    Type    string `json:"type"`
    AppName string `json:"appName"`
}

type ConfigHistoryInfo struct {
    Id               int64  `json:"id"`
    DataId           string `json:"dataId"`
    Group            string `json:"group"`
    Tenant           string `json:"tenant"`
    Content          string `json:"content"`
    Md5              string `json:"md5"`
    Type             string `json:"type"`
    CreatedTime      string `json:"createdTime"`
    LastModifiedTime string `json:"lastModifiedTime"`
    SrcUser          string `json:"srcUser"`
    SrcIp            string `json:"srcIp"`
    OpType           string `json:"opType"`
    PublishType      string `json:"publishType"`
}

type ConfigListenerInfo struct {
    ListenersGroupkeyStatus map[string]string `json:"listenersGroupkeyStatus"`
}

// 指标
type MetricsInfo struct {
    Status          string `json:"status"`
    ServiceCount    int    `json:"serviceCount"`
    InstanceCount   int    `json:"instanceCount"`
    SubscribeCount  int    `json:"subscriberCount"`
    ResponsibleServiceCount int `json:"responsibleServiceCount"`
}
```

### 5.7 VO 新增类型

```go
// vo/maintainer_param.go

// --- Namespace ---
type CreateNamespaceParam struct {
    NamespaceId   string `param:"namespaceId"`
    NamespaceName string `param:"namespaceName"`
    NamespaceDesc string `param:"namespaceDesc"`
}

type UpdateNamespaceParam struct {
    NamespaceId   string `param:"namespaceId"`
    NamespaceName string `param:"namespaceName"`
    NamespaceDesc string `param:"namespaceDesc"`
}

// --- Service ---
type MaintainerServiceParam struct {
    NamespaceId      string            `param:"namespaceId"`
    GroupName        string            `param:"groupName"`
    ServiceName      string            `param:"serviceName"`
    Ephemeral        bool              `param:"ephemeral"`
    ProtectThreshold float64           `param:"protectThreshold"`
    Metadata         map[string]string `param:"metadata"`
    Selector         string            `param:"selector"`
}

type ListServicesParam struct {
    NamespaceId      string `param:"namespaceId"`
    GroupNameParam   string `param:"groupNameParam"`
    ServiceNameParam string `param:"serviceNameParam"`
    IgnoreEmptyService bool `param:"hasIpCount"`
    PageNo           int    `param:"pageNo"`
    PageSize         int    `param:"pageSize"`
}

type ListServicesDetailParam struct {
    NamespaceId      string `param:"namespaceId"`
    GroupNameParam   string `param:"groupNameParam"`
    ServiceNameParam string `param:"serviceNameParam"`
    PageNo           int    `param:"pageNo"`
    PageSize         int    `param:"pageSize"`
}

type GetSubscribersParam struct {
    NamespaceId string `param:"namespaceId"`
    GroupName   string `param:"groupName"`
    ServiceName string `param:"serviceName"`
    PageNo      int    `param:"pageNo"`
    PageSize    int    `param:"pageSize"`
    Aggregation bool   `param:"aggregation"`
}

// --- Instance ---
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

type BatchInstanceMetadataParam struct {
    NamespaceId string            `param:"namespaceId"`
    GroupName   string            `param:"groupName"`
    ServiceName string            `param:"serviceName"`
    Instances   string            `param:"instances"`   // JSON array string
    Metadata    map[string]string `param:"metadata"`
    Ephemeral   bool              `param:"ephemeral"`
}

type ListInstancesParam struct {
    NamespaceId string `param:"namespaceId"`
    GroupName   string `param:"groupName"`
    ServiceName string `param:"serviceName"`
    ClusterName string `param:"clusterName"`
    HealthyOnly bool   `param:"healthyOnly"`
}

type GetInstanceDetailParam struct {
    NamespaceId string `param:"namespaceId"`
    GroupName   string `param:"groupName"`
    ServiceName string `param:"serviceName"`
    Ip          string `param:"ip"`
    Port        int    `param:"port"`
    ClusterName string `param:"clusterName"`
}

// --- Naming Client ---
type GetServiceClientParam struct {
    NamespaceId string `param:"namespaceId"`
    GroupName   string `param:"groupName"`
    ServiceName string `param:"serviceName"`
    Ip          string `param:"ip"`
    Port        int    `param:"port"`
}

// --- Health ---
type UpdateInstanceHealthParam struct {
    NamespaceId string `param:"namespaceId"`
    GroupName   string `param:"groupName"`
    ServiceName string `param:"serviceName"`
    Ip          string `param:"ip"`
    Port        int    `param:"port"`
    ClusterName string `param:"clusterName"`
    Healthy     bool   `param:"healthy"`
}

type UpdateClusterParam struct {
    NamespaceId       string            `param:"namespaceId"`
    GroupName         string            `param:"groupName"`
    ServiceName       string            `param:"serviceName"`
    ClusterName       string            `param:"clusterName"`
    HealthChecker     string            `param:"checkType"`
    HealthCheckPort   int               `param:"healthyCheckPort"`
    UseInstancePort   bool              `param:"useInstancePortForCheck"`
    Metadata          map[string]string `param:"metadata"`
}

// --- Config ---
type MaintainerConfigParam struct {
    DataId      string `param:"dataId"`
    Group       string `param:"group"`
    NamespaceId string `param:"namespaceId"`
}

type MaintainerPublishConfigParam struct {
    DataId      string `param:"dataId"`
    Group       string `param:"group"`
    NamespaceId string `param:"namespaceId"`
    Content     string `param:"content"`
    AppName     string `param:"appName"`
    SrcUser     string `param:"srcUser"`
    ConfigTags  string `param:"configTags"`
    Desc        string `param:"desc"`
    Type        string `param:"type"`
}

type UpdateConfigMetadataParam struct {
    DataId      string `param:"dataId"`
    Group       string `param:"group"`
    NamespaceId string `param:"namespaceId"`
    Desc        string `param:"desc"`
    ConfigTags  string `param:"configTags"`
}

type SearchConfigParam struct {
    DataId       string `param:"dataId"`
    Group        string `param:"group"`
    NamespaceId  string `param:"namespaceId"`
    Search       string `param:"search"`      // "blur" 或 "accurate"
    ConfigDetail string `param:"configDetail"` // 内容模糊搜索
    Type         string `param:"type"`
    ConfigTags   string `param:"configTags"`
    AppName      string `param:"appName"`
    PageNo       int    `param:"pageNo"`
    PageSize     int    `param:"pageSize"`
}

type GetListenersParam struct {
    DataId      string `param:"dataId"`
    Group       string `param:"group"`
    NamespaceId string `param:"namespaceId"`
    Aggregation bool   `param:"aggregation"`
}

type GetSubClientConfigParam struct {
    Ip          string `param:"ip"`
    All         bool   `param:"all"`
    NamespaceId string `param:"namespaceId"`
    Aggregation bool   `param:"aggregation"`
}

type CloneConfigParam struct {
    NamespaceId string                 `json:"namespaceId"`
    CloneInfos  []ConfigCloneInfo      `json:"cloneInfos"`
    SrcUser     string                 `json:"srcUser"`
    Policy      string                 `json:"policy"`
}

type ConfigCloneInfo struct {
    CfgId      int64  `json:"cfgId"`
    DataId     string `json:"dataId"`
    Group      string `json:"group"`
}

// --- Beta ---
type PublishBetaConfigParam struct {
    DataId      string `param:"dataId"`
    Group       string `param:"group"`
    NamespaceId string `param:"namespaceId"`
    Content     string `param:"content"`
    BetaIps     string `param:"betaIps"`
}

type StopBetaParam struct {
    DataId      string `param:"dataId"`
    Group       string `param:"group"`
    NamespaceId string `param:"namespaceId"`
}

type QueryBetaParam = StopBetaParam

// --- History ---
type ListConfigHistoryParam struct {
    DataId      string `param:"dataId"`
    Group       string `param:"group"`
    NamespaceId string `param:"namespaceId"`
    PageNo      int    `param:"pageNo"`
    PageSize    int    `param:"pageSize"`
}

type GetConfigHistoryParam struct {
    DataId      string `param:"dataId"`
    Group       string `param:"group"`
    NamespaceId string `param:"namespaceId"`
    Nid         int64  `param:"nid"`
}

// --- AI MCP ---
type ListMcpServerParam struct {
    NamespaceId string `param:"namespaceId"`
    McpName     string `param:"mcpName"`
    PageNo      int    `param:"pageNo"`
    PageSize    int    `param:"pageSize"`
}

type SearchMcpServerParam = ListMcpServerParam

type GetMcpServerDetailParam struct {
    NamespaceId string `param:"namespaceId"`
    McpName     string `param:"mcpName"`
    McpId       string `param:"mcpId"`
    Version     string `param:"version"`
}

type CreateMcpServerParam struct {
    NamespaceId  string                 `json:"namespaceId"`
    McpName      string                 `json:"mcpName"`
    ServerSpec   interface{}            `json:"serverSpec"`
    ToolSpec     interface{}            `json:"toolSpec"`
    EndpointSpec interface{}            `json:"endpointSpec"`
}

type UpdateMcpServerParam struct {
    NamespaceId      string      `json:"namespaceId"`
    McpName          string      `json:"mcpName"`
    IsLatest         bool        `json:"isLatest"`
    ServerSpec       interface{} `json:"serverSpec"`
    ToolSpec         interface{} `json:"toolSpec"`
    EndpointSpec     interface{} `json:"endpointSpec"`
    OverrideExisting bool        `json:"overrideExisting"`
}

type DeleteMcpServerParam struct {
    NamespaceId string `param:"namespaceId"`
    McpName     string `param:"mcpName"`
    McpId       string `param:"mcpId"`
    Version     string `param:"version"`
}

// --- AI A2A ---
type RegisterAgentParam struct {
    AgentCard interface{} `json:"agentCard"`
}

type GetMaintainerAgentCardParam struct {
    AgentName string `param:"agentName"`
    Version   string `param:"version"`
}

type UpdateAgentCardParam struct {
    AgentCard interface{} `json:"agentCard"`
}

type DeleteAgentParam struct {
    AgentName string `param:"agentName"`
    Version   string `param:"version"`
}

type ListAgentVersionParam struct {
    AgentName string `param:"agentName"`
}

type SearchAgentParam struct {
    AgentName string `param:"agentName"`
    PageNo    int    `param:"pageNo"`
    PageSize  int    `param:"pageSize"`
}

type ListAgentCardsParam struct {
    PageNo   int `param:"pageNo"`
    PageSize int `param:"pageSize"`
}
```

---

## 六、开发完成后完整目录结构

```
nacos-sdk-go/
├── go.mod
├── go.sum
├── README.md
│
├── api/grpc/                                     # [现有] gRPC protobuf 生成代码
│   ├── nacos_grpc_service.pb.go
│   └── nacos_grpc_service_grpc.pb.go
│
├── clients/                                      # 客户端实现
│   ├── client_factory.go                         # [修改] 新增 3 个 Maintainer 工厂方法
│   ├── client_factory_test.go                    # [现有]
│   │
│   ├── nacos_client/                             # [现有] Base Client
│   │   ├── nacos_client.go
│   │   └── nacos_client_interface.go
│   │
│   ├── config_client/                            # [现有] 数据面 Config Client
│   │   ├── config_client.go
│   │   ├── config_client_interface.go
│   │   ├── config_proxy.go
│   │   ├── config_proxy_interface.go
│   │   ├── config_connection_event_listener.go
│   │   ├── config_connection_event_listener_test.go
│   │   ├── config_client_test.go
│   │   └── limiter.go
│   │
│   ├── naming_client/                            # [现有] 数据面 Naming Client
│   │   ├── naming_client.go
│   │   ├── naming_client_interface.go
│   │   ├── naming_client_test.go
│   │   ├── naming_proxy_delegate.go
│   │   ├── naming_instance_chooser.go
│   │   ├── service_info_updater.go
│   │   ├── naming_proxy/
│   │   ├── naming_http/
│   │   ├── naming_grpc/
│   │   ├── naming_cache/
│   │   └── redo/
│   │
│   ├── ai_client/                                # [现有] 数据面 AI Client
│   │   ├── ai_client.go
│   │   ├── ai_client_interface.go
│   │   ├── ai_proxy.go
│   │   ├── ai_cache/
│   │   └── redo/
│   │
│   ├── maintainer_client/                        # [新增] ===== 管控面 Maintainer Client =====
│   │   │
│   │   ├── remote/                               # [新增] Admin HTTP 代理层
│   │   │   └── maintainer_http_proxy.go          #        请求构建、JSON body、鉴权、重试
│   │   │
│   │   ├── core/                                 # [新增] Core 管理（公共能力）
│   │   │   ├── core_maintainer_interface.go       #        ICoreMaintainerClient 接口定义
│   │   │   └── core_maintainer_client.go          #        实现: namespace/cluster/loader/health/plugin
│   │   │
│   │   ├── naming/                               # [新增] Naming 管理
│   │   │   ├── naming_maintainer_interface.go     #        INamingMaintainerClient 接口定义
│   │   │   └── naming_maintainer_client.go        #        实现: service/instance/client/health
│   │   │
│   │   ├── config/                               # [新增] Config 管理
│   │   │   ├── config_maintainer_interface.go     #        IConfigMaintainerClient 接口定义
│   │   │   └── config_maintainer_client.go        #        实现: CRUD/history/beta/listener/ops
│   │   │
│   │   └── ai/                                   # [新增] AI 管理
│   │       ├── ai_maintainer_interface.go         #        IAiMaintainerClient 接口定义
│   │       └── ai_maintainer_client.go            #        实现: MCP server/A2A agent 管理
│   │
│   └── cache/                                    # [现有] 缓存工具
│       ├── concurrent_map.go
│       ├── const.go
│       └── disk_cache.go
│
├── common/                                       # 公共基础设施
│   ├── constant/
│   │   ├── config.go                             # [现有] ClientConfig, ServerConfig
│   │   ├── const.go                              # [现有] 常量
│   │   ├── admin_api_path.go                     # [新增] Admin API 路径常量
│   │   ├── ai_constant.go                        # [现有]
│   │   ├── client_config_options.go              # [现有]
│   │   ├── server_config_options.go              # [现有]
│   │   ├── server_tls_options.go                 # [现有]
│   │   └── preserved_metadata_keys.go            # [现有]
│   │
│   ├── http_agent/                               # [现有] HTTP 传输
│   │   ├── http_agent_interface.go               # [现有] IHttpAgent 接口
│   │   ├── http_agent.go                         # [现有]
│   │   ├── get.go                                # [现有]
│   │   ├── post.go                               # [现有]
│   │   ├── put.go                                # [现有]
│   │   └── delete.go                             # [现有]
│   │
│   ├── nacos_server/
│   │   ├── nacos_server.go                       # [修改] 新增 ReqAdminApi 方法
│   │   └── nacos_server_test.go                  # [现有]
│   │
│   ├── security/                                 # [现有] 鉴权
│   │   ├── security_proxy.go                     # [修改] 新增 BuildAdminResource
│   │   ├── nacos_auth_client.go                  # [现有]
│   │   ├── ram_auth_client.go                    # [现有]
│   │   ├── ram_credential_provider.go            # [现有]
│   │   ├── resource_injector.go                  # [现有]
│   │   └── signature_util.go                     # [现有]
│   │
│   ├── nacos_error/                              # [现有]
│   │   └── nacos_error.go
│   │
│   ├── remote/rpc/                               # [现有] gRPC (Maintainer 不使用)
│   │   └── ...
│   │
│   ├── logger/                                   # [现有]
│   ├── monitor/                                  # [现有]
│   ├── encoding/                                 # [现有]
│   ├── encryption/                               # [现有]
│   ├── filter/                                   # [现有]
│   ├── file/                                     # [现有]
│   ├── tls/                                      # [现有]
│   └── redo/                                     # [现有]
│
├── vo/                                           # 值对象（入参）
│   ├── client_param.go                           # [现有]
│   ├── config_param.go                           # [现有]
│   ├── service_param.go                          # [现有]
│   ├── ai_param.go                               # [现有]
│   └── maintainer_param.go                       # [新增] 管控面入参 VO
│
├── model/                                        # 数据模型（返回值）
│   ├── config.go                                 # [现有]
│   ├── service.go                                # [现有]
│   ├── ai_mcp.go                                 # [现有]
│   ├── ai_agent.go                               # [现有]
│   └── maintainer.go                             # [新增] 管控面 Model + 泛型 Page/RestResult
│
├── mock/                                         # [现有] Mock 接口
│   ├── mock_config_client_interface.go
│   ├── mock_http_agent_interface.go
│   ├── mock_nacos_client_interface.go
│   └── mock_naming_client_interface.go
│
├── util/                                         # [现有] 工具
│   ├── common.go
│   ├── content.go
│   ├── md5.go
│   ├── object2param.go
│   └── semaphore.go
│
├── inner/uuid/                                   # [现有] UUID
│
├── example/                                      # 示例
│   ├── config/                                   # [现有]
│   ├── service/                                  # [现有]
│   ├── ai/                                       # [现有]
│   └── maintainer/                               # [新增] Maintainer 示例
│       └── main.go
│
└── docs/
    └── maintainer-client-design.md               # 本文档
```

### 文件变更统计

| 类型 | 文件数 | 说明 |
|------|--------|------|
| **新增** | ~13 个 | maintainer_client/ 目录 8 个 + model 1 个 + vo 1 个 + constant 1 个 + example 1 个 + doc 1 个 |
| **修改** | ~3 个 | client_factory.go、nacos_server.go、security_proxy.go |
| **不变** | 全部现有文件 | 数据面客户端代码不受影响 |

---

## 七、关键设计决策

### 7.1 MaintainerHttpProxy 内部处理 JSON body，不修改 IHttpAgent

**原因**：
- `IHttpAgent` 是公共接口，已有 mock，修改影响面大
- Maintainer 的 JSON body 需求相对特定（仅 clone config、batch metadata 等少数接口）
- 在 proxy 层用 `net/http` 直接构造 JSON 请求，代码局部、好维护

### 7.2 Core 能力通过 struct 嵌入复用

**原因**：
- Go 没有 Java 的 interface default method
- 通过 `CoreMaintainerClient` struct 嵌入 Naming/Config/AI Client
- 所有 Core 方法只实现一次，三种 Client 自动获得

### 7.3 Admin 鉴权复用 Config 类型签名

**原因**：
- 观察 Java 实现，Admin 接口的 RAM 签名逻辑与 Config 类型一致
- `BuildAdminResource` 内部使用 `REQUEST_TYPE_CONFIG`，复用 `ConfigResourceInjector`
- 无需新增 resource type 和 injector

### 7.4 泛型 Page 和 RestResult

**原因**：
- go.mod 已是 `go 1.24.4`，完全支持泛型
- `Page[T]` 和 `RestResult[T]` 避免为每种类型定义 Page struct
- 响应解析统一使用 `ParseResult[T](response)` 泛型函数

### 7.5 Maintainer VO 与数据面 VO 分离

**原因**：
- 管控面参数与数据面参数含义不同（如 maintainer 可以操作任意 namespace，数据面绑定当前 namespace）
- 避免用户混淆，如 `vo.RegisterInstanceParam`（数据面） vs `vo.MaintainerInstanceParam`（管控面）
- Model 层部分类型可复用（如 `model.Instance`），但管控面返回更丰富字段

---

## 八、实现优先级

| 阶段 | 内容 | 依赖 | 预估工作量 |
|------|------|------|-----------|
| **P0** | 基础设施：`admin_api_path.go` + `MaintainerHttpProxy` + `RestResult`/`Page` 泛型 + `NacosServer.ReqAdminApi` + `BuildAdminResource` | 无 | 1~2 天 |
| **P1** | `CoreMaintainerClient` 完整实现（namespace、cluster、loader、health、plugin）| P0 | 1~2 天 |
| **P2** | `NamingMaintainerClient` 完整实现 + 工厂方法 | P0, P1 | 2~3 天 |
| **P3** | `ConfigMaintainerClient` 完整实现 + 工厂方法 | P0, P1 | 2~3 天 |
| **P4** | `AiMaintainerClient` 完整实现 + 工厂方法 | P0, P1 | 1~2 天 |
| **P5** | Example 示例 + 单元测试 | P0~P4 | 2 天 |

---

## 九、使用示例

```go
package main

import (
    "fmt"
    "github.com/nacos-group/nacos-sdk-go/v2/clients"
    "github.com/nacos-group/nacos-sdk-go/v2/common/constant"
    "github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func main() {
    sc := []constant.ServerConfig{
        *constant.NewServerConfig("127.0.0.1", 8848),
    }
    cc := *constant.NewClientConfig(
        constant.WithNamespaceId("public"),
        constant.WithUsername("nacos"),
        constant.WithPassword("nacos"),
    )

    // 创建 Naming Maintainer Client
    namingMaintainer, _ := clients.NewNamingMaintainerClient(vo.NacosClientParam{
        ClientConfig:  &cc,
        ServerConfigs: sc,
    })
    defer namingMaintainer.CloseClient()

    // Namespace 管理（Core 能力）
    namespaces, _ := namingMaintainer.GetNamespaceList()
    fmt.Printf("Namespaces: %+v\n", namespaces)

    namingMaintainer.CreateNamespace(vo.CreateNamespaceParam{
        NamespaceId:   "dev",
        NamespaceName: "开发环境",
        NamespaceDesc: "开发测试使用",
    })

    // 服务管理
    namingMaintainer.CreateService(vo.MaintainerServiceParam{
        NamespaceId: "dev",
        GroupName:   "DEFAULT_GROUP",
        ServiceName: "my-service",
    })

    services, _ := namingMaintainer.ListServices(vo.ListServicesParam{
        NamespaceId: "dev",
        PageNo:      1,
        PageSize:    100,
    })
    fmt.Printf("Services: %+v\n", services)

    // 创建 Config Maintainer Client
    configMaintainer, _ := clients.NewConfigMaintainerClient(vo.NacosClientParam{
        ClientConfig:  &cc,
        ServerConfigs: sc,
    })
    defer configMaintainer.CloseClient()

    // 配置管理
    configMaintainer.PublishConfig(vo.MaintainerPublishConfigParam{
        DataId:  "app.properties",
        Group:   "DEFAULT_GROUP",
        Content: "key=value",
        Type:    "properties",
    })

    configs, _ := configMaintainer.ListConfigs(vo.SearchConfigParam{
        NamespaceId: "public",
        PageNo:      1,
        PageSize:    100,
        Search:      "accurate",
    })
    fmt.Printf("Configs: %+v\n", configs)
}
```

---

## 十、开发执行步骤

### Step 1：类型定义（常量 + Model + VO）

**目标**：定义所有数据结构，后续步骤直接引用，不再反复改类型。

**新建文件**：

| 文件 | 内容 |
|------|------|
| `common/constant/admin_api_path.go` | 18 个 `/v3/admin/...` 路径常量 |
| `model/maintainer.go` | 泛型 `Page[T]`、`RestResult[T]`，以及全部管控面 Model（约 20 个 struct） |
| `vo/maintainer_param.go` | 全部管控面入参 VO（约 30 个 struct） |

**产出物**：3 个新文件。

**验证方式**：`go build ./...` 编译通过，无类型错误。

---

### Step 2：扩展现有基础设施（NacosServer + SecurityProxy）

**目标**：在现有组件上做最小改动，为 Maintainer 提供 HTTP 底层支持。

**修改文件**：

| 文件 | 改动 |
|------|------|
| `common/security/security_proxy.go` | 新增 `BuildAdminResource(namespace, group, resource string) RequestResource` 函数（约 5 行） |
| `common/nacos_server/nacos_server.go` | 新增 `callAdminServer` 内部方法 + `ReqAdminApi` 公开方法 |

`callAdminServer` 关键逻辑：
- 与 `callConfigServer` 结构一致（复用重试、failover）
- 增加判断：当 `body []byte` 不为 nil 时，用 `bytes.NewReader(body)` 构造 `http.NewRequest`，设 `Content-Type: application/json`
- 否则走原有 form-urlencoded 路径

`ReqAdminApi` 签名：
```go
func (server *NacosServer) ReqAdminApi(api string, params map[string]string,
    headers map[string]string, body []byte, method string,
    resource security.RequestResource, timeoutMS uint64) (string, error)
```

**产出物**：2 个文件各新增一个方法。

**验证方式**：`go test ./common/...` 现有测试全部通过，新方法编译无误。

---

### Step 3：MaintainerHttpProxy 实现

**目标**：封装 admin 请求统一入口，屏蔽鉴权/重试/解析细节。

**新建文件**：

| 文件 | 内容 |
|------|------|
| `clients/maintainer_client/remote/maintainer_http_proxy.go` | `MaintainerHttpProxy` struct + 方法 |

核心结构：

```
MaintainerHttpProxy
├── nacosServer  *nacos_server.NacosServer   // 复用
├── clientCfg    constant.ClientConfig       // 复用
│
├── NewMaintainerHttpProxy(ctx, serverCfgs, clientCfg, httpAgent, provider)
│
├── ReqApi(api, params, method) (string, error)               // form-urlencoded
├── ReqApiWithBody(api, params, body, method) (string, error)  // JSON body
│
├── ParseResult[T](response) (T, error)        // 泛型响应解析
├── ParseBoolResult(response) (bool, error)
├── ParseStringResult(response) (string, error)
└── Close()
```

内部逻辑：
- `ReqApi` 调用 `nacosServer.ReqAdminApi(api, params, headers, nil, method, resource, timeout)`
- `ReqApiWithBody` 将 body 用 `json.Marshal` 序列化后传入
- `ParseResult[T]` 先 unmarshal 为 `RestResult[T]`，code != 0 时返回 `NacosError`

**产出物**：1 个新文件。

**验证方式**：`go build ./clients/maintainer_client/...` 编译通过。

---

### Step 4：CoreMaintainerClient 实现

**目标**：实现所有 Maintainer 共享的 Core 管理能力（25 个方法）。

**新建文件**：

| 文件 | 内容 |
|------|------|
| `clients/maintainer_client/core/core_maintainer_interface.go` | `ICoreMaintainerClient` 接口定义 |
| `clients/maintainer_client/core/core_maintainer_client.go` | `CoreMaintainerClient` struct + 全部实现 |

实现方法按 admin API 分组：

| 分组 | 方法 | HTTP | 路径 |
|------|------|------|------|
| 状态 | `GetServerState` | GET | `/v3/admin/core/state` |
| | `Liveness` | GET | `/v3/admin/core/state/liveness` |
| | `Readiness` | GET | `/v3/admin/core/ops/readiness` |
| Raft | `RaftOps` | POST | `/v3/admin/core/ops/raft` |
| ID | `GetIdGenerators` | GET | `/v3/admin/core/ops/ids` |
| 日志 | `UpdateLogLevel` | PUT | `/v3/admin/core/ops/log` |
| 集群 | `ListClusterNodes` | GET | `/v3/admin/core/cluster/node/list` |
| | `UpdateLookupMode` | PUT | `/v3/admin/core/cluster/lookup` |
| Loader | `GetCurrentClients` | GET | `/v3/admin/core/loader/current` |
| | `ReloadConnectionCount` | POST | `/v3/admin/core/loader/reloadCurrent` |
| | `SmartReloadCluster` | POST | `/v3/admin/core/loader/smartReloadCluster` |
| | `ReloadSingleClient` | POST | `/v3/admin/core/loader/reloadClient` |
| | `GetClusterLoaderMetrics` | GET | `/v3/admin/core/loader/cluster` |
| Namespace | `GetNamespaceList` | GET | `/v3/admin/core/namespace/list` |
| | `GetNamespace` | GET | `/v3/admin/core/namespace` |
| | `CreateNamespace` | POST | `/v3/admin/core/namespace` |
| | `UpdateNamespace` | PUT | `/v3/admin/core/namespace` |
| | `DeleteNamespace` | DELETE | `/v3/admin/core/namespace` |
| | `CheckNamespaceIdExist` | GET | `/v3/admin/core/namespace/check` |
| Plugin | `ListPlugins` | GET | `/v3/admin/core/plugin/list` |
| | `GetPluginDetail` | GET | `/v3/admin/core/plugin/{type}/{name}` |
| | `UpdatePluginStatus` | PUT | `/v3/admin/core/plugin/{type}/{name}/status` |
| | `UpdatePluginConfig` | PUT | `/v3/admin/core/plugin/{type}/{name}/config` |
| | `GetPluginAvailability` | GET | `/v3/admin/core/plugin/{type}/{name}/availability` |
| 生命周期 | `CloseClient` | — | cancel context、释放资源 |

每个方法的实现模式统一：
```go
func (c *CoreMaintainerClient) GetNamespaceList() ([]model.Namespace, error) {
    resp, err := c.proxy.ReqApi(constant.AdminCoreNamespacePath+"/list",
        nil, http.MethodGet)
    if err != nil {
        return nil, err
    }
    return remote.ParseResult[[]model.Namespace](resp)
}
```

**产出物**：2 个新文件（接口 + 实现）。

**验证方式**：编译通过 + 启动 Nacos Server 手动验证 Namespace CRUD。

---

### Step 5：NamingMaintainerClient 实现

**目标**：实现服务发现管控能力（约 28 个方法）。

**新建文件**：

| 文件 | 内容 |
|------|------|
| `clients/maintainer_client/naming/naming_maintainer_interface.go` | `INamingMaintainerClient` 接口（嵌入 `ICoreMaintainerClient`） |
| `clients/maintainer_client/naming/naming_maintainer_client.go` | `NamingMaintainerClient` struct（嵌入 `*CoreMaintainerClient`）+ 实现 |

实现方法按分组：

| 分组 | 方法 | HTTP | 路径 |
|------|------|------|------|
| Service | `CreateService` | POST | `/v3/admin/ns/service` |
| | `UpdateService` | PUT | `/v3/admin/ns/service` |
| | `RemoveService` | DELETE | `/v3/admin/ns/service` |
| | `GetServiceDetail` | GET | `/v3/admin/ns/service` |
| | `ListServices` | GET | `/v3/admin/ns/service/list` |
| | `ListServicesWithDetail` | GET | `/v3/admin/ns/service/list` (withInstances) |
| | `GetSubscribers` | GET | `/v3/admin/ns/service/subscribers` |
| | `ListSelectorTypes` | GET | `/v3/admin/ns/service/selector/types` |
| Instance | `RegisterInstance` | POST | `/v3/admin/ns/instance` |
| | `DeregisterInstance` | DELETE | `/v3/admin/ns/instance` |
| | `UpdateInstance` | PUT | `/v3/admin/ns/instance` |
| | `BatchUpdateInstanceMetadata` | PUT | `/v3/admin/ns/instance/metadata/batch` |
| | `BatchDeleteInstanceMetadata` | DELETE | `/v3/admin/ns/instance/metadata/batch` |
| | `PartialUpdateInstance` | PUT | `/v3/admin/ns/instance/partial` |
| | `ListInstances` | GET | `/v3/admin/ns/instance/list` |
| | `GetInstanceDetail` | GET | `/v3/admin/ns/instance` |
| Client | `GetClientList` | GET | `/v3/admin/ns/client/list` |
| | `GetClientDetail` | GET | `/v3/admin/ns/client` |
| | `GetPublishedServiceList` | GET | `/v3/admin/ns/client/publish/list` |
| | `GetSubscribeServiceList` | GET | `/v3/admin/ns/client/subscribe/list` |
| | `GetPublishedClientList` | GET | `/v3/admin/ns/client/service/publisher/list` |
| | `GetSubscribeClientList` | GET | `/v3/admin/ns/client/service/subscriber/list` |
| Health | `GetMetrics` | GET | `/v3/admin/ns/ops/metrics` |
| | `UpdateInstanceHealthStatus` | PUT | `/v3/admin/ns/health/instance` |
| | `GetHealthCheckers` | GET | `/v3/admin/ns/health/checkers` |
| | `UpdateCluster` | PUT | `/v3/admin/ns/cluster` |

构造函数：
```go
func NewNamingMaintainerClient(nc nacos_client.INacosClient,
    provider security.RamCredentialProvider) (*NamingMaintainerClient, error)
```
内部创建 `CoreMaintainerClient`（含 `MaintainerHttpProxy`），嵌入自身。

**产出物**：2 个新文件。

**验证方式**：编译通过 + 手动验证 Service/Instance CRUD。

---

### Step 6：ConfigMaintainerClient 实现

**目标**：实现配置管控能力（约 18 个方法）。

**新建文件**：

| 文件 | 内容 |
|------|------|
| `clients/maintainer_client/config/config_maintainer_interface.go` | `IConfigMaintainerClient` 接口 |
| `clients/maintainer_client/config/config_maintainer_client.go` | 实现 |

实现方法按分组：

| 分组 | 方法 | HTTP | 路径 |
|------|------|------|------|
| CRUD | `GetConfig` | GET | `/v3/admin/cs/config` |
| | `PublishConfig` | POST | `/v3/admin/cs/config` |
| | `UpdateConfigMetadata` | PUT | `/v3/admin/cs/config` |
| | `DeleteConfig` | DELETE | `/v3/admin/cs/config` |
| | `DeleteConfigs` | DELETE | `/v3/admin/cs/config` (ids) |
| Search | `ListConfigs` | GET | `/v3/admin/cs/config/list` (search=accurate) |
| | `SearchConfigs` | GET | `/v3/admin/cs/config/list` (search=blur) |
| | `GetConfigListByNamespace` | GET | `/v3/admin/cs/config/list` |
| Listener | `GetListeners` | GET | `/v3/admin/cs/listener` |
| | `GetAllSubClientConfigByIp` | GET | `/v3/admin/cs/listener` |
| Clone | `CloneConfig` | POST | `/v3/admin/cs/config` (JSON body) |
| Beta | `PublishBetaConfig` | POST | `/v3/admin/cs/config` (beta) |
| | `StopBeta` | DELETE | `/v3/admin/cs/config` (beta) |
| | `QueryBeta` | GET | `/v3/admin/cs/config` (beta) |
| History | `ListConfigHistory` | GET | `/v3/admin/cs/history` |
| | `GetConfigHistoryInfo` | GET | `/v3/admin/cs/history` |
| | `GetPreviousConfigHistoryInfo` | GET | `/v3/admin/cs/history/previous` |
| Ops | `UpdateLocalCacheFromStore` | POST | `/v3/admin/cs/ops/localCache` |

**产出物**：2 个新文件。

**验证方式**：编译通过 + 手动验证 Config 发布/查询/历史。

---

### Step 7：AiMaintainerClient 实现

**目标**：实现 AI（MCP/A2A）管控能力（约 13 个方法）。

**新建文件**：

| 文件 | 内容 |
|------|------|
| `clients/maintainer_client/ai/ai_maintainer_interface.go` | `IAiMaintainerClient` 接口 |
| `clients/maintainer_client/ai/ai_maintainer_client.go` | 实现 |

实现方法按分组：

| 分组 | 方法 | HTTP | 路径 |
|------|------|------|------|
| MCP | `ListMcpServer` | GET | `/v3/admin/ai/mcp/list` |
| | `SearchMcpServer` | GET | `/v3/admin/ai/mcp/search` |
| | `GetMcpServerDetail` | GET | `/v3/admin/ai/mcp` |
| | `CreateMcpServer` | POST | `/v3/admin/ai/mcp` (JSON body) |
| | `UpdateMcpServer` | PUT | `/v3/admin/ai/mcp` (JSON body) |
| | `DeleteMcpServer` | DELETE | `/v3/admin/ai/mcp` |
| A2A | `RegisterAgent` | POST | `/v3/admin/ai/a2a` (JSON body) |
| | `GetAgentCard` | GET | `/v3/admin/ai/a2a` |
| | `UpdateAgentCard` | PUT | `/v3/admin/ai/a2a` (JSON body) |
| | `DeleteAgent` | DELETE | `/v3/admin/ai/a2a` |
| | `ListAllVersionOfAgent` | GET | `/v3/admin/ai/a2a/version/list` |
| | `SearchAgentCardsByName` | GET | `/v3/admin/ai/a2a/list` |
| | `ListAgentCards` | GET | `/v3/admin/ai/a2a/list` |

**产出物**：2 个新文件。

**验证方式**：编译通过 + 手动验证 MCP Server 创建/查询。

---

### Step 8：集成 — 工厂方法 + 示例

**目标**：对外暴露统一入口，提供完整使用示例。

**修改文件**：

| 文件 | 改动 |
|------|------|
| `clients/client_factory.go` | 新增 `NewNamingMaintainerClient`、`NewConfigMaintainerClient`、`NewAiMaintainerClient` 三个工厂函数，内部复用 `setConfig(param)` 创建 NacosClient |

**新建文件**：

| 文件 | 内容 |
|------|------|
| `example/maintainer/main.go` | 完整使用示例：创建 client → namespace CRUD → service CRUD → config CRUD → 关闭 |

工厂方法实现模式（与现有 NewConfigClient、NewNamingClient 一致）：
```go
func NewNamingMaintainerClient(param vo.NacosClientParam) (
    naming.INamingMaintainerClient, error) {
    nacosClient, err := setConfig(param)    // 复用
    if err != nil {
        return nil, err
    }
    return naming.NewNamingMaintainerClient(nacosClient, param.RamCredentialProvider)
}
```

**产出物**：1 个修改文件 + 1 个新文件。

**最终验证**：
1. `go build ./...` —— 全量编译通过
2. `go test ./...` —— 全量测试通过（含现有测试不受影响）
3. `go vet ./...` —— 无代码问题
4. 启动本地 Nacos Server，运行 `go run example/maintainer/main.go` —— 端到端验证

---

### 总体变更统计

| 步骤 | 新增文件 | 修改文件 | 新增代码量预估 |
|------|----------|----------|----------------|
| Step 1 | 3 | 0 | ~400 行 |
| Step 2 | 0 | 2 | ~80 行 |
| Step 3 | 1 | 0 | ~150 行 |
| Step 4 | 2 | 0 | ~350 行 |
| Step 5 | 2 | 0 | ~400 行 |
| Step 6 | 2 | 0 | ~300 行 |
| Step 7 | 2 | 0 | ~250 行 |
| Step 8 | 1 | 1 | ~100 行 |
| **合计** | **13** | **3** | **~2030 行** |
