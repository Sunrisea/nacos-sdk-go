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

package main

import (
	"fmt"
	"time"

	"github.com/a2aproject/a2a-go/a2a"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/ai_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func main() {
	fmt.Println("Nacos AI Client Example")
	fmt.Println("========================")

	// Create ServerConfig
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", 8848, constant.WithContextPath("/nacos")),
	}

	// Create ClientConfig
	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(""),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithUsername("nacos"),
		constant.WithPassword("nacos"),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("debug"),
	)

	// Create AI client
	client, err := clients.NewAIClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		panic(err)
	}
	defer client.CloseClient()

	// Run MCP Server operations
	mcpServerOperations(client)

	// Run Agent Card operations
	agentCardOperations(client)

	// Run subscription operations
	subscriptionOperations(client)
}

// mcpServerOperations demonstrates MCP server management operations
func mcpServerOperations(client ai_client.IAIClient) {
	fmt.Println("\n--- MCP Server Operations ---")

	// 1. Release (publish) MCP Server
	fmt.Println("\n1. Release MCP Server:")
	serverSpec := model.NewMcpServerBasicInfo()
	serverSpec.Name = "demo-mcp-server"
	serverSpec.Description = "Demo MCP Server for testing"
	serverSpec.Protocol = "mcp-sse"
	serverSpec.FrontProtocol = "mcp-sse"
	serverSpec.VersionDetail = &model.ServerVersionDetail{
		Version: "v1.0.0",
	}
	serverSpec.RemoteServerConfig = &model.McpServerRemoteServiceConfig{
		ExportPath: "/mcp/demo",
	}

	toolSpec := &model.McpToolSpecification{
		Tools: []model.McpTool{
			{
				Name:        "search",
				Description: "Search tool for finding information",
			},
			{
				Name:        "calculate",
				Description: "Calculator tool for math operations",
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

	mcpId, err := client.ReleaseMcpServer(vo.ReleaseMcpServerParam{
		ServerSpec:      serverSpec,
		ToolSpec:        toolSpec,
		McpEndpointSpec: endpointSpec,
	})
	if err != nil {
		fmt.Printf("   Failed to release MCP Server: %v\n", err)
	} else {
		fmt.Printf("   Released MCP Server, mcpId: %s\n", mcpId)
	}

	time.Sleep(500 * time.Millisecond)

	// 2. Get MCP Server information
	fmt.Println("\n2. Get MCP Server:")
	mcpServer, err := client.GetMcpServer(vo.GetMcpServerParam{
		McpName: "demo-mcp-server",
		Version: "v1.0.0",
	})
	if err != nil {
		fmt.Printf("   Failed to get MCP Server: %v\n", err)
	} else {
		fmt.Printf("   Retrieved MCP Server: Name=%s, Version=%s\n", mcpServer.Name, mcpServer.Version)
		fmt.Printf("   Description: %s\n", mcpServer.Description)
	}

	// 3. Register MCP Server endpoint
	fmt.Println("\n3. Register MCP Server Endpoint:")
	err = client.RegisterMcpServerEndpoint(vo.RegisterMcpServerEndpointParam{
		McpName: "demo-mcp-server",
		Address: "192.168.1.100",
		Port:    8080,
		Version: "v1.0.0",
	})
	if err != nil {
		fmt.Printf("   Failed to register endpoint: %v\n", err)
	} else {
		fmt.Println("   Registered MCP Server endpoint successfully")
	}

	time.Sleep(500 * time.Millisecond)

	// 4. Deregister MCP Server endpoint
	fmt.Println("\n4. Deregister MCP Server Endpoint:")
	err = client.DeregisterMcpServerEndpoint(vo.DeregisterMcpServerEndpointParam{
		McpName: "demo-mcp-server",
		Address: "192.168.1.100",
		Port:    8080,
	})
	if err != nil {
		fmt.Printf("   Failed to deregister endpoint: %v\n", err)
	} else {
		fmt.Println("   Deregistered MCP Server endpoint successfully")
	}
}

// agentCardOperations demonstrates A2A Agent Card management operations
func agentCardOperations(client ai_client.IAIClient) {
	fmt.Println("\n--- Agent Card Operations ---")

	// 1. Release (publish) Agent Card
	fmt.Println("\n1. Release Agent Card:")
	agentCard := &a2a.AgentCard{
		Name:            "demo-agent",
		Version:         "v1.0.0",
		ProtocolVersion: "0.2.1",
		Description:     "Demo AI Agent for testing",
		URL:             "https://demo-agent.example.com",
		Skills: []a2a.AgentSkill{
			{
				ID:          "search",
				Name:        "Search Skill",
				Description: "Can search for information",
			},
			{
				ID:          "summarize",
				Name:        "Summarize Skill",
				Description: "Can summarize long texts",
			},
		},
	}

	err := client.ReleaseAgentCard(vo.ReleaseAgentCardParam{
		AgentCard:        agentCard,
		RegistrationType: constant.A2A_ENDPOINT_TYPE_SERVICE,
		SetAsLatest:      true,
	})
	if err != nil {
		fmt.Printf("   Failed to release Agent Card: %v\n", err)
	} else {
		fmt.Println("   Released Agent Card successfully")
	}

	time.Sleep(500 * time.Millisecond)

	// 2. Get Agent Card information
	fmt.Println("\n2. Get Agent Card:")
	agentCardInfo, err := client.GetAgentCard(vo.GetAgentCardParam{
		AgentName:        "demo-agent",
		Version:          "v1.0.0",
		RegistrationType: constant.A2A_ENDPOINT_TYPE_URL,
	})
	if err != nil {
		fmt.Printf("   Failed to get Agent Card: %v\n", err)
	} else {
		fmt.Printf("   Retrieved Agent Card: Name=%s\n", agentCardInfo.Name)
		if agentCardInfo.AgentCard != nil {
			fmt.Printf("   Description: %s, URL: %s\n", agentCardInfo.Description, agentCardInfo.URL)
		}
	}

	// 3. Register Agent endpoint
	fmt.Println("\n3. Register Agent Endpoint:")
	err = client.RegisterAgentEndpoint(vo.RegisterAgentEndpointParam{
		AgentName:  "demo-agent",
		Address:    "192.168.1.100",
		Port:       9090,
		Path:       "/api/v1/agent",
		Transport:  constant.A2A_ENDPOINT_DEFAULT_TRANSPORT,
		SupportTLS: true,
		Version:    "v1.0.0",
	})
	if err != nil {
		fmt.Printf("   Failed to register endpoint: %v\n", err)
	} else {
		fmt.Println("   Registered Agent endpoint successfully")
	}

	time.Sleep(500 * time.Second)

	// 4. Deregister Agent endpoint
	fmt.Println("\n4. Deregister Agent Endpoint:")
	err = client.DeregisterAgentEndpoint(vo.DeregisterAgentEndpointParam{
		AgentName: "demo-agent",
		Address:   "192.168.1.100",
		Port:      9090,
		Version:   "v1.0.0",
	})
	if err != nil {
		fmt.Printf("   Failed to deregister endpoint: %v\n", err)
	} else {
		fmt.Println("   Deregistered Agent endpoint successfully")
	}
}

// subscriptionOperations demonstrates subscription patterns for MCP and Agent changes
func subscriptionOperations(client ai_client.IAIClient) {
	fmt.Println("\n--- Subscription Operations ---")

	// 1. Subscribe to MCP Server changes
	fmt.Println("\n1. Subscribe to MCP Server changes:")
	mcpCallback := func(mcpId, namespaceId, mcpName string, mcpServer model.McpServerDetailInfo) {
		fmt.Printf("   [MCP Callback] Server changed: mcpId=%s, name=%s, backendEndpoints=%d\n",
			mcpId, mcpName, len(mcpServer.BackendEndpoints))
	}

	mcpSubscribeParam := vo.SubscribeMcpServerParam{
		McpName:           "demo-mcp-server",
		Version:           "v1.0.0",
		SubscribeCallback: mcpCallback,
	}
	mcpServer, err := client.SubscribeMcpServer(mcpSubscribeParam)
	if err != nil {
		fmt.Printf("   Failed to subscribe to MCP Server: %v\n", err)
	} else {
		fmt.Printf("   Subscribed to MCP Server: %s\n", mcpServer.Name)
	}

	// 2. Subscribe to Agent Card changes
	fmt.Println("\n2. Subscribe to Agent Card changes:")
	agentCallback := func(agentName string, agentCard model.AgentCardDetailInfo) {
		fmt.Printf("   [Agent Callback] Card changed: name=%s, latestVersion=%v\n",
			agentName, agentCard.LatestVersion)
	}

	agentSubscribeParam := vo.SubscribeAgentCardParam{
		AgentName:         "demo-agent",
		Version:           "v1.0.0",
		SubscribeCallback: agentCallback,
	}
	agentCardInfo, err := client.SubscribeAgentCard(agentSubscribeParam)
	if err != nil {
		fmt.Printf("   Failed to subscribe to Agent Card: %v\n", err)
	} else {
		fmt.Printf("   Subscribed to Agent Card: %s\n", agentCardInfo.Name)
	}

	// 3. Trigger changes to verify subscription callbacks
	fmt.Println("\n3. Trigger changes to verify callbacks:")

	// Register MCP Server endpoint to trigger MCP callback
	fmt.Println("   Registering MCP Server endpoint...")
	err = client.RegisterMcpServerEndpoint(vo.RegisterMcpServerEndpointParam{
		McpName: "demo-mcp-server",
		Address: "192.168.1.200",
		Port:    8080,
		Version: "v1.0.0", // 版本号与订阅时保持一致
	})
	if err != nil {
		fmt.Printf("   Failed to register MCP endpoint: %v\n", err)
	} else {
		fmt.Println("   Registered MCP Server endpoint, waiting for callback...")
	}

	// Register Agent endpoint to trigger Agent callback
	fmt.Println("   Registering Agent endpoint...")
	err = client.RegisterAgentEndpoint(vo.RegisterAgentEndpointParam{
		AgentName:  "demo-agent",
		Address:    "192.168.1.200",
		Port:       9090,
		Path:       "/api/v1/agent",
		Transport:  constant.A2A_ENDPOINT_DEFAULT_TRANSPORT,
		SupportTLS: true,
		Version:    "v1.0.0", // 版本号与订阅时保持一致
	})
	if err != nil {
		fmt.Printf("   Failed to register Agent endpoint: %v\n", err)
	} else {
		fmt.Println("   Registered Agent endpoint, waiting for callback...")
	}

	// Wait for callbacks to be triggered (polling interval is 10 seconds)
	fmt.Println("\n4. Waiting for callbacks (15 seconds, polling interval is 10s)...")
	time.Sleep(15 * time.Second)

	// 5. Cleanup: Deregister endpoints
	fmt.Println("\n5. Cleanup - Deregister endpoints:")
	err = client.DeregisterMcpServerEndpoint(vo.DeregisterMcpServerEndpointParam{
		McpName: "demo-mcp-server",
		Address: "192.168.1.200",
		Port:    8080,
	})
	if err != nil {
		fmt.Printf("   Failed to deregister MCP endpoint: %v\n", err)
	} else {
		fmt.Println("   Deregistered MCP Server endpoint")
	}

	err = client.DeregisterAgentEndpoint(vo.DeregisterAgentEndpointParam{
		AgentName: "demo-agent",
		Address:   "192.168.1.200",
		Port:      9090,
		Version:   "v1.0.0",
	})
	if err != nil {
		fmt.Printf("   Failed to deregister Agent endpoint: %v\n", err)
	} else {
		fmt.Println("   Deregistered Agent endpoint")
	}

	// Wait for deregister callbacks (polling interval is 10 seconds)
	fmt.Println("\n   Waiting for deregister callbacks (15 seconds)...")
	time.Sleep(15 * time.Second)

	// 6. Unsubscribe
	fmt.Println("\n6. Unsubscribe:")
	err = client.UnsubscribeMcpServer(mcpSubscribeParam)
	if err != nil {
		fmt.Printf("   Failed to unsubscribe MCP Server: %v\n", err)
	} else {
		fmt.Println("   Unsubscribed from MCP Server")
	}

	err = client.UnsubscribeAgentCard(agentSubscribeParam)
	if err != nil {
		fmt.Printf("   Failed to unsubscribe Agent Card: %v\n", err)
	} else {
		fmt.Println("   Unsubscribed from Agent Card")
	}

	fmt.Println("\n--- End of Example ---")
}
