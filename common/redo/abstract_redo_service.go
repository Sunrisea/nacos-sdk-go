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
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
)

const (
	// RedoTimeout is the timeout for waiting on redo channel
	RedoTimeout = 30 * time.Second
)

// RedoTask is the interface for redo task implementations
type RedoTask interface {
	// RedoTask executes the redo logic
	RedoTask() error
}

// AbstractRedoService provides the base implementation for redo services.
// It manages the redo data map, connection state, and triggers redo tasks
// when connection is restored.
type AbstractRedoService struct {
	mu           sync.RWMutex
	module       string
	connected    bool
	stopChan     chan struct{}
	redoChan     chan struct{}
	redoDataMap  map[string]map[string]*RedoData // dataType -> key -> RedoData
	dataLocks    map[string]*sync.RWMutex        // dataType -> lock
	redoTask     RedoTask
	taskRunning  bool
	taskStopChan chan struct{}
}

// NewAbstractRedoService creates a new AbstractRedoService
func NewAbstractRedoService(module string) *AbstractRedoService {
	return &AbstractRedoService{
		module:      module,
		connected:   false,
		stopChan:    make(chan struct{}),
		redoChan:    make(chan struct{}, 1),
		redoDataMap: make(map[string]map[string]*RedoData),
		dataLocks:   make(map[string]*sync.RWMutex),
	}
}

// SetRedoTask sets the redo task implementation
func (s *AbstractRedoService) SetRedoTask(task RedoTask) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.redoTask = task
}

// Start starts the redo service background task
func (s *AbstractRedoService) Start() {
	s.mu.Lock()
	if s.taskRunning {
		s.mu.Unlock()
		return
	}
	s.taskRunning = true
	s.taskStopChan = make(chan struct{})
	s.mu.Unlock()

	go s.executeRedoLoop()
}

// Stop stops the redo service
func (s *AbstractRedoService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.taskRunning {
		return
	}
	s.taskRunning = false
	close(s.taskStopChan)
}

// executeRedoLoop is the main loop that executes redo tasks
func (s *AbstractRedoService) executeRedoLoop() {
	for {
		select {
		case <-s.taskStopChan:
			logger.Infof("[%s] redo service stopped", s.module)
			return
		case <-s.redoChan:
			s.doRedo()
		case <-time.After(RedoTimeout):
			// Periodic check, no-op but keeps the loop running
			logger.Debugf("[%s] redo task timeout, continuing...", s.module)
		}
	}
}

// doRedo executes the redo task if connected
func (s *AbstractRedoService) doRedo() {
	s.mu.RLock()
	connected := s.connected
	task := s.redoTask
	s.mu.RUnlock()

	if !connected {
		logger.Warnf("[%s] connection is disconnected, skip current redo task", s.module)
		return
	}

	if task == nil {
		logger.Warnf("[%s] redo task is not set", s.module)
		return
	}

	if err := task.RedoTask(); err != nil {
		logger.Errorf("[%s] execute redo task error: %v", s.module, err)
	}
}

// TriggerRedo triggers a redo task execution
func (s *AbstractRedoService) TriggerRedo() {
	select {
	case s.redoChan <- struct{}{}:
	default:
		// Channel is full, redo is already scheduled
	}
}

// OnConnected is called when connection is established
func (s *AbstractRedoService) OnConnected() {
	s.mu.Lock()
	s.connected = true
	s.mu.Unlock()

	logger.Infof("[%s] connection established, triggering redo", s.module)
	s.TriggerRedo()
}

// OnDisConnect is called when connection is lost.
// It marks all registered data as unregistered so they will be re-registered on reconnect.
func (s *AbstractRedoService) OnDisConnect() {
	s.mu.Lock()
	s.connected = false
	s.mu.Unlock()

	logger.Infof("[%s] connection lost, marking all data as unregistered", s.module)

	// Mark all data as unregistered
	for dataType := range s.redoDataMap {
		lock := s.getLockForType(dataType)
		lock.Lock()
		if dataMap, ok := s.redoDataMap[dataType]; ok {
			for _, redoData := range dataMap {
				redoData.SetRegistered(false)
			}
		}
		lock.Unlock()
	}
}

// IsConnected returns whether the service is connected
func (s *AbstractRedoService) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connected
}

// getLockForType returns the lock for a specific data type
func (s *AbstractRedoService) getLockForType(dataType string) *sync.RWMutex {
	s.mu.Lock()
	defer s.mu.Unlock()

	if lock, ok := s.dataLocks[dataType]; ok {
		return lock
	}
	lock := &sync.RWMutex{}
	s.dataLocks[dataType] = lock
	return lock
}

// CacheRedoData stores redo data in the cache
func (s *AbstractRedoService) CacheRedoData(dataType, key string, data *RedoData) {
	lock := s.getLockForType(dataType)
	lock.Lock()
	defer lock.Unlock()

	if _, ok := s.redoDataMap[dataType]; !ok {
		s.redoDataMap[dataType] = make(map[string]*RedoData)
	}
	s.redoDataMap[dataType][key] = data
}

// GetRedoData retrieves redo data from the cache
func (s *AbstractRedoService) GetRedoData(dataType, key string) (*RedoData, bool) {
	lock := s.getLockForType(dataType)
	lock.RLock()
	defer lock.RUnlock()

	if dataMap, ok := s.redoDataMap[dataType]; ok {
		if data, exists := dataMap[key]; exists {
			return data, true
		}
	}
	return nil, false
}

// RemoveRedoData removes redo data from the cache
func (s *AbstractRedoService) RemoveRedoData(dataType, key string) {
	lock := s.getLockForType(dataType)
	lock.Lock()
	defer lock.Unlock()

	if dataMap, ok := s.redoDataMap[dataType]; ok {
		delete(dataMap, key)
	}
}

// FindRedoDataByType finds all redo data of a specific type that needs redo
func (s *AbstractRedoService) FindRedoDataByType(dataType string) []*RedoData {
	lock := s.getLockForType(dataType)
	lock.RLock()
	defer lock.RUnlock()

	var result []*RedoData
	if dataMap, ok := s.redoDataMap[dataType]; ok {
		for _, data := range dataMap {
			if data.IsNeedRedo() {
				result = append(result, data)
			}
		}
	}
	return result
}

// DataRegistered marks the data as registered
func (s *AbstractRedoService) DataRegistered(dataType, key string) {
	lock := s.getLockForType(dataType)
	lock.Lock()
	defer lock.Unlock()

	if dataMap, ok := s.redoDataMap[dataType]; ok {
		if data, exists := dataMap[key]; exists {
			data.Registered()
		}
	}
}

// DataDeregister marks the data for deregistration
func (s *AbstractRedoService) DataDeregister(dataType, key string) {
	lock := s.getLockForType(dataType)
	lock.Lock()
	defer lock.Unlock()

	if dataMap, ok := s.redoDataMap[dataType]; ok {
		if data, exists := dataMap[key]; exists {
			data.SetUnregistering(true)
			data.SetExpectedRegistered(false)
		}
	}
}

// DataDeregistered marks the data as deregistered
func (s *AbstractRedoService) DataDeregistered(dataType, key string) {
	lock := s.getLockForType(dataType)
	lock.Lock()
	defer lock.Unlock()

	if dataMap, ok := s.redoDataMap[dataType]; ok {
		if data, exists := dataMap[key]; exists {
			data.SetRegistered(false)
			data.SetUnregistering(false)
		}
	}
}

// IsDataRegistered checks if the data is registered
func (s *AbstractRedoService) IsDataRegistered(dataType, key string) bool {
	lock := s.getLockForType(dataType)
	lock.RLock()
	defer lock.RUnlock()

	if dataMap, ok := s.redoDataMap[dataType]; ok {
		if data, exists := dataMap[key]; exists {
			return data.IsRegistered()
		}
	}
	return false
}
