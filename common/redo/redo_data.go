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

import "sync"

// RedoType represents the type of redo operation needed
type RedoType string

const (
	// RedoTypeRegister indicates the data needs to be registered
	RedoTypeRegister RedoType = "REGISTER"
	// RedoTypeUnregister indicates the data needs to be unregistered
	RedoTypeUnregister RedoType = "UNREGISTER"
	// RedoTypeNone indicates no redo operation needed
	RedoTypeNone RedoType = "NONE"
	// RedoTypeRemove indicates the data should be removed from local cache
	RedoTypeRemove RedoType = "REMOVE"
)

// RedoData represents data that may need to be re-registered after reconnection.
// It tracks the expected state, current registered state, and unregistering state
// to determine what redo operation is needed.
type RedoData struct {
	mu                 sync.RWMutex
	data               interface{}
	expectedRegistered bool // Whether the data is expected to be registered
	registered         bool // Whether the data is currently registered
	unregistering      bool // Whether the data is in the process of being unregistered
}

// NewRedoData creates a new RedoData with the given data.
// By default, expectedRegistered is true and registered is false.
func NewRedoData(data interface{}) *RedoData {
	return &RedoData{
		data:               data,
		expectedRegistered: true,
		registered:         false,
		unregistering:      false,
	}
}

// GetData returns the underlying data
func (r *RedoData) GetData() interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.data
}

// SetData updates the underlying data
func (r *RedoData) SetData(data interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data = data
}

// IsRegistered returns whether the data is currently registered
func (r *RedoData) IsRegistered() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.registered
}

// SetRegistered sets the registered state
func (r *RedoData) SetRegistered(registered bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registered = registered
}

// Registered marks the data as registered
func (r *RedoData) Registered() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registered = true
}

// IsUnregistering returns whether the data is being unregistered
func (r *RedoData) IsUnregistering() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.unregistering
}

// SetUnregistering sets the unregistering state
func (r *RedoData) SetUnregistering(unregistering bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.unregistering = unregistering
}

// IsExpectedRegistered returns whether the data is expected to be registered
func (r *RedoData) IsExpectedRegistered() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.expectedRegistered
}

// SetExpectedRegistered sets the expected registered state
func (r *RedoData) SetExpectedRegistered(expected bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.expectedRegistered = expected
}

// GetRedoType determines what redo operation is needed based on current state.
// State transition logic:
//   - registered=true, unregistering=false, expected=true  -> NONE (already registered)
//   - registered=true, unregistering=false, expected=false -> UNREGISTER (need to unregister)
//   - registered=true, unregistering=true, expected=*      -> UNREGISTER (continue unregistering)
//   - registered=false, unregistering=false, expected=*    -> REGISTER (need to register)
//   - registered=false, unregistering=true, expected=true  -> REGISTER (need to re-register)
//   - registered=false, unregistering=true, expected=false -> REMOVE (remove from cache)
func (r *RedoData) GetRedoType() RedoType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.registered && !r.unregistering {
		if r.expectedRegistered {
			return RedoTypeNone
		}
		return RedoTypeUnregister
	}

	if r.registered && r.unregistering {
		return RedoTypeUnregister
	}

	if !r.registered && !r.unregistering {
		return RedoTypeRegister
	}

	// !registered && unregistering
	if r.expectedRegistered {
		return RedoTypeRegister
	}
	return RedoTypeRemove
}

// IsNeedRedo returns true if redo operation is needed
func (r *RedoData) IsNeedRedo() bool {
	return r.GetRedoType() != RedoTypeNone
}
