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

package rpc

import (
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
)

// RecAbilityContext manages the ability negotiation context during connection setup
type RecAbilityContext struct {
	mu         sync.Mutex
	blocker    chan struct{}
	needToSync bool
	connection IConnection
}

// NewRecAbilityContext creates a new RecAbilityContext
func NewRecAbilityContext() *RecAbilityContext {
	return &RecAbilityContext{
		blocker:    make(chan struct{}, 1),
		needToSync: false,
	}
}

// Reset resets the context for a new connection and marks that we need to sync abilities
func (r *RecAbilityContext) Reset(connection IConnection) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connection = connection
	r.blocker = make(chan struct{}, 1)
	r.needToSync = true
}

// Release releases the blocker with the received abilities from server
func (r *RecAbilityContext) Release(abilities map[string]bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.connection != nil {
		r.connection.SetAbilityTable(abilities)
		r.connection = nil
	}

	// Signal that abilities have been received
	select {
	case r.blocker <- struct{}{}:
	default:
	}

	r.needToSync = false
}

// IsNeedToSync returns whether we need to wait for abilities sync
func (r *RecAbilityContext) IsNeedToSync() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.needToSync
}

// AwaitAbilities waits for the server abilities with a timeout
func (r *RecAbilityContext) AwaitAbilities(timeoutMs int64) error {
	timeout := time.Duration(timeoutMs) * time.Millisecond
	select {
	case <-r.blocker:
		return nil
	case <-time.After(timeout):
		logger.Warnf("Waiting for server abilities timeout after %d ms", timeoutMs)
		return ErrAbilityNegotiationTimeout
	}
}

// Check checks if the connection has received the ability table
func (r *RecAbilityContext) Check(connection IConnection) bool {
	if !connection.IsAbilitiesSet() {
		logger.Errorf("Client didn't receive server abilities table even empty table but server supports ability negotiation")
		connection.SetAbandon(true)
		return false
	}
	return true
}
