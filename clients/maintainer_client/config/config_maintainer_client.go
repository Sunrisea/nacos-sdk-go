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

package config

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/maintainer_client/core"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/maintainer_client/remote"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/nacos_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/common/security"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// ConfigMaintainerClient implements IConfigMaintainerClient.
type ConfigMaintainerClient struct {
	*core.CoreMaintainerClient
}

func NewConfigMaintainerClient(nc nacos_client.INacosClient, provider security.RamCredentialProvider) (*ConfigMaintainerClient, error) {
	coreClient, err := core.NewCoreMaintainerClient(nc, provider)
	if err != nil {
		return nil, err
	}
	return &ConfigMaintainerClient{CoreMaintainerClient: coreClient}, nil
}

func (c *ConfigMaintainerClient) buildConfigResource(namespace, group, dataId string) security.RequestResource {
	return remote.BuildResource(namespace, group, dataId)
}

// --- Config CRUD ---

func (c *ConfigMaintainerClient) GetConfig(param vo.MaintainerConfigParam) (model.ConfigDetailInfo, error) {
	params := remote.TransformParam(param)
	res := c.buildConfigResource(param.NamespaceId, param.Group, param.DataId)
	return remote.DoGet[model.ConfigDetailInfo](c.Proxy, constant.AdminConfigPath, params, res)
}

func (c *ConfigMaintainerClient) PublishConfig(param vo.MaintainerPublishConfigParam) (bool, error) {
	params := remote.TransformParam(param)
	res := c.buildConfigResource(param.NamespaceId, param.Group, param.DataId)
	return remote.DoPost[bool](c.Proxy, constant.AdminConfigPath, params, res)
}

func (c *ConfigMaintainerClient) UpdateConfigMetadata(param vo.UpdateConfigMetadataParam) (bool, error) {
	params := remote.TransformParam(param)
	res := c.buildConfigResource(param.NamespaceId, param.Group, param.DataId)
	return remote.DoPut[bool](c.Proxy, constant.AdminConfigPath+"/metadata", params, res)
}

func (c *ConfigMaintainerClient) DeleteConfig(param vo.MaintainerConfigParam) (bool, error) {
	params := remote.TransformParam(param)
	res := c.buildConfigResource(param.NamespaceId, param.Group, param.DataId)
	return remote.DoDelete[bool](c.Proxy, constant.AdminConfigPath, params, res)
}

func (c *ConfigMaintainerClient) DeleteConfigs(ids []int64) (bool, error) {
	idStrs := make([]string, 0, len(ids))
	for _, id := range ids {
		idStrs = append(idStrs, strconv.FormatInt(id, 10))
	}
	params := map[string]string{
		"ids": strings.Join(idStrs, ","),
	}
	return remote.DoDelete[bool](c.Proxy, constant.AdminConfigPath+"/batch", params, remote.BuildEmptyResource())
}

// --- Search ---

func (c *ConfigMaintainerClient) SearchConfig(param vo.MaintainerSearchConfigParam) (model.Page[model.ConfigBasicInfo], error) {
	params := remote.TransformParam(param)
	res := c.buildConfigResource(param.NamespaceId, param.Group, param.DataId)
	return remote.DoGet[model.Page[model.ConfigBasicInfo]](c.Proxy, constant.AdminConfigPath+"/list", params, res)
}

func (c *ConfigMaintainerClient) GetConfigListByNamespace(namespaceId string) ([]model.ConfigBasicInfo, error) {
	params := map[string]string{
		"namespaceId": namespaceId,
	}
	res := c.buildConfigResource(namespaceId, "", "")
	return remote.DoGet[[]model.ConfigBasicInfo](c.Proxy, constant.AdminConfigHistoryPath+"/configs", params, res)
}

// --- Listener ---

func (c *ConfigMaintainerClient) GetListeners(param vo.GetListenersParam) (model.ConfigListenerInfo, error) {
	params := remote.TransformParam(param)
	res := c.buildConfigResource(param.NamespaceId, param.Group, param.DataId)
	return remote.DoGet[model.ConfigListenerInfo](c.Proxy, constant.AdminConfigPath+"/listener", params, res)
}

func (c *ConfigMaintainerClient) GetAllSubClientConfigByIp(param vo.GetSubClientConfigParam) (model.ConfigListenerInfo, error) {
	params := remote.TransformParam(param)
	res := c.buildConfigResource(param.NamespaceId, "", "")
	return remote.DoGet[model.ConfigListenerInfo](c.Proxy, constant.AdminConfigListenerPath, params, res)
}

// --- Clone ---

func (c *ConfigMaintainerClient) CloneConfig(param vo.CloneConfigParam) (map[string]interface{}, error) {
	queryParams := map[string]string{
		"targetNamespaceId": param.NamespaceId,
		"srcUser":           param.SrcUser,
		"policy":            param.Policy,
	}
	res := c.buildConfigResource(param.NamespaceId, "", "")
	return remote.DoPostJson[map[string]interface{}](c.Proxy, constant.AdminConfigPath+"/clone", queryParams, param.CloneInfos, res)
}

// --- Beta ---

func (c *ConfigMaintainerClient) PublishBetaConfig(param vo.PublishBetaConfigParam) (bool, error) {
	params := remote.TransformParam(param)
	headers := map[string]string{
		"betaIps": param.BetaIps,
	}
	delete(params, "betaIps")
	res := c.buildConfigResource(param.NamespaceId, param.Group, param.DataId)
	resp, err := c.Proxy.ReqApiWithHeaders(constant.AdminConfigPath, params, headers, "POST", res)
	if err != nil {
		return false, err
	}
	return remote.ParseResult[bool](resp)
}

func (c *ConfigMaintainerClient) StopBeta(param vo.StopBetaParam) (bool, error) {
	params := remote.TransformParam(param)
	res := c.buildConfigResource(param.NamespaceId, param.Group, param.DataId)
	return remote.DoDelete[bool](c.Proxy, constant.AdminConfigPath+"/beta", params, res)
}

func (c *ConfigMaintainerClient) QueryBeta(param vo.StopBetaParam) (model.ConfigGrayInfo, error) {
	params := remote.TransformParam(param)
	res := c.buildConfigResource(param.NamespaceId, param.Group, param.DataId)
	return remote.DoGet[model.ConfigGrayInfo](c.Proxy, constant.AdminConfigPath+"/beta", params, res)
}

// --- History ---

func (c *ConfigMaintainerClient) ListConfigHistory(param vo.ListConfigHistoryParam) (model.Page[model.ConfigHistoryInfo], error) {
	params := remote.TransformParam(param)
	res := c.buildConfigResource(param.NamespaceId, param.Group, param.DataId)
	return remote.DoGet[model.Page[model.ConfigHistoryInfo]](c.Proxy, constant.AdminConfigHistoryPath+"/list", params, res)
}

func (c *ConfigMaintainerClient) GetConfigHistoryInfo(param vo.GetConfigHistoryParam) (model.ConfigHistoryInfo, error) {
	params := remote.TransformParam(param)
	res := c.buildConfigResource(param.NamespaceId, param.Group, param.DataId)
	return remote.DoGet[model.ConfigHistoryInfo](c.Proxy, constant.AdminConfigHistoryPath, params, res)
}

func (c *ConfigMaintainerClient) GetPreviousConfigHistoryInfo(dataId, group, namespaceId string, id int64) (model.ConfigHistoryInfo, error) {
	params := map[string]string{
		"dataId":      dataId,
		"groupName":   group,
		"namespaceId": namespaceId,
		"id":          strconv.FormatInt(id, 10),
	}
	res := c.buildConfigResource(namespaceId, group, dataId)
	return remote.DoGet[model.ConfigHistoryInfo](c.Proxy, constant.AdminConfigHistoryPath+"/previous", params, res)
}

// --- Ops ---

func (c *ConfigMaintainerClient) UpdateLocalCacheFromStore() (string, error) {
	return remote.DoPost[string](c.Proxy, constant.AdminConfigOpsPath+"/localCache", nil, remote.BuildEmptyResource())
}

func (c *ConfigMaintainerClient) SetLogLevel(logName, logLevel string) (string, error) {
	params := map[string]string{
		"logName":  logName,
		"logLevel": logLevel,
	}
	resp, err := c.Proxy.ReqApi(constant.AdminConfigOpsPath+"/log", params, http.MethodPut, remote.BuildEmptyResource())
	if err != nil {
		return "", err
	}
	return remote.ParseResult[string](resp)
}
