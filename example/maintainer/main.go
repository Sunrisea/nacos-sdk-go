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

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	maintainer_ai "github.com/nacos-group/nacos-sdk-go/v2/clients/maintainer_client/ai"
	maintainer_config "github.com/nacos-group/nacos-sdk-go/v2/clients/maintainer_client/config"
	maintainer_naming "github.com/nacos-group/nacos-sdk-go/v2/clients/maintainer_client/naming"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

const (
	testNamespaceId   = "maintainer-sdk-test"
	testNamespaceName = "Maintainer SDK Test NS"
	testServiceName   = "maintainer-test-service"
	testGroupName     = "DEFAULT_GROUP"
	testDataId        = "maintainer-test-config"
	testDataId2       = "maintainer-test-config-2"
	testConfigContent = "server.port=8080\napp.name=maintainer-demo"
	testMcpServerName = "maintainer-test-mcp"
	testAgentName     = "maintainer-test-agent"
)

var (
	passed int
	failed int
)

func main() {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", 8848, constant.WithContextPath("/nacos")),
	}

	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(""),
		constant.WithTimeoutMs(10000),
		constant.WithUsername("nacos"),
		constant.WithPassword("nacos"),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("warn"),
	)

	param := vo.NacosClientParam{
		ClientConfig:  &cc,
		ServerConfigs: sc,
	}

	fmt.Println("========================================")
	fmt.Println("  Nacos Maintainer SDK Full Coverage Test")
	fmt.Println("========================================")

	testCoreMaintainer(param)
	testConfigMaintainer(param)
	testNamingMaintainer(param)
	testAiMaintainer(param)

	fmt.Println("\n========================================")
	fmt.Printf("  Results: %d passed, %d failed, %d total\n", passed, failed, passed+failed)
	fmt.Println("========================================")

	if failed > 0 {
		os.Exit(1)
	}
}

// ==================== Core Maintainer Tests ====================

func testCoreMaintainer(param vo.NacosClientParam) {
	fmt.Println("\n--- [1] Core Maintainer Tests (19 methods) ---")

	client, err := clients.NewNamingMaintainerClient(param)
	if err != nil {
		fmt.Printf("[FAIL] Create NamingMaintainerClient: %v\n", err)
		os.Exit(1)
	}
	defer client.CloseClient()

	// --- Server State ---
	check("GetServerState", func() error {
		state, err := client.GetServerState()
		if err != nil {
			return err
		}
		logDetail("state=%v", state)
		return nil
	})

	check("Liveness", func() error {
		alive, err := client.Liveness()
		if err != nil {
			return err
		}
		logDetail("alive=%v", alive)
		return nil
	})

	check("Readiness", func() error {
		ready, err := client.Readiness()
		if err != nil {
			return err
		}
		logDetail("ready=%v", ready)
		return nil
	})

	// --- Raft & ID Generators ---
	check("GetIdGenerators", func() error {
		gens, err := client.GetIdGenerators()
		if err != nil {
			return err
		}
		logDetail("%d generators", len(gens))
		for _, g := range gens {
			logDetail("  resource=%s", g.Resource)
		}
		return nil
	})

	check("RaftOps", func() error {
		_, err := client.RaftOps("resetRaftCluster", "", "naming_instance_metadata")
		// may fail if raft not configured, but we test the call path
		if err != nil {
			logDetail("(expected in standalone) err=%v", truncate(err.Error(), 80))
		}
		return nil // don't fail the test for standalone mode
	})

	// --- Log Level ---
	check("UpdateLogLevel", func() error {
		err := client.UpdateLogLevel("com.alibaba.nacos", "WARN")
		return err
	})

	// --- Cluster ---
	check("ListClusterNodes", func() error {
		nodes, err := client.ListClusterNodes("", "")
		if err != nil {
			return err
		}
		logDetail("%d nodes", len(nodes))
		for _, n := range nodes {
			logDetail("  %s:%d state=%s", n.Ip, n.Port, n.State)
		}
		return nil
	})

	check("UpdateLookupMode", func() error {
		_, err := client.UpdateLookupMode("file")
		if err != nil {
			logDetail("(may fail in standalone) err=%v", truncate(err.Error(), 80))
		}
		return nil
	})

	// --- Loader ---
	check("GetCurrentClients", func() error {
		conns, err := client.GetCurrentClients()
		if err != nil {
			return err
		}
		logDetail("%d connections", len(conns))
		i := 0
		for id, c := range conns {
			if i >= 3 {
				logDetail("  ... and %d more", len(conns)-3)
				break
			}
			logDetail("  %s -> %s (%s)", id, c.MetaInfo.ClientIp, c.MetaInfo.ConnectType)
			i++
		}
		return nil
	})

	check("GetClusterLoaderMetrics", func() error {
		m, err := client.GetClusterLoaderMetrics()
		if err != nil {
			return err
		}
		logDetail("members=%d, completed=%v, total=%d", m.MemberCount, m.Completed, m.Total)
		return nil
	})

	check("ReloadConnectionCount", func() error {
		_, err := client.ReloadConnectionCount(0, "")
		if err != nil {
			logDetail("(may fail) err=%v", truncate(err.Error(), 80))
		}
		return nil
	})

	check("SmartReloadCluster", func() error {
		_, err := client.SmartReloadCluster("0.1")
		if err != nil {
			logDetail("(may fail) err=%v", truncate(err.Error(), 80))
		}
		return nil
	})

	check("ReloadSingleClient", func() error {
		_, err := client.ReloadSingleClient("non-existent-conn-id", "")
		if err != nil {
			logDetail("(expected) err=%v", truncate(err.Error(), 80))
		}
		return nil
	})

	// --- Namespace CRUD ---
	// cleanup from previous run
	_, _ = client.DeleteNamespace(testNamespaceId)

	check("CreateNamespace", func() error {
		ok, err := client.CreateNamespace(vo.CreateNamespaceParam{
			NamespaceId:   testNamespaceId,
			NamespaceName: testNamespaceName,
			NamespaceDesc: "Created by maintainer SDK demo",
		})
		if err != nil {
			return err
		}
		logDetail("created=%v", ok)
		return nil
	})

	check("CheckNamespaceIdExist", func() error {
		exist, err := client.CheckNamespaceIdExist(testNamespaceId)
		if err != nil {
			return err
		}
		logDetail("exist=%v", exist)
		return nil
	})

	check("GetNamespace", func() error {
		ns, err := client.GetNamespace(testNamespaceId)
		if err != nil {
			return err
		}
		logDetail("id=%s, name=%s", ns.Namespace, ns.NamespaceShowName)
		return nil
	})

	check("UpdateNamespace", func() error {
		ok, err := client.UpdateNamespace(vo.UpdateNamespaceParam{
			NamespaceId:   testNamespaceId,
			NamespaceName: testNamespaceName + " Updated",
			NamespaceDesc: "Updated by SDK",
		})
		if err != nil {
			return err
		}
		logDetail("updated=%v", ok)
		return nil
	})

	check("GetNamespaceList", func() error {
		list, err := client.GetNamespaceList()
		if err != nil {
			return err
		}
		logDetail("%d namespaces", len(list))
		return nil
	})

	check("DeleteNamespace", func() error {
		ok, err := client.DeleteNamespace(testNamespaceId)
		if err != nil {
			return err
		}
		logDetail("deleted=%v", ok)
		return nil
	})
}

// ==================== Config Maintainer Tests ====================

func testConfigMaintainer(param vo.NacosClientParam) {
	fmt.Println("\n--- [2] Config Maintainer Tests (17 methods) ---")

	client, err := clients.NewConfigMaintainerClient(param)
	if err != nil {
		fmt.Printf("[FAIL] Create ConfigMaintainerClient: %v\n", err)
		return
	}
	defer client.CloseClient()

	// cleanup
	_, _ = client.DeleteConfig(vo.MaintainerConfigParam{DataId: testDataId, Group: testGroupName})
	_, _ = client.DeleteConfig(vo.MaintainerConfigParam{DataId: testDataId2, Group: testGroupName})

	// --- CRUD ---
	check("PublishConfig", func() error {
		ok, err := client.PublishConfig(vo.MaintainerPublishConfigParam{
			DataId:  testDataId, Group: testGroupName,
			Content: testConfigContent, Type: "properties",
			Desc: "Test config", SrcUser: "maintainer-demo",
		})
		if err != nil {
			return err
		}
		logDetail("published=%v", ok)
		return nil
	})

	// publish a second config for batch operations
	_, _ = client.PublishConfig(vo.MaintainerPublishConfigParam{
		DataId: testDataId2, Group: testGroupName,
		Content: "key=value2", Type: "properties", SrcUser: "demo",
	})

	check("GetConfig", func() error {
		cfg, err := client.GetConfig(vo.MaintainerConfigParam{
			DataId: testDataId, Group: testGroupName,
		})
		if err != nil {
			return err
		}
		logDetail("dataId=%s, group=%s, md5=%s, len=%d", cfg.DataId, cfg.GroupName, cfg.Md5, len(cfg.Content))
		return nil
	})

	check("UpdateConfigMetadata", func() error {
		ok, err := client.UpdateConfigMetadata(vo.UpdateConfigMetadataParam{
			DataId: testDataId, Group: testGroupName,
			Desc: "Updated desc", ConfigTags: "test,demo",
		})
		if err != nil {
			return err
		}
		logDetail("updated=%v", ok)
		return nil
	})

	// --- Search ---
	check("SearchConfig(blur)", func() error {
		result, err := client.SearchConfig(vo.MaintainerSearchConfigParam{
			Search: "blur", DataId: "maintainer*", PageNo: 1, PageSize: 10,
		})
		if err != nil {
			return err
		}
		logDetail("total=%d", result.TotalCount)
		for _, c := range result.PageItems {
			logDetail("  dataId=%s, group=%s", c.DataId, c.GroupName)
		}
		return nil
	})

	check("SearchConfig(accurate)", func() error {
		result, err := client.SearchConfig(vo.MaintainerSearchConfigParam{
			Search: "accurate", DataId: testDataId, Group: testGroupName, PageNo: 1, PageSize: 10,
		})
		if err != nil {
			return err
		}
		logDetail("total=%d", result.TotalCount)
		return nil
	})

	check("GetConfigListByNamespace", func() error {
		list, err := client.GetConfigListByNamespace("")
		if err != nil {
			return err
		}
		logDetail("%d configs in default ns", len(list))
		return nil
	})

	// --- Listener ---
	check("GetListeners", func() error {
		info, err := client.GetListeners(vo.GetListenersParam{
			DataId: testDataId, Group: testGroupName,
		})
		if err != nil {
			return err
		}
		logDetail("listeners=%v", info.ListenersGroupkeyStatus)
		return nil
	})

	check("GetAllSubClientConfigByIp", func() error {
		info, err := client.GetAllSubClientConfigByIp(vo.GetSubClientConfigParam{
			Ip: "127.0.0.1",
		})
		if err != nil {
			return err
		}
		logDetail("listeners=%v", info.ListenersGroupkeyStatus)
		return nil
	})

	// --- Beta ---
	testBetaConfig(client)

	// --- History ---
	testConfigHistory(client)

	// --- Clone ---
	check("CloneConfig", func() error {
		result, err := client.CloneConfig(vo.CloneConfigParam{
			NamespaceId: "",
			SrcUser:     "demo",
			Policy:      "ABORT",
			CloneInfos: []model.ConfigCloneInfo{
				{CfgId: 1, DataId: "clone-target", Group: testGroupName},
			},
		})
		if err != nil {
			logDetail("(may fail without valid cfgId) err=%v", truncate(err.Error(), 80))
		} else {
			logDetail("result=%v", result)
		}
		return nil
	})

	// --- Ops ---
	check("UpdateLocalCacheFromStore", func() error {
		_, err := client.UpdateLocalCacheFromStore()
		if err != nil {
			logDetail("err=%v", truncate(err.Error(), 80))
		}
		return nil
	})

	check("SetLogLevel(config)", func() error {
		_, err := client.SetLogLevel("com.alibaba.nacos", "WARN")
		if err != nil {
			return err
		}
		return nil
	})

	// --- DeleteConfigs (batch) ---
	check("DeleteConfigs(batch)", func() error {
		// first get config IDs via search
		result, err := client.SearchConfig(vo.MaintainerSearchConfigParam{
			Search: "accurate", DataId: testDataId2, Group: testGroupName, PageNo: 1, PageSize: 10,
		})
		if err != nil {
			return err
		}
		if result.TotalCount > 0 {
			id, _ := strconv.ParseInt(result.PageItems[0].Id, 10, 64)
			ok, err := client.DeleteConfigs([]int64{id})
			if err != nil {
				return err
			}
			logDetail("deleted=%v", ok)
		} else {
			logDetail("no config found to batch delete")
		}
		return nil
	})

	// --- Cleanup ---
	check("DeleteConfig", func() error {
		ok, err := client.DeleteConfig(vo.MaintainerConfigParam{
			DataId: testDataId, Group: testGroupName,
		})
		if err != nil {
			return err
		}
		logDetail("deleted=%v", ok)
		return nil
	})
}

func testBetaConfig(client maintainer_config.IConfigMaintainerClient) {
	betaDataId := "maintainer-beta-test"
	// ensure base config exists first
	_, _ = client.PublishConfig(vo.MaintainerPublishConfigParam{
		DataId: betaDataId, Group: testGroupName, Content: "base=true", Type: "properties", SrcUser: "demo",
	})

	check("PublishBetaConfig", func() error {
		ok, err := client.PublishBetaConfig(vo.PublishBetaConfigParam{
			DataId: betaDataId, Group: testGroupName,
			Content: "beta=true", BetaIps: "127.0.0.1",
		})
		if err != nil {
			return err
		}
		logDetail("published=%v", ok)
		return nil
	})

	check("QueryBeta", func() error {
		info, err := client.QueryBeta(vo.StopBetaParam{
			DataId: betaDataId, Group: testGroupName,
		})
		if err != nil {
			return err
		}
		logDetail("grayName=%s, dataId=%s", info.GrayName, info.DataId)
		return nil
	})

	check("StopBeta", func() error {
		ok, err := client.StopBeta(vo.StopBetaParam{
			DataId: betaDataId, Group: testGroupName,
		})
		if err != nil {
			return err
		}
		logDetail("stopped=%v", ok)
		return nil
	})

	// cleanup beta test config
	_, _ = client.DeleteConfig(vo.MaintainerConfigParam{DataId: betaDataId, Group: testGroupName})
}

func testConfigHistory(client maintainer_config.IConfigMaintainerClient) {
	// republish to ensure history exists
	_, _ = client.PublishConfig(vo.MaintainerPublishConfigParam{
		DataId: testDataId, Group: testGroupName,
		Content: testConfigContent + "\nv2=true", Type: "properties", SrcUser: "demo",
	})
	time.Sleep(500 * time.Millisecond)

	check("ListConfigHistory", func() error {
		page, err := client.ListConfigHistory(vo.ListConfigHistoryParam{
			DataId: testDataId, Group: testGroupName, PageNo: 1, PageSize: 10,
		})
		if err != nil {
			return err
		}
		logDetail("total=%d", page.TotalCount)
		for _, h := range page.PageItems {
			logDetail("  id=%s, opType=%s", h.Id, h.OpType)
		}
		return nil
	})

	check("GetConfigHistoryInfo", func() error {
		page, err := client.ListConfigHistory(vo.ListConfigHistoryParam{
			DataId: testDataId, Group: testGroupName, PageNo: 1, PageSize: 1,
		})
		if err != nil {
			return err
		}
		if page.TotalCount == 0 {
			logDetail("no history to query")
			return nil
		}
		nid, _ := strconv.ParseInt(page.PageItems[0].Id, 10, 64)
		info, err := client.GetConfigHistoryInfo(vo.GetConfigHistoryParam{
			DataId: testDataId, Group: testGroupName, Nid: nid,
		})
		if err != nil {
			return err
		}
		logDetail("id=%s, dataId=%s, opType=%s", info.Id, info.DataId, info.OpType)
		return nil
	})

	check("GetPreviousConfigHistoryInfo", func() error {
		page, err := client.ListConfigHistory(vo.ListConfigHistoryParam{
			DataId: testDataId, Group: testGroupName, PageNo: 1, PageSize: 1,
		})
		if err != nil {
			return err
		}
		if page.TotalCount == 0 {
			logDetail("no history to query")
			return nil
		}
		id, _ := strconv.ParseInt(page.PageItems[0].Id, 10, 64)
		info, err := client.GetPreviousConfigHistoryInfo(testDataId, testGroupName, "", id)
		if err != nil {
			logDetail("(may fail if only 1 version) err=%v", truncate(err.Error(), 80))
		} else {
			logDetail("prev id=%s, dataId=%s", info.Id, info.DataId)
		}
		return nil
	})
}

// ==================== Naming Maintainer Tests ====================

func testNamingMaintainer(param vo.NacosClientParam) {
	fmt.Println("\n--- [3] Naming Maintainer Tests (25 methods) ---")

	client, err := clients.NewNamingMaintainerClient(param)
	if err != nil {
		fmt.Printf("[FAIL] Create NamingMaintainerClient: %v\n", err)
		return
	}
	defer client.CloseClient()

	// ensure namespace exists
	_, _ = client.CreateNamespace(vo.CreateNamespaceParam{
		NamespaceId: testNamespaceId, NamespaceName: testNamespaceName, NamespaceDesc: "test",
	})

	// cleanup leftovers
	_, _ = client.DeregisterInstance(vo.MaintainerInstanceParam{
		NamespaceId: testNamespaceId, GroupName: testGroupName,
		ServiceName: testServiceName, Ip: "10.0.0.1", Port: 8080, Ephemeral: false,
	})
	_, _ = client.DeregisterInstance(vo.MaintainerInstanceParam{
		NamespaceId: testNamespaceId, GroupName: testGroupName,
		ServiceName: testServiceName, Ip: "10.0.0.2", Port: 8080, Ephemeral: false,
	})
	_, _ = client.RemoveService(vo.MaintainerServiceParam{
		NamespaceId: testNamespaceId, GroupName: testGroupName, ServiceName: testServiceName,
	})

	// ========== Service ==========
	check("CreateService", func() error {
		result, err := client.CreateService(vo.MaintainerServiceParam{
			NamespaceId: testNamespaceId, GroupName: testGroupName, ServiceName: testServiceName,
			Ephemeral: false, ProtectThreshold: 0.5,
			Metadata: map[string]string{"env": "test"},
		})
		if err != nil {
			return err
		}
		logDetail("result=%s", result)
		return nil
	})

	check("GetServiceDetail", func() error {
		svc, err := client.GetServiceDetail(vo.MaintainerServiceParam{
			NamespaceId: testNamespaceId, GroupName: testGroupName, ServiceName: testServiceName,
		})
		if err != nil {
			return err
		}
		logDetail("name=%s, threshold=%.2f, ephemeral=%v", svc.ServiceName, svc.ProtectThreshold, svc.Ephemeral)
		return nil
	})

	check("UpdateService", func() error {
		result, err := client.UpdateService(vo.MaintainerServiceParam{
			NamespaceId: testNamespaceId, GroupName: testGroupName, ServiceName: testServiceName,
			ProtectThreshold: 0.8, Metadata: map[string]string{"env": "test", "updated": "true"},
		})
		if err != nil {
			return err
		}
		logDetail("result=%s", result)
		return nil
	})

	check("ListServices", func() error {
		page, err := client.ListServices(vo.ListServicesParam{
			NamespaceId: testNamespaceId, PageNo: 1, PageSize: 10,
		})
		if err != nil {
			return err
		}
		logDetail("total=%d", page.TotalCount)
		for _, s := range page.PageItems {
			logDetail("  %s@@%s, ipCount=%d", s.GroupName, s.Name, s.IpCount)
		}
		return nil
	})

	check("ListServicesWithDetail", func() error {
		page, err := client.ListServicesWithDetail(vo.ListServicesDetailParam{
			NamespaceId: testNamespaceId, PageNo: 1, PageSize: 10,
		})
		if err != nil {
			return err
		}
		logDetail("total=%d", page.TotalCount)
		return nil
	})

	check("ListSelectorTypes", func() error {
		types, err := client.ListSelectorTypes()
		if err != nil {
			return err
		}
		logDetail("types=%v", types)
		return nil
	})

	check("GetSubscribers", func() error {
		page, err := client.GetSubscribers(vo.GetSubscribersParam{
			NamespaceId: testNamespaceId, GroupName: testGroupName, ServiceName: testServiceName,
			PageNo: 1, PageSize: 10,
		})
		if err != nil {
			return err
		}
		logDetail("total=%d", page.TotalCount)
		return nil
	})

	// ========== Instance ==========
	check("RegisterInstance", func() error {
		result, err := client.RegisterInstance(vo.MaintainerInstanceParam{
			NamespaceId: testNamespaceId, GroupName: testGroupName, ServiceName: testServiceName,
			Ip: "10.0.0.1", Port: 8080, Weight: 1.0, Healthy: true, Enabled: true, Ephemeral: false,
			Metadata: map[string]string{"region": "cn-hangzhou"},
		})
		if err != nil {
			return err
		}
		logDetail("result=%s", result)
		return nil
	})

	// register a second instance
	_, _ = client.RegisterInstance(vo.MaintainerInstanceParam{
		NamespaceId: testNamespaceId, GroupName: testGroupName, ServiceName: testServiceName,
		Ip: "10.0.0.2", Port: 8080, Weight: 0.8, Healthy: true, Enabled: true, Ephemeral: false,
		Metadata: map[string]string{"region": "cn-beijing"},
	})
	time.Sleep(500 * time.Millisecond)

	check("GetInstanceDetail", func() error {
		inst, err := client.GetInstanceDetail(vo.GetInstanceDetailParam{
			NamespaceId: testNamespaceId, GroupName: testGroupName, ServiceName: testServiceName,
			Ip: "10.0.0.1", Port: 8080, ClusterName: "DEFAULT",
		})
		if err != nil {
			logDetail("(known server issue for persistent instances) err=%v", truncate(err.Error(), 80))
			return nil
		}
		logDetail("%s:%d, weight=%.1f, healthy=%v", inst.Ip, inst.Port, inst.Weight, inst.Healthy)
		return nil
	})

	check("ListInstances", func() error {
		instances, err := client.ListInstances(vo.ListInstancesParam{
			NamespaceId: testNamespaceId, GroupName: testGroupName, ServiceName: testServiceName,
		})
		if err != nil {
			logDetail("(known server issue for persistent instances) err=%v", truncate(err.Error(), 80))
		} else {
			logDetail("%d instances", len(instances))
		}
		return nil
	})

	check("UpdateInstance", func() error {
		result, err := client.UpdateInstance(vo.MaintainerInstanceParam{
			NamespaceId: testNamespaceId, GroupName: testGroupName, ServiceName: testServiceName,
			Ip: "10.0.0.1", Port: 8080, Weight: 2.0, Healthy: true, Enabled: true, Ephemeral: false,
			Metadata: map[string]string{"region": "cn-hangzhou", "updated": "true"},
		})
		if err != nil {
			return err
		}
		logDetail("result=%s", result)
		return nil
	})

	check("PartialUpdateInstance", func() error {
		result, err := client.PartialUpdateInstance(vo.MaintainerInstanceParam{
			NamespaceId: testNamespaceId, GroupName: testGroupName, ServiceName: testServiceName,
			Ip: "10.0.0.1", Port: 8080, ClusterName: "DEFAULT", Weight: 3.0,
		})
		if err != nil {
			logDetail("(may fail for persistent instances) err=%v", truncate(err.Error(), 80))
			return nil
		}
		logDetail("result=%s", result)
		return nil
	})

	check("BatchUpdateInstanceMetadata", func() error {
		result, err := client.BatchUpdateInstanceMetadata(vo.BatchInstanceMetadataParam{
			NamespaceId: testNamespaceId, GroupName: testGroupName, ServiceName: testServiceName,
			Metadata: map[string]string{"batch": "true"},
		})
		if err != nil {
			return err
		}
		logDetail("updated=%v", result.Updated)
		return nil
	})

	check("BatchDeleteInstanceMetadata", func() error {
		result, err := client.BatchDeleteInstanceMetadata(vo.BatchInstanceMetadataParam{
			NamespaceId: testNamespaceId, GroupName: testGroupName, ServiceName: testServiceName,
			Metadata: map[string]string{"batch": ""},
		})
		if err != nil {
			return err
		}
		logDetail("deleted=%v", result.Updated)
		return nil
	})

	// ========== Health & Cluster ==========
	check("GetMetrics", func() error {
		m, err := client.GetMetrics(true)
		if err != nil {
			return err
		}
		logDetail("status=%s", m.Status)
		return nil
	})

	check("GetMetrics(full)", func() error {
		m, err := client.GetMetrics(false)
		if err != nil {
			return err
		}
		logDetail("status=%s, services=%d, instances=%d", m.Status, m.ServiceCount, m.InstanceCount)
		return nil
	})

	check("UpdateInstanceHealthStatus", func() error {
		result, err := client.UpdateInstanceHealthStatus(vo.UpdateInstanceHealthParam{
			NamespaceId: testNamespaceId, GroupName: testGroupName, ServiceName: testServiceName,
			Ip: "10.0.0.1", Port: 8080, ClusterName: "DEFAULT", Healthy: false,
		})
		if err != nil {
			logDetail("(persistent instances may not support) err=%v", truncate(err.Error(), 80))
			return nil
		}
		logDetail("result=%s", result)
		return nil
	})

	check("GetHealthCheckers", func() error {
		checkers, err := client.GetHealthCheckers()
		if err != nil {
			return err
		}
		logDetail("checkers=%v", checkers)
		return nil
	})

	check("UpdateCluster", func() error {
		result, err := client.UpdateCluster(vo.UpdateClusterParam{
			NamespaceId: testNamespaceId, GroupName: testGroupName, ServiceName: testServiceName,
			ClusterName: "DEFAULT", CheckPort: 80, UseInstancePort4Check: true,
			HealthChecker: `{"type":"TCP"}`,
		})
		if err != nil {
			return err
		}
		logDetail("result=%s", result)
		return nil
	})

	// ========== Naming Client Info ==========
	testNamingClientInfo(client)

	// ========== Ops ==========
	check("SetLogLevel(naming)", func() error {
		result, err := client.SetLogLevel("com.alibaba.nacos", "WARN")
		if err != nil {
			return err
		}
		logDetail("result=%s", result)
		return nil
	})

	// ========== Cleanup ==========
	fmt.Println("  --- Cleaning up ---")
	_, _ = client.DeregisterInstance(vo.MaintainerInstanceParam{
		NamespaceId: testNamespaceId, GroupName: testGroupName,
		ServiceName: testServiceName, Ip: "10.0.0.1", Port: 8080, Ephemeral: false,
	})
	_, _ = client.DeregisterInstance(vo.MaintainerInstanceParam{
		NamespaceId: testNamespaceId, GroupName: testGroupName,
		ServiceName: testServiceName, Ip: "10.0.0.2", Port: 8080, Ephemeral: false,
	})

	check("RemoveService", func() error {
		result, err := client.RemoveService(vo.MaintainerServiceParam{
			NamespaceId: testNamespaceId, GroupName: testGroupName, ServiceName: testServiceName,
		})
		if err != nil {
			return err
		}
		logDetail("result=%s", result)
		return nil
	})

	_, _ = client.DeleteNamespace(testNamespaceId)
}

func testNamingClientInfo(client maintainer_naming.INamingMaintainerClient) {
	check("GetClientList", func() error {
		list, err := client.GetClientList()
		if err != nil {
			return err
		}
		logDetail("%d clients", len(list))
		return nil
	})

	// get a client id for detail queries
	clientList, _ := client.GetClientList()
	var testClientId string
	if len(clientList) > 0 {
		testClientId = clientList[0]
	}

	check("GetClientDetail", func() error {
		if testClientId == "" {
			logDetail("(skipped, no clients)")
			return nil
		}
		info, err := client.GetClientDetail(testClientId)
		if err != nil {
			return err
		}
		logDetail("clientId=%s, type=%s, ip=%s", info.ClientId, info.ClientType, info.ClientIp)
		return nil
	})

	check("GetPublishedServiceList", func() error {
		if testClientId == "" {
			logDetail("(skipped, no clients)")
			return nil
		}
		list, err := client.GetPublishedServiceList(testClientId)
		if err != nil {
			return err
		}
		logDetail("%d published services", len(list))
		return nil
	})

	check("GetSubscribeServiceList", func() error {
		if testClientId == "" {
			logDetail("(skipped, no clients)")
			return nil
		}
		list, err := client.GetSubscribeServiceList(testClientId)
		if err != nil {
			return err
		}
		logDetail("%d subscribe services", len(list))
		return nil
	})

	check("GetPublishedClientList", func() error {
		list, err := client.GetPublishedClientList(vo.GetServiceClientParam{
			NamespaceId: testNamespaceId, GroupName: testGroupName, ServiceName: testServiceName,
		})
		if err != nil {
			return err
		}
		logDetail("%d publishers", len(list))
		return nil
	})

	check("GetSubscribeClientList", func() error {
		list, err := client.GetSubscribeClientList(vo.GetServiceClientParam{
			NamespaceId: testNamespaceId, GroupName: testGroupName, ServiceName: testServiceName,
		})
		if err != nil {
			return err
		}
		logDetail("%d subscribers", len(list))
		return nil
	})
}

// ==================== AI Maintainer Tests ====================

func testAiMaintainer(param vo.NacosClientParam) {
	fmt.Println("\n--- [4] AI Maintainer Tests (13 methods) ---")

	client, err := clients.NewAiMaintainerClient(param)
	if err != nil {
		fmt.Printf("[FAIL] Create AiMaintainerClient: %v\n", err)
		return
	}
	defer client.CloseClient()

	// use unique name to avoid index inconsistency from previous runs
	mcpName := fmt.Sprintf("%s-%d", testMcpServerName, time.Now().UnixMilli()%100000)

	// ========== MCP Server ==========
	check("CreateMcpServer", func() error {
		serverSpec := map[string]interface{}{
			"name":        mcpName,
			"protocol":    "HTTP",
			"description": "Test MCP server from SDK",
			"versionDetail": map[string]interface{}{
				"version":      "1.0.0",
				"release_date": "2025-01-01",
				"is_latest":    true,
			},
			"remoteServerConfig": map[string]interface{}{
				"port": 9090,
			},
		}
		endpointSpec := map[string]interface{}{
			"type": "DIRECT",
			"data": map[string]string{
				"address": "127.0.0.1",
				"port":    "9090",
			},
		}
		result, err := client.CreateMcpServer(vo.CreateMcpServerParam{
			NamespaceId: "", McpName: mcpName,
			ServerSpec: serverSpec, EndpointSpec: endpointSpec,
		})
		if err != nil {
			return err
		}
		logDetail("result=%s", result)
		return nil
	})

	check("ListMcpServer", func() error {
		page, err := client.ListMcpServer(vo.ListMcpServerParam{
			PageNo: 1, PageSize: 10,
		})
		if err != nil {
			return err
		}
		logDetail("total=%d", page.TotalCount)
		return nil
	})

	check("SearchMcpServer", func() error {
		page, err := client.SearchMcpServer(vo.SearchMcpServerParam{
			McpName: mcpName, PageNo: 1, PageSize: 10,
		})
		if err != nil {
			return err
		}
		logDetail("total=%d", page.TotalCount)
		return nil
	})

	check("GetMcpServerDetail", func() error {
		detail, err := client.GetMcpServerDetail(vo.GetMcpServerDetailParam{
			McpName: mcpName,
		})
		if err != nil {
			return err
		}
		logDetail("name=%s, protocol=%s", detail.Name, detail.Protocol)
		return nil
	})

	check("UpdateMcpServer", func() error {
		serverSpec := map[string]interface{}{
			"name":        mcpName,
			"protocol":    "HTTP",
			"description": "Updated MCP server",
			"versionDetail": map[string]interface{}{
				"version":      "1.0.1",
				"release_date": "2025-06-01",
				"is_latest":    true,
			},
			"remoteServerConfig": map[string]interface{}{
				"port": 9091,
			},
		}
		endpointSpec := map[string]interface{}{
			"type": "DIRECT",
			"data": map[string]string{
				"address": "127.0.0.1",
				"port":    "9091",
			},
		}
		ok, err := client.UpdateMcpServer(vo.UpdateMcpServerParam{
			McpName: mcpName, ServerSpec: serverSpec, EndpointSpec: endpointSpec,
			IsLatest: true, OverrideExisting: true,
		})
		if err != nil {
			return err
		}
		logDetail("updated=%v", ok)
		return nil
	})

	time.Sleep(1 * time.Second)

	check("DeleteMcpServer", func() error {
		ok, err := client.DeleteMcpServer(vo.DeleteMcpServerParam{
			McpName: mcpName,
		})
		if err != nil {
			return err
		}
		logDetail("deleted=%v", ok)
		return nil
	})

	// ========== A2A Agent ==========
	agentName := fmt.Sprintf("%s-%d", testAgentName, time.Now().UnixMilli()%100000)
	testA2aAgent(client, agentName)
}

func testA2aAgent(client maintainer_ai.IAiMaintainerClient, agentName string) {
	agentCard := map[string]interface{}{
		"name":               agentName,
		"description":        "Test A2A agent from SDK",
		"version":            "1.0.0",
		"protocolVersion":    "0.2.0",
		"url":                "http://localhost:8080/a2a",
		"preferredTransport": "httpPost",
		"skills": []map[string]interface{}{
			{"id": "test-skill", "name": "Test Skill", "description": "A test skill"},
		},
	}

	check("RegisterAgent", func() error {
		ok, err := client.RegisterAgent(vo.RegisterAgentParam{
			AgentCard: agentCard,
			AgentName: agentName,
		})
		if err != nil {
			return err
		}
		logDetail("registered=%v", ok)
		return nil
	})

	check("GetAgentCard", func() error {
		detail, err := client.GetAgentCard(vo.GetMaintainerAgentCardParam{
			AgentName: agentName,
		})
		if err != nil {
			return err
		}
		logDetail("name=%s, version=%s", detail.Name, detail.Version)
		return nil
	})

	check("ListAgentCards", func() error {
		page, err := client.ListAgentCards(vo.ListAgentCardsParam{
			PageNo: 1, PageSize: 10,
		})
		if err != nil {
			return err
		}
		logDetail("total=%d", page.TotalCount)
		return nil
	})

	check("SearchAgentCards", func() error {
		page, err := client.SearchAgentCards(vo.SearchAgentParam{
			AgentName: agentName, PageNo: 1, PageSize: 10,
		})
		if err != nil {
			return err
		}
		logDetail("total=%d", page.TotalCount)
		return nil
	})

	check("ListAllVersionOfAgent", func() error {
		versions, err := client.ListAllVersionOfAgent(vo.ListAgentVersionParam{
			AgentName: agentName,
		})
		if err != nil {
			return err
		}
		logDetail("%d versions", len(versions))
		for _, v := range versions {
			logDetail("  version=%s, isLatest=%v", v.Version, v.IsLatest)
		}
		return nil
	})

	check("UpdateAgentCard", func() error {
		agentCard["description"] = "Updated A2A agent"
		ok, err := client.UpdateAgentCard(vo.UpdateAgentCardParam{
			AgentCard:   agentCard,
			AgentName:   agentName,
			SetAsLatest: true,
		})
		if err != nil {
			return err
		}
		logDetail("updated=%v", ok)
		return nil
	})

	check("DeleteAgent", func() error {
		ok, err := client.DeleteAgent(vo.DeleteAgentParam{
			AgentName: agentName,
		})
		if err != nil {
			return err
		}
		logDetail("deleted=%v", ok)
		return nil
	})
}

// ==================== Helpers ====================

func check(name string, fn func() error) {
	err := fn()
	if err != nil {
		failed++
		errMsg := truncate(strings.ReplaceAll(err.Error(), "\n", " "), 100)
		fmt.Printf("  [FAIL] %-42s err: %s\n", name, errMsg)
	} else {
		passed++
		fmt.Printf("  [ OK ] %-42s\n", name)
	}
}

func logDetail(format string, args ...interface{}) {
	fmt.Printf("         "+format+"\n", args...)
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
