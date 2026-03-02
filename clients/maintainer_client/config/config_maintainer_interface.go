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
	"github.com/nacos-group/nacos-sdk-go/v2/clients/maintainer_client/core"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// IConfigMaintainerClient defines config module admin operations.
// It embeds ICoreMaintainerClient to inherit core admin capabilities.
type IConfigMaintainerClient interface {
	core.ICoreMaintainerClient

	// --- Config CRUD ---

	// GetConfig retrieves a configuration by dataId, group, and namespaceId.
	GetConfig(param vo.MaintainerConfigParam) (model.ConfigDetailInfo, error)

	// PublishConfig publishes a configuration with full metadata.
	PublishConfig(param vo.MaintainerPublishConfigParam) (bool, error)

	// UpdateConfigMetadata updates config description and tags without changing content.
	UpdateConfigMetadata(param vo.UpdateConfigMetadataParam) (bool, error)

	// DeleteConfig deletes a single configuration.
	DeleteConfig(param vo.MaintainerConfigParam) (bool, error)

	// DeleteConfigs deletes multiple configurations by their IDs.
	DeleteConfigs(ids []int64) (bool, error)

	// --- Search ---

	// SearchConfig searches/lists configurations with various filters and pagination.
	SearchConfig(param vo.MaintainerSearchConfigParam) (model.Page[model.ConfigBasicInfo], error)

	// GetConfigListByNamespace returns all configs in a namespace.
	GetConfigListByNamespace(namespaceId string) ([]model.ConfigBasicInfo, error)

	// --- Listener ---

	// GetListeners retrieves the listeners for a specific configuration.
	GetListeners(param vo.GetListenersParam) (model.ConfigListenerInfo, error)

	// GetAllSubClientConfigByIp gets all subscribed client configs by IP.
	GetAllSubClientConfigByIp(param vo.GetSubClientConfigParam) (model.ConfigListenerInfo, error)

	// --- Clone ---

	// CloneConfig clones configurations within the same namespace (uses JSON body).
	CloneConfig(param vo.CloneConfigParam) (map[string]interface{}, error)

	// --- Beta ---

	// PublishBetaConfig publishes a beta configuration to specific IPs.
	PublishBetaConfig(param vo.PublishBetaConfigParam) (bool, error)

	// StopBeta stops the beta release for a configuration.
	StopBeta(param vo.StopBetaParam) (bool, error)

	// QueryBeta queries the current beta/gray configuration info.
	QueryBeta(param vo.StopBetaParam) (model.ConfigGrayInfo, error)

	// --- History ---

	// ListConfigHistory lists config change history with pagination.
	ListConfigHistory(param vo.ListConfigHistoryParam) (model.Page[model.ConfigHistoryInfo], error)

	// GetConfigHistoryInfo retrieves a specific config history entry by nid.
	GetConfigHistoryInfo(param vo.GetConfigHistoryParam) (model.ConfigHistoryInfo, error)

	// GetPreviousConfigHistoryInfo retrieves the previous config version by id.
	GetPreviousConfigHistoryInfo(dataId, group, namespaceId string, id int64) (model.ConfigHistoryInfo, error)

	// --- Ops ---

	// UpdateLocalCacheFromStore refreshes the local cache from the config store.
	UpdateLocalCacheFromStore() (string, error)

	// SetLogLevel sets the log level for the config module.
	SetLogLevel(logName, logLevel string) (string, error)
}
