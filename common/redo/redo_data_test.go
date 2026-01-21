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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRedoData(t *testing.T) {
	data := "test data"
	redoData := NewRedoData(data)

	assert.NotNil(t, redoData)
	assert.Equal(t, data, redoData.GetData())
	assert.True(t, redoData.IsExpectedRegistered())
	assert.False(t, redoData.IsRegistered())
	assert.False(t, redoData.IsUnregistering())
}

func TestRedoDataStateTransitions(t *testing.T) {
	redoData := NewRedoData("test")

	// Initial state: not registered, expected to register
	assert.Equal(t, RedoTypeRegister, redoData.GetRedoType())
	assert.True(t, redoData.IsNeedRedo())

	// After registration
	redoData.Registered()
	assert.Equal(t, RedoTypeNone, redoData.GetRedoType())
	assert.False(t, redoData.IsNeedRedo())

	// Start unregistering
	redoData.SetExpectedRegistered(false)
	assert.Equal(t, RedoTypeUnregister, redoData.GetRedoType())
	assert.True(t, redoData.IsNeedRedo())

	// After unregistration completes
	redoData.SetRegistered(false)
	redoData.SetUnregistering(true)
	assert.Equal(t, RedoTypeRemove, redoData.GetRedoType())
}

func TestRedoTypeRegisterAfterDisconnect(t *testing.T) {
	redoData := NewRedoData("test")

	// Initial registration
	redoData.Registered()
	assert.Equal(t, RedoTypeNone, redoData.GetRedoType())

	// Simulate disconnect - reset registered state
	redoData.SetRegistered(false)
	assert.Equal(t, RedoTypeRegister, redoData.GetRedoType())
	assert.True(t, redoData.IsNeedRedo())
}

func TestRedoDataConcurrency(t *testing.T) {
	redoData := NewRedoData("test")

	done := make(chan bool)

	// Concurrent reads and writes
	go func() {
		for i := 0; i < 100; i++ {
			redoData.SetRegistered(true)
			redoData.SetRegistered(false)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			_ = redoData.IsRegistered()
			_ = redoData.GetRedoType()
		}
		done <- true
	}()

	<-done
	<-done
}

func TestRedoTypeAllStates(t *testing.T) {
	tests := []struct {
		name               string
		registered         bool
		unregistering      bool
		expectedRegistered bool
		expectedRedoType   RedoType
	}{
		{
			name:               "registered, not unregistering, expected registered -> NONE",
			registered:         true,
			unregistering:      false,
			expectedRegistered: true,
			expectedRedoType:   RedoTypeNone,
		},
		{
			name:               "registered, not unregistering, not expected -> UNREGISTER",
			registered:         true,
			unregistering:      false,
			expectedRegistered: false,
			expectedRedoType:   RedoTypeUnregister,
		},
		{
			name:               "registered, unregistering -> UNREGISTER",
			registered:         true,
			unregistering:      true,
			expectedRegistered: true,
			expectedRedoType:   RedoTypeUnregister,
		},
		{
			name:               "not registered, not unregistering -> REGISTER",
			registered:         false,
			unregistering:      false,
			expectedRegistered: true,
			expectedRedoType:   RedoTypeRegister,
		},
		{
			name:               "not registered, unregistering, expected -> REGISTER",
			registered:         false,
			unregistering:      true,
			expectedRegistered: true,
			expectedRedoType:   RedoTypeRegister,
		},
		{
			name:               "not registered, unregistering, not expected -> REMOVE",
			registered:         false,
			unregistering:      true,
			expectedRegistered: false,
			expectedRedoType:   RedoTypeRemove,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redoData := NewRedoData("test")
			redoData.SetRegistered(tt.registered)
			redoData.SetUnregistering(tt.unregistering)
			redoData.SetExpectedRegistered(tt.expectedRegistered)

			assert.Equal(t, tt.expectedRedoType, redoData.GetRedoType())
		})
	}
}
