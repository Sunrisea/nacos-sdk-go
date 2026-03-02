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
	"github.com/nacos-group/nacos-sdk-go/v2/clients/maintainer_client/core"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// INamingMaintainerClient defines naming module admin operations.
// It embeds ICoreMaintainerClient to inherit core admin capabilities.
type INamingMaintainerClient interface {
	core.ICoreMaintainerClient

	// --- Service ---

	CreateService(param vo.MaintainerServiceParam) (string, error)
	UpdateService(param vo.MaintainerServiceParam) (string, error)
	RemoveService(param vo.MaintainerServiceParam) (string, error)
	GetServiceDetail(param vo.MaintainerServiceParam) (model.ServiceDetailInfo, error)
	ListServices(param vo.ListServicesParam) (model.Page[model.ServiceView], error)
	ListServicesWithDetail(param vo.ListServicesDetailParam) (model.Page[model.ServiceDetailInfo], error)
	GetSubscribers(param vo.GetSubscribersParam) (model.Page[model.SubscriberInfo], error)
	ListSelectorTypes() ([]string, error)

	// --- Instance ---

	RegisterInstance(param vo.MaintainerInstanceParam) (string, error)
	DeregisterInstance(param vo.MaintainerInstanceParam) (string, error)
	UpdateInstance(param vo.MaintainerInstanceParam) (string, error)
	BatchUpdateInstanceMetadata(param vo.BatchInstanceMetadataParam) (model.InstanceMetadataBatchResult, error)
	BatchDeleteInstanceMetadata(param vo.BatchInstanceMetadataParam) (model.InstanceMetadataBatchResult, error)
	PartialUpdateInstance(param vo.MaintainerInstanceParam) (string, error)
	ListInstances(param vo.ListInstancesParam) ([]model.Instance, error)
	GetInstanceDetail(param vo.GetInstanceDetailParam) (model.Instance, error)

	// --- Naming Client Info ---

	GetClientList() ([]string, error)
	GetClientDetail(clientId string) (model.ClientSummaryInfo, error)
	GetPublishedServiceList(clientId string) ([]model.ClientServiceInfo, error)
	GetSubscribeServiceList(clientId string) ([]model.ClientServiceInfo, error)
	GetPublishedClientList(param vo.GetServiceClientParam) ([]model.ClientPublisherInfo, error)
	GetSubscribeClientList(param vo.GetServiceClientParam) ([]model.ClientSubscriberInfo, error)

	// --- Health & Cluster ---

	GetMetrics(onlyStatus bool) (model.MetricsInfo, error)
	UpdateInstanceHealthStatus(param vo.UpdateInstanceHealthParam) (string, error)
	GetHealthCheckers() (map[string]interface{}, error)
	UpdateCluster(param vo.UpdateClusterParam) (string, error)

	// --- Ops ---

	// SetLogLevel sets the log level for the naming module.
	SetLogLevel(logName, logLevel string) (string, error)
}
