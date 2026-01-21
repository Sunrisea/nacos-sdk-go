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
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client/naming_proxy"
	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
	"github.com/nacos-group/nacos-sdk-go/v2/common/redo"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/util"
)

const (
	NamingModule = "naming"
)

// ServiceInfoProcessor is the interface for processing service info
type ServiceInfoProcessor interface {
	ProcessService(service *model.Service)
}

// NamingGrpcRedoService handles redo operations for naming module.
// It implements rpc.IConnectionEventListener interface and can be directly
// registered as connection listener.
type NamingGrpcRedoService struct {
	*redo.AbstractRedoService
	proxy            naming_proxy.INamingProxy
	serviceProcessor ServiceInfoProcessor
}

// NewNamingGrpcRedoService creates a new naming redo service
func NewNamingGrpcRedoService(proxy naming_proxy.INamingProxy, serviceProcessor ServiceInfoProcessor) *NamingGrpcRedoService {
	service := &NamingGrpcRedoService{
		AbstractRedoService: redo.NewAbstractRedoService(NamingModule),
		proxy:               proxy,
		serviceProcessor:    serviceProcessor,
	}
	// Set self as the redo task
	service.SetRedoTask(service)
	return service
}

// SetProxy sets the naming proxy (used for lazy initialization to avoid circular dependency)
func (s *NamingGrpcRedoService) SetProxy(proxy naming_proxy.INamingProxy) {
	s.proxy = proxy
}

// RedoTask implements redo.RedoTask interface
func (s *NamingGrpcRedoService) RedoTask() error {
	// Redo instances first
	if err := s.redoForInstances(); err != nil {
		logger.Errorf("[%s] redo for instances failed: %v", NamingModule, err)
	}

	// Redo subscribes
	if err := s.redoForSubscribes(); err != nil {
		logger.Errorf("[%s] redo for subscribes failed: %v", NamingModule, err)
	}

	return nil
}

// ========== Instance Redo Operations ==========

// redoForInstances processes all instance redo data
func (s *NamingGrpcRedoService) redoForInstances() error {
	redoDataList := s.FindRedoDataByType(InstanceRedoDataType)
	for _, data := range redoDataList {
		instanceData, ok := data.GetData().(*InstanceRedoData)
		if !ok {
			continue
		}
		if err := s.redoForInstance(data, instanceData); err != nil {
			logger.Errorf("[%s] redo instance operation %s for service:%s group:%s failed: %v",
				NamingModule, data.GetRedoType(), instanceData.ServiceName, instanceData.GroupName, err)
		}
	}
	return nil
}

// redoForInstance handles a single instance redo operation
func (s *NamingGrpcRedoService) redoForInstance(redoData *redo.RedoData, instanceData *InstanceRedoData) error {
	if redoData == nil || instanceData == nil {
		return nil
	}

	redoType := redoData.GetRedoType()
	logger.Infof("[%s] redo instance operation %s for service:%s group:%s",
		NamingModule, redoType, instanceData.ServiceName, instanceData.GroupName)

	switch redoType {
	case redo.RedoTypeRegister:
		if s.proxy != nil && !s.proxy.ServerHealthy() {
			logger.Warnf("[%s] server is not healthy, skip register instance", NamingModule)
			return nil
		}
		return s.doRegisterInstance(instanceData)

	case redo.RedoTypeUnregister:
		if s.proxy != nil && !s.proxy.ServerHealthy() {
			logger.Warnf("[%s] server is not healthy, skip deregister instance", NamingModule)
			return nil
		}
		return s.doDeregisterInstance(instanceData)

	case redo.RedoTypeRemove:
		s.RemoveInstanceForRedo(instanceData.ServiceName, instanceData.GroupName)
		return nil

	default:
		return nil
	}
}

// doRegisterInstance performs the actual registration
func (s *NamingGrpcRedoService) doRegisterInstance(redoData *InstanceRedoData) error {
	if s.proxy == nil {
		return fmt.Errorf("proxy is not set")
	}

	var err error
	if redoData.IsBatch {
		_, err = s.proxy.BatchRegisterInstance(redoData.ServiceName, redoData.GroupName, redoData.Instances)
	} else {
		_, err = s.proxy.RegisterInstance(redoData.ServiceName, redoData.GroupName, *redoData.Instance)
	}

	if err != nil {
		return err
	}

	// Mark as registered
	s.InstanceRegistered(redoData.ServiceName, redoData.GroupName)
	return nil
}

// doDeregisterInstance performs the actual deregistration
func (s *NamingGrpcRedoService) doDeregisterInstance(redoData *InstanceRedoData) error {
	if s.proxy == nil {
		return fmt.Errorf("proxy is not set")
	}

	if redoData.Instance == nil {
		return fmt.Errorf("instance is nil")
	}

	_, err := s.proxy.DeregisterInstance(redoData.ServiceName, redoData.GroupName, *redoData.Instance)
	if err != nil {
		return err
	}

	// Mark as deregistered
	s.InstanceDeregistered(redoData.ServiceName, redoData.GroupName)
	return nil
}

// CacheInstanceForRedo caches instance data for redo
func (s *NamingGrpcRedoService) CacheInstanceForRedo(serviceName, groupName string, instance model.Instance) {
	key := util.GetGroupName(serviceName, groupName)
	redoData := NewInstanceRedoData(serviceName, groupName, instance)
	s.CacheRedoData(InstanceRedoDataType, key, redoData)
	logger.Debugf("[%s] cached instance for redo: serviceName=%s, groupName=%s", NamingModule, serviceName, groupName)
}

// CacheInstancesForRedo caches batch instances data for redo
func (s *NamingGrpcRedoService) CacheInstancesForRedo(serviceName, groupName string, instances []model.Instance) {
	key := util.GetGroupName(serviceName, groupName)
	redoData := NewBatchInstanceRedoData(serviceName, groupName, instances)
	s.CacheRedoData(InstanceRedoDataType, key, redoData)
	logger.Debugf("[%s] cached batch instances for redo: serviceName=%s, groupName=%s, count=%d",
		NamingModule, serviceName, groupName, len(instances))
}

// InstanceRegistered marks instance as registered
func (s *NamingGrpcRedoService) InstanceRegistered(serviceName, groupName string) {
	key := util.GetGroupName(serviceName, groupName)
	s.DataRegistered(InstanceRedoDataType, key)
	logger.Debugf("[%s] instance registered: serviceName=%s, groupName=%s", NamingModule, serviceName, groupName)
}

// InstanceDeregister marks instance for deregistration
func (s *NamingGrpcRedoService) InstanceDeregister(serviceName, groupName string) {
	key := util.GetGroupName(serviceName, groupName)
	s.DataDeregister(InstanceRedoDataType, key)
	logger.Debugf("[%s] instance marked for deregister: serviceName=%s, groupName=%s", NamingModule, serviceName, groupName)
}

// InstanceDeregistered marks instance as deregistered
func (s *NamingGrpcRedoService) InstanceDeregistered(serviceName, groupName string) {
	key := util.GetGroupName(serviceName, groupName)
	s.DataDeregistered(InstanceRedoDataType, key)
	logger.Debugf("[%s] instance deregistered: serviceName=%s, groupName=%s", NamingModule, serviceName, groupName)
}

// RemoveInstanceForRedo removes instance from redo cache
func (s *NamingGrpcRedoService) RemoveInstanceForRedo(serviceName, groupName string) {
	key := util.GetGroupName(serviceName, groupName)
	s.RemoveRedoData(InstanceRedoDataType, key)
	logger.Debugf("[%s] removed instance from redo cache: serviceName=%s, groupName=%s", NamingModule, serviceName, groupName)
}

// FindInstanceRedoData finds instance redo data by service key
func (s *NamingGrpcRedoService) FindInstanceRedoData(serviceName, groupName string) *redo.RedoData {
	key := util.GetGroupName(serviceName, groupName)
	if data, ok := s.GetRedoData(InstanceRedoDataType, key); ok {
		return data
	}
	return nil
}

// ========== Subscribe Redo Operations ==========

// redoForSubscribes processes all subscribe redo data
func (s *NamingGrpcRedoService) redoForSubscribes() error {
	redoDataList := s.FindRedoDataByType(SubscribeRedoDataType)
	for _, data := range redoDataList {
		subscribeData, ok := data.GetData().(*SubscribeRedoData)
		if !ok {
			continue
		}

		if err := s.redoForSubscribe(data, subscribeData); err != nil {
			logger.Errorf("[%s] redo subscribe operation %s for service:%s group:%s failed: %v",
				NamingModule, data.GetRedoType(), subscribeData.ServiceName, subscribeData.GroupName, err)
		}
	}
	return nil
}

// redoForSubscribe handles a single subscribe redo operation
func (s *NamingGrpcRedoService) redoForSubscribe(redoData *redo.RedoData, subscribeData *SubscribeRedoData) error {
	if redoData == nil || subscribeData == nil {
		return nil
	}

	redoType := redoData.GetRedoType()
	logger.Infof("[%s] redo subscribe operation %s for service:%s group:%s",
		NamingModule, redoType, subscribeData.ServiceName, subscribeData.GroupName)

	switch redoType {
	case redo.RedoTypeRegister:
		if s.proxy != nil && !s.proxy.ServerHealthy() {
			logger.Warnf("[%s] server is not healthy, skip subscribe", NamingModule)
			return nil
		}
		return s.doSubscribe(subscribeData)

	case redo.RedoTypeUnregister:
		if s.proxy != nil && !s.proxy.ServerHealthy() {
			logger.Warnf("[%s] server is not healthy, skip unsubscribe", NamingModule)
			return nil
		}
		return s.doUnsubscribe(subscribeData)

	case redo.RedoTypeRemove:
		s.RemoveSubscribeForRedo(subscribeData.ServiceName, subscribeData.GroupName, subscribeData.Clusters)
		return nil

	default:
		return nil
	}
}

// doSubscribe performs the actual subscription
func (s *NamingGrpcRedoService) doSubscribe(subscribeData *SubscribeRedoData) error {
	if s.proxy == nil {
		return fmt.Errorf("proxy is not set")
	}

	service, err := s.proxy.Subscribe(subscribeData.ServiceName, subscribeData.GroupName, subscribeData.Clusters)
	if err != nil {
		return err
	}

	// Process service info
	if s.serviceProcessor != nil {
		s.serviceProcessor.ProcessService(&service)
	}

	// Mark as registered
	s.SubscribeRegistered(subscribeData.ServiceName, subscribeData.GroupName, subscribeData.Clusters)
	return nil
}

// doUnsubscribe performs the actual unsubscription
func (s *NamingGrpcRedoService) doUnsubscribe(subscribeData *SubscribeRedoData) error {
	if s.proxy == nil {
		return fmt.Errorf("proxy is not set")
	}

	err := s.proxy.Unsubscribe(subscribeData.ServiceName, subscribeData.GroupName, subscribeData.Clusters)
	if err != nil {
		return err
	}

	// Mark as deregistered
	s.SubscribeDeregistered(subscribeData.ServiceName, subscribeData.GroupName, subscribeData.Clusters)
	return nil
}

// CacheSubscribeForRedo caches subscribe data for redo
func (s *NamingGrpcRedoService) CacheSubscribeForRedo(serviceName, groupName, clusters string) {
	fullServiceName := util.GetGroupName(serviceName, groupName)
	key := util.GetServiceCacheKey(fullServiceName, clusters)
	redoData := NewSubscribeRedoData(serviceName, groupName, clusters)
	s.CacheRedoData(SubscribeRedoDataType, key, redoData)
	logger.Debugf("[%s] cached subscribe for redo: serviceName=%s, groupName=%s, clusters=%s",
		NamingModule, serviceName, groupName, clusters)
}

// IsSubscribeRegistered checks if subscription is registered
func (s *NamingGrpcRedoService) IsSubscribeRegistered(serviceName, groupName, clusters string) bool {
	fullServiceName := util.GetGroupName(serviceName, groupName)
	key := util.GetServiceCacheKey(fullServiceName, clusters)
	return s.IsDataRegistered(SubscribeRedoDataType, key)
}

// SubscribeRegistered marks subscription as registered
func (s *NamingGrpcRedoService) SubscribeRegistered(serviceName, groupName, clusters string) {
	fullServiceName := util.GetGroupName(serviceName, groupName)
	key := util.GetServiceCacheKey(fullServiceName, clusters)
	s.DataRegistered(SubscribeRedoDataType, key)
	logger.Debugf("[%s] subscribe registered: serviceName=%s, groupName=%s, clusters=%s",
		NamingModule, serviceName, groupName, clusters)
}

// SubscribeDeregister marks subscription for deregistration
func (s *NamingGrpcRedoService) SubscribeDeregister(serviceName, groupName, clusters string) {
	fullServiceName := util.GetGroupName(serviceName, groupName)
	key := util.GetServiceCacheKey(fullServiceName, clusters)
	s.DataDeregister(SubscribeRedoDataType, key)
	logger.Debugf("[%s] subscribe marked for deregister: serviceName=%s, groupName=%s, clusters=%s",
		NamingModule, serviceName, groupName, clusters)
}

// SubscribeDeregistered marks subscription as deregistered
func (s *NamingGrpcRedoService) SubscribeDeregistered(serviceName, groupName, clusters string) {
	fullServiceName := util.GetGroupName(serviceName, groupName)
	key := util.GetServiceCacheKey(fullServiceName, clusters)
	s.DataDeregistered(SubscribeRedoDataType, key)
	logger.Debugf("[%s] subscribe deregistered: serviceName=%s, groupName=%s, clusters=%s",
		NamingModule, serviceName, groupName, clusters)
}

// RemoveSubscribeForRedo removes subscription from redo cache
func (s *NamingGrpcRedoService) RemoveSubscribeForRedo(serviceName, groupName, clusters string) {
	fullServiceName := util.GetGroupName(serviceName, groupName)
	key := util.GetServiceCacheKey(fullServiceName, clusters)
	s.RemoveRedoData(SubscribeRedoDataType, key)
	logger.Debugf("[%s] removed subscribe from redo cache: serviceName=%s, groupName=%s, clusters=%s",
		NamingModule, serviceName, groupName, clusters)
}

// IsSubscriberCached checks if a subscriber is cached (for backward compatibility)
func (s *NamingGrpcRedoService) IsSubscriberCached(serviceName, groupName, clusters string) bool {
	fullServiceName := util.GetGroupName(serviceName, groupName)
	key := util.GetServiceCacheKey(fullServiceName, clusters)
	_, ok := s.GetRedoData(SubscribeRedoDataType, key)
	return ok
}
