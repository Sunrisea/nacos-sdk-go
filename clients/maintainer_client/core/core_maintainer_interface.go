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
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// ICoreMaintainerClient defines the core module admin operations.
type ICoreMaintainerClient interface {

	// GetServerState returns the server state key-value map.
	GetServerState() (map[string]string, error)

	// Liveness checks server liveness.
	Liveness() (bool, error)

	// Readiness checks server readiness.
	Readiness() (bool, error)

	// RaftOps executes a Raft operation.
	RaftOps(command, value, groupId string) (string, error)

	// GetIdGenerators returns the status of all ID generators.
	GetIdGenerators() ([]model.IdGeneratorInfo, error)

	// UpdateLogLevel changes the log level for a specific logger.
	UpdateLogLevel(logName, logLevel string) error

	// ListClusterNodes lists cluster nodes filtered by address and state.
	ListClusterNodes(address, state string) ([]model.NacosMember, error)

	// UpdateLookupMode updates the cluster lookup mode.
	UpdateLookupMode(lookupType string) (bool, error)

	// GetCurrentClients returns current client connections.
	GetCurrentClients() (map[string]model.ConnectionInfo, error)

	// ReloadConnectionCount reloads SDK connection count on the current server.
	ReloadConnectionCount(count int, redirectAddress string) (string, error)

	// SmartReloadCluster smartly reloads the cluster based on a loader factor.
	SmartReloadCluster(loaderFactor string) (string, error)

	// ReloadSingleClient reloads a single client connection.
	ReloadSingleClient(connectionId, redirectAddress string) (string, error)

	// GetClusterLoaderMetrics retrieves the current cluster loader metrics.
	GetClusterLoaderMetrics() (model.ServerLoaderMetrics, error)

	// GetNamespaceList returns all namespaces.
	GetNamespaceList() ([]model.Namespace, error)

	// GetNamespace returns a specific namespace by ID.
	GetNamespace(namespaceId string) (model.Namespace, error)

	// CreateNamespace creates a new namespace.
	CreateNamespace(param vo.CreateNamespaceParam) (bool, error)

	// UpdateNamespace updates an existing namespace.
	UpdateNamespace(param vo.UpdateNamespaceParam) (bool, error)

	// DeleteNamespace deletes a namespace by ID.
	DeleteNamespace(namespaceId string) (bool, error)

	// CheckNamespaceIdExist checks if a namespace ID already exists.
	CheckNamespaceIdExist(namespaceId string) (bool, error)

	// CloseClient shuts down the client and releases resources.
	CloseClient()
}
