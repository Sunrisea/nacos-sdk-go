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

package naming

import (
	"net/http"
	"strconv"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/maintainer_client/core"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/maintainer_client/remote"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/nacos_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/common/security"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// NamingMaintainerClient implements INamingMaintainerClient.
type NamingMaintainerClient struct {
	*core.CoreMaintainerClient
}

func NewNamingMaintainerClient(nc nacos_client.INacosClient, provider security.RamCredentialProvider) (*NamingMaintainerClient, error) {
	coreClient, err := core.NewCoreMaintainerClient(nc, provider)
	if err != nil {
		return nil, err
	}
	return &NamingMaintainerClient{CoreMaintainerClient: coreClient}, nil
}

func (c *NamingMaintainerClient) buildNamingResource(namespace, group, service string) security.RequestResource {
	return remote.BuildResource(namespace, group, service)
}

// --- Service ---

func (c *NamingMaintainerClient) CreateService(param vo.MaintainerServiceParam) (string, error) {
	params := remote.TransformParam(param)
	res := c.buildNamingResource(param.NamespaceId, param.GroupName, param.ServiceName)
	return remote.DoPost[string](c.Proxy, constant.AdminNamingServicePath, params, res)
}

func (c *NamingMaintainerClient) UpdateService(param vo.MaintainerServiceParam) (string, error) {
	params := remote.TransformParam(param)
	res := c.buildNamingResource(param.NamespaceId, param.GroupName, param.ServiceName)
	return remote.DoPut[string](c.Proxy, constant.AdminNamingServicePath, params, res)
}

func (c *NamingMaintainerClient) RemoveService(param vo.MaintainerServiceParam) (string, error) {
	params := remote.TransformParam(param)
	res := c.buildNamingResource(param.NamespaceId, param.GroupName, param.ServiceName)
	return remote.DoDelete[string](c.Proxy, constant.AdminNamingServicePath, params, res)
}

func (c *NamingMaintainerClient) GetServiceDetail(param vo.MaintainerServiceParam) (model.ServiceDetailInfo, error) {
	params := remote.TransformParam(param)
	res := c.buildNamingResource(param.NamespaceId, param.GroupName, param.ServiceName)
	return remote.DoGet[model.ServiceDetailInfo](c.Proxy, constant.AdminNamingServicePath, params, res)
}

func (c *NamingMaintainerClient) ListServices(param vo.ListServicesParam) (model.Page[model.ServiceView], error) {
	params := remote.TransformParam(param)
	params["withInstances"] = "false"
	res := c.buildNamingResource(param.NamespaceId, "", "")
	return remote.DoGet[model.Page[model.ServiceView]](c.Proxy, constant.AdminNamingServicePath+"/list", params, res)
}

func (c *NamingMaintainerClient) ListServicesWithDetail(param vo.ListServicesDetailParam) (model.Page[model.ServiceDetailInfo], error) {
	params := remote.TransformParam(param)
	params["withInstances"] = "true"
	res := c.buildNamingResource(param.NamespaceId, "", "")
	return remote.DoGet[model.Page[model.ServiceDetailInfo]](c.Proxy, constant.AdminNamingServicePath+"/list", params, res)
}

func (c *NamingMaintainerClient) GetSubscribers(param vo.GetSubscribersParam) (model.Page[model.SubscriberInfo], error) {
	params := remote.TransformParam(param)
	res := c.buildNamingResource(param.NamespaceId, param.GroupName, param.ServiceName)
	return remote.DoGet[model.Page[model.SubscriberInfo]](c.Proxy, constant.AdminNamingServicePath+"/subscribers", params, res)
}

func (c *NamingMaintainerClient) ListSelectorTypes() ([]string, error) {
	return remote.DoGet[[]string](c.Proxy, constant.AdminNamingServicePath+"/selector/types", nil, remote.BuildEmptyResource())
}

// --- Instance ---

func (c *NamingMaintainerClient) RegisterInstance(param vo.MaintainerInstanceParam) (string, error) {
	params := remote.TransformParam(param)
	res := c.buildNamingResource(param.NamespaceId, param.GroupName, param.ServiceName)
	return remote.DoPost[string](c.Proxy, constant.AdminNamingInstancePath, params, res)
}

func (c *NamingMaintainerClient) DeregisterInstance(param vo.MaintainerInstanceParam) (string, error) {
	params := remote.TransformParam(param)
	res := c.buildNamingResource(param.NamespaceId, param.GroupName, param.ServiceName)
	return remote.DoDelete[string](c.Proxy, constant.AdminNamingInstancePath, params, res)
}

func (c *NamingMaintainerClient) UpdateInstance(param vo.MaintainerInstanceParam) (string, error) {
	params := remote.TransformParam(param)
	res := c.buildNamingResource(param.NamespaceId, param.GroupName, param.ServiceName)
	return remote.DoPut[string](c.Proxy, constant.AdminNamingInstancePath, params, res)
}

func (c *NamingMaintainerClient) BatchUpdateInstanceMetadata(param vo.BatchInstanceMetadataParam) (model.InstanceMetadataBatchResult, error) {
	params := remote.TransformParam(param)
	res := c.buildNamingResource(param.NamespaceId, param.GroupName, param.ServiceName)
	return remote.DoPut[model.InstanceMetadataBatchResult](c.Proxy, constant.AdminNamingInstancePath+"/metadata/batch", params, res)
}

func (c *NamingMaintainerClient) BatchDeleteInstanceMetadata(param vo.BatchInstanceMetadataParam) (model.InstanceMetadataBatchResult, error) {
	params := remote.TransformParam(param)
	res := c.buildNamingResource(param.NamespaceId, param.GroupName, param.ServiceName)
	return remote.DoDelete[model.InstanceMetadataBatchResult](c.Proxy, constant.AdminNamingInstancePath+"/metadata/batch", params, res)
}

func (c *NamingMaintainerClient) PartialUpdateInstance(param vo.MaintainerInstanceParam) (string, error) {
	params := remote.TransformParam(param)
	res := c.buildNamingResource(param.NamespaceId, param.GroupName, param.ServiceName)
	return remote.DoPut[string](c.Proxy, constant.AdminNamingInstancePath+"/partial", params, res)
}

func (c *NamingMaintainerClient) ListInstances(param vo.ListInstancesParam) ([]model.Instance, error) {
	params := remote.TransformParam(param)
	res := c.buildNamingResource(param.NamespaceId, param.GroupName, param.ServiceName)
	return remote.DoGet[[]model.Instance](c.Proxy, constant.AdminNamingInstancePath+"/list", params, res)
}

func (c *NamingMaintainerClient) GetInstanceDetail(param vo.GetInstanceDetailParam) (model.Instance, error) {
	params := remote.TransformParam(param)
	res := c.buildNamingResource(param.NamespaceId, param.GroupName, param.ServiceName)
	return remote.DoGet[model.Instance](c.Proxy, constant.AdminNamingInstancePath, params, res)
}

// --- Naming Client Info ---

func (c *NamingMaintainerClient) GetClientList() ([]string, error) {
	return remote.DoGet[[]string](c.Proxy, constant.AdminNamingClientPath+"/list", nil, remote.BuildEmptyResource())
}

func (c *NamingMaintainerClient) GetClientDetail(clientId string) (model.ClientSummaryInfo, error) {
	params := map[string]string{"clientId": clientId}
	return remote.DoGet[model.ClientSummaryInfo](c.Proxy, constant.AdminNamingClientPath, params, remote.BuildEmptyResource())
}

func (c *NamingMaintainerClient) GetPublishedServiceList(clientId string) ([]model.ClientServiceInfo, error) {
	params := map[string]string{"clientId": clientId}
	return remote.DoGet[[]model.ClientServiceInfo](c.Proxy, constant.AdminNamingClientPath+"/publish/list", params, remote.BuildEmptyResource())
}

func (c *NamingMaintainerClient) GetSubscribeServiceList(clientId string) ([]model.ClientServiceInfo, error) {
	params := map[string]string{"clientId": clientId}
	return remote.DoGet[[]model.ClientServiceInfo](c.Proxy, constant.AdminNamingClientPath+"/subscribe/list", params, remote.BuildEmptyResource())
}

func (c *NamingMaintainerClient) GetPublishedClientList(param vo.GetServiceClientParam) ([]model.ClientPublisherInfo, error) {
	params := remote.TransformParam(param)
	res := c.buildNamingResource(param.NamespaceId, param.GroupName, param.ServiceName)
	return remote.DoGet[[]model.ClientPublisherInfo](c.Proxy, constant.AdminNamingClientPath+"/service/publisher/list", params, res)
}

func (c *NamingMaintainerClient) GetSubscribeClientList(param vo.GetServiceClientParam) ([]model.ClientSubscriberInfo, error) {
	params := remote.TransformParam(param)
	res := c.buildNamingResource(param.NamespaceId, param.GroupName, param.ServiceName)
	return remote.DoGet[[]model.ClientSubscriberInfo](c.Proxy, constant.AdminNamingClientPath+"/service/subscriber/list", params, res)
}

// --- Health & Cluster ---

func (c *NamingMaintainerClient) GetMetrics(onlyStatus bool) (model.MetricsInfo, error) {
	params := map[string]string{"onlyStatus": strconv.FormatBool(onlyStatus)}
	return remote.DoGet[model.MetricsInfo](c.Proxy, constant.AdminNamingOpsPath+"/metrics", params, remote.BuildEmptyResource())
}

func (c *NamingMaintainerClient) UpdateInstanceHealthStatus(param vo.UpdateInstanceHealthParam) (string, error) {
	params := remote.TransformParam(param)
	res := c.buildNamingResource(param.NamespaceId, param.GroupName, param.ServiceName)
	return remote.DoPut[string](c.Proxy, constant.AdminNamingHealthPath+"/instance", params, res)
}

func (c *NamingMaintainerClient) GetHealthCheckers() (map[string]interface{}, error) {
	resp, err := c.Proxy.ReqApi(constant.AdminNamingHealthPath+"/checkers", nil, http.MethodGet, remote.BuildEmptyResource())
	if err != nil {
		return nil, err
	}
	return remote.ParseResult[map[string]interface{}](resp)
}

func (c *NamingMaintainerClient) UpdateCluster(param vo.UpdateClusterParam) (string, error) {
	params := remote.TransformParam(param)
	res := c.buildNamingResource(param.NamespaceId, param.GroupName, param.ServiceName)
	return remote.DoPut[string](c.Proxy, constant.AdminNamingClusterPath, params, res)
}

func (c *NamingMaintainerClient) SetLogLevel(logName, logLevel string) (string, error) {
	params := map[string]string{
		"logName":  logName,
		"logLevel": logLevel,
	}
	resp, err := c.Proxy.ReqApi(constant.AdminNamingOpsPath+"/log", params, http.MethodPut, remote.BuildEmptyResource())
	if err != nil {
		return "", err
	}
	return remote.ParseResult[string](resp)
}
