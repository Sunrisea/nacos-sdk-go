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

package redo

import (
	"github.com/nacos-group/nacos-sdk-go/v2/common/redo"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
)

// Redo data type constants
const (
	InstanceRedoDataType  = "InstanceRedoData"
	SubscribeRedoDataType = "SubscribeRedoData"
)

// InstanceRedoData represents instance registration data for redo.
// This struct is stored in RedoData.data field to preserve all necessary information.
type InstanceRedoData struct {
	ServiceName string
	GroupName   string
	Instance    *model.Instance
	Instances   []model.Instance // For batch registration
	IsBatch     bool
}

// NewInstanceRedoData creates a new redo.RedoData for single instance
func NewInstanceRedoData(serviceName, groupName string, instance model.Instance) *redo.RedoData {
	data := &InstanceRedoData{
		ServiceName: serviceName,
		GroupName:   groupName,
		Instance:    &instance,
		IsBatch:     false,
	}
	return redo.NewRedoData(data)
}

// NewBatchInstanceRedoData creates a new redo.RedoData for batch instances
func NewBatchInstanceRedoData(serviceName, groupName string, instances []model.Instance) *redo.RedoData {
	data := &InstanceRedoData{
		ServiceName: serviceName,
		GroupName:   groupName,
		Instances:   instances,
		IsBatch:     true,
	}
	return redo.NewRedoData(data)
}

// GetInstance returns the single instance (for non-batch)
func (d *InstanceRedoData) GetInstance() *model.Instance {
	return d.Instance
}

// GetInstances returns the instances (for batch)
func (d *InstanceRedoData) GetInstances() []model.Instance {
	return d.Instances
}

// SubscribeRedoData represents subscription data for redo.
// This struct is stored in RedoData.data field to preserve all necessary information.
type SubscribeRedoData struct {
	ServiceName string
	GroupName   string
	Clusters    string
}

// NewSubscribeRedoData creates a new redo.RedoData for subscription
func NewSubscribeRedoData(serviceName, groupName, clusters string) *redo.RedoData {
	data := &SubscribeRedoData{
		ServiceName: serviceName,
		GroupName:   groupName,
		Clusters:    clusters,
	}
	return redo.NewRedoData(data)
}
