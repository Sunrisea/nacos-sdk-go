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
	"github.com/nacos-group/nacos-sdk-go/v2/common/remote/rpc/rpc_request"
	"github.com/nacos-group/nacos-sdk-go/v2/common/remote/rpc/rpc_response"
	"google.golang.org/grpc"
)

type IConnection interface {
	request(request rpc_request.IRequest, timeoutMills int64, client *RpcClient) (rpc_response.IResponse, error)
	close()
	getConnectionId() string
	getServerInfo() ServerInfo
	setAbandon(flag bool)
	getAbandon() bool
	// Ability negotiation methods
	SetAbilityTable(abilities map[string]bool)
	GetAbilityTable() map[string]bool
	IsAbilitiesSet() bool
	SetAbandon(abandon bool)
	GetConnectionAbility(abilityKey AbilityKey) AbilityStatus
}

type Connection struct {
	conn         *grpc.ClientConn
	connectionId string
	abandon      bool
	serverInfo   ServerInfo
	abilityTable map[string]bool
	abilitiesSet bool
}

func (c *Connection) getConnectionId() string {
	return c.connectionId
}

func (c *Connection) getServerInfo() ServerInfo {
	return c.serverInfo
}

func (c *Connection) setAbandon(flag bool) {
	c.abandon = flag
}

func (c *Connection) getAbandon() bool {
	return c.abandon
}

func (c *Connection) close() {
	_ = c.conn.Close()
}

// SetAbilityTable sets the server's ability table
func (c *Connection) SetAbilityTable(abilities map[string]bool) {
	c.abilityTable = abilities
	c.abilitiesSet = true
}

// GetAbilityTable returns the server's ability table
func (c *Connection) GetAbilityTable() map[string]bool {
	return c.abilityTable
}

// IsAbilitiesSet returns whether the ability table has been set
func (c *Connection) IsAbilitiesSet() bool {
	return c.abilitiesSet
}

// SetAbandon sets the abandon flag (exported version)
func (c *Connection) SetAbandon(abandon bool) {
	c.abandon = abandon
}

// GetConnectionAbility returns the status of a specific ability
func (c *Connection) GetConnectionAbility(abilityKey AbilityKey) AbilityStatus {
	if c.abilityTable == nil {
		return AbilityStatusUnknown
	}
	if supported, ok := c.abilityTable[abilityKey.KeyName]; ok && supported {
		return AbilityStatusSupported
	}
	return AbilityStatusNotSupported
}
