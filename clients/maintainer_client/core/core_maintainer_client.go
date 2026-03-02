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

package core

import (
	"context"
	"net/http"
	"strconv"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/nacos_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/maintainer_client/remote"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/common/security"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// CoreMaintainerClient implements ICoreMaintainerClient.
// It is designed to be embedded into NamingMaintainerClient, ConfigMaintainerClient, etc.
type CoreMaintainerClient struct {
	ctx    context.Context
	cancel context.CancelFunc
	nacos_client.INacosClient
	Proxy *remote.MaintainerHttpProxy
}

// NewCoreMaintainerClient creates a CoreMaintainerClient with its own MaintainerHttpProxy.
func NewCoreMaintainerClient(nc nacos_client.INacosClient, provider security.RamCredentialProvider) (*CoreMaintainerClient, error) {
	clientCfg, err := nc.GetClientConfig()
	if err != nil {
		return nil, err
	}
	serverCfgs, err := nc.GetServerConfig()
	if err != nil {
		return nil, err
	}
	httpAgent, err := nc.GetHttpAgent()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	proxy, err := remote.NewMaintainerHttpProxy(ctx, serverCfgs, clientCfg, httpAgent, provider)
	if err != nil {
		cancel()
		return nil, err
	}

	return &CoreMaintainerClient{
		ctx:          ctx,
		cancel:       cancel,
		INacosClient: nc,
		Proxy:        proxy,
	}, nil
}

func (c *CoreMaintainerClient) GetServerState() (map[string]string, error) {
	return remote.DoGet[map[string]string](c.Proxy, constant.AdminCoreStatePath, nil, remote.BuildEmptyResource())
}

func (c *CoreMaintainerClient) Liveness() (bool, error) {
	resp, err := c.Proxy.ReqApi(constant.AdminCoreStatePath+"/liveness", nil, http.MethodGet, remote.BuildEmptyResource())
	if err != nil {
		return false, err
	}
	_ = resp
	return true, nil
}

func (c *CoreMaintainerClient) Readiness() (bool, error) {
	resp, err := c.Proxy.ReqApi(constant.AdminCoreStatePath+"/readiness", nil, http.MethodGet, remote.BuildEmptyResource())
	if err != nil {
		return false, err
	}
	_ = resp
	return true, nil
}

func (c *CoreMaintainerClient) RaftOps(command, value, groupId string) (string, error) {
	body := map[string]string{
		"command": command,
		"value":   value,
		"groupId": groupId,
	}
	return remote.DoPostJson[string](c.Proxy, constant.AdminCoreOpsPath+"/raft", nil, body, remote.BuildEmptyResource())
}

func (c *CoreMaintainerClient) GetIdGenerators() ([]model.IdGeneratorInfo, error) {
	return remote.DoGet[[]model.IdGeneratorInfo](c.Proxy, constant.AdminCoreOpsPath+"/ids", nil, remote.BuildEmptyResource())
}

func (c *CoreMaintainerClient) UpdateLogLevel(logName, logLevel string) error {
	body := map[string]string{
		"logName":  logName,
		"logLevel": logLevel,
	}
	_, err := c.Proxy.ReqApiWithJsonBody(constant.AdminCoreOpsPath+"/log", nil, body, http.MethodPut, remote.BuildEmptyResource())
	return err
}

func (c *CoreMaintainerClient) ListClusterNodes(address, state string) ([]model.NacosMember, error) {
	params := map[string]string{
		"address": address,
		"state":   state,
	}
	return remote.DoGet[[]model.NacosMember](c.Proxy, constant.AdminCoreClusterPath+"/node/list", params, remote.BuildEmptyResource())
}

func (c *CoreMaintainerClient) UpdateLookupMode(lookupType string) (bool, error) {
	params := map[string]string{
		"type": lookupType,
	}
	return remote.DoPut[bool](c.Proxy, constant.AdminCoreClusterPath+"/lookup", params, remote.BuildEmptyResource())
}

func (c *CoreMaintainerClient) GetCurrentClients() (map[string]model.ConnectionInfo, error) {
	return remote.DoGet[map[string]model.ConnectionInfo](c.Proxy, constant.AdminCoreLoaderPath+"/current", nil, remote.BuildEmptyResource())
}

func (c *CoreMaintainerClient) ReloadConnectionCount(count int, redirectAddress string) (string, error) {
	params := map[string]string{
		"count":           strconv.Itoa(count),
		"redirectAddress": redirectAddress,
	}
	return remote.DoPost[string](c.Proxy, constant.AdminCoreLoaderPath+"/reloadCurrent", params, remote.BuildEmptyResource())
}

func (c *CoreMaintainerClient) SmartReloadCluster(loaderFactor string) (string, error) {
	params := map[string]string{
		"loaderFactorStr": loaderFactor,
	}
	return remote.DoPost[string](c.Proxy, constant.AdminCoreLoaderPath+"/smartReloadCluster", params, remote.BuildEmptyResource())
}

func (c *CoreMaintainerClient) ReloadSingleClient(connectionId, redirectAddress string) (string, error) {
	params := map[string]string{
		"connectionId":    connectionId,
		"redirectAddress": redirectAddress,
	}
	return remote.DoPost[string](c.Proxy, constant.AdminCoreLoaderPath+"/reloadClient", params, remote.BuildEmptyResource())
}

func (c *CoreMaintainerClient) GetClusterLoaderMetrics() (model.ServerLoaderMetrics, error) {
	return remote.DoGet[model.ServerLoaderMetrics](c.Proxy, constant.AdminCoreLoaderPath+"/cluster", nil, remote.BuildEmptyResource())
}

// --- Namespace ---

func (c *CoreMaintainerClient) GetNamespaceList() ([]model.Namespace, error) {
	return remote.DoGet[[]model.Namespace](c.Proxy, constant.AdminCoreNamespacePath+"/list", nil, remote.BuildEmptyResource())
}

func (c *CoreMaintainerClient) GetNamespace(namespaceId string) (model.Namespace, error) {
	params := map[string]string{
		"namespaceId": namespaceId,
	}
	return remote.DoGet[model.Namespace](c.Proxy, constant.AdminCoreNamespacePath, params, remote.BuildEmptyResource())
}

func (c *CoreMaintainerClient) CreateNamespace(param vo.CreateNamespaceParam) (bool, error) {
	params := remote.TransformParam(param)
	return remote.DoPost[bool](c.Proxy, constant.AdminCoreNamespacePath, params, remote.BuildEmptyResource())
}

func (c *CoreMaintainerClient) UpdateNamespace(param vo.UpdateNamespaceParam) (bool, error) {
	params := remote.TransformParam(param)
	return remote.DoPut[bool](c.Proxy, constant.AdminCoreNamespacePath, params, remote.BuildEmptyResource())
}

func (c *CoreMaintainerClient) DeleteNamespace(namespaceId string) (bool, error) {
	params := map[string]string{
		"namespaceId": namespaceId,
	}
	return remote.DoDelete[bool](c.Proxy, constant.AdminCoreNamespacePath, params, remote.BuildEmptyResource())
}

func (c *CoreMaintainerClient) CheckNamespaceIdExist(namespaceId string) (bool, error) {
	params := map[string]string{
		"namespaceId": namespaceId,
	}
	resp, err := c.Proxy.ReqApi(constant.AdminCoreNamespacePath+"/check", params, http.MethodGet, remote.BuildEmptyResource())
	if err != nil {
		return false, err
	}
	count, err := remote.ParseResult[int](resp)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// --- Lifecycle ---

func (c *CoreMaintainerClient) CloseClient() {
	c.cancel()
}
