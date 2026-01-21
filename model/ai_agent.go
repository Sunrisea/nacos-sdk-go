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

package model

import (
	"github.com/a2aproject/a2a-go/a2a"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
)

// AgentEndpoint represents the endpoint information for an A2A agent
type AgentEndpoint struct {
	Transport  string `json:"transport,omitempty"`
	Address    string `json:"address,omitempty"`
	Port       int    `json:"port,omitempty"`
	Path       string `json:"path,omitempty"`
	SupportTLS bool   `json:"supportTls,omitempty"`
	Version    string `json:"version,omitempty"`
}

// NewAgentEndpoint creates a new AgentEndpoint with default values
func NewAgentEndpoint() *AgentEndpoint {
	return &AgentEndpoint{
		Transport:  constant.A2A_ENDPOINT_DEFAULT_TRANSPORT,
		Path:       "",
		SupportTLS: false,
	}
}

// Equal checks if two AgentEndpoint instances are equal
func (e *AgentEndpoint) Equal(other *AgentEndpoint) bool {
	if other == nil {
		return false
	}
	return e.Transport == other.Transport &&
		e.Address == other.Address &&
		e.Port == other.Port &&
		e.Path == other.Path &&
		e.SupportTLS == other.SupportTLS &&
		e.Version == other.Version
}

// AgentCardDetailInfo extends a2a.AgentCard with Nacos-specific fields
type AgentCardDetailInfo struct {
	*a2a.AgentCard
	RegistrationType string `json:"registrationType,omitempty"`
	LatestVersion    bool   `json:"latestVersion,omitempty"`
}

// NewAgentCardDetailInfo creates a new AgentCardDetailInfo with default values
func NewAgentCardDetailInfo() *AgentCardDetailInfo {
	return &AgentCardDetailInfo{
		AgentCard:        &a2a.AgentCard{},
		RegistrationType: constant.A2A_ENDPOINT_TYPE_URL,
		LatestVersion:    false,
	}
}

// AgentCardDetailInfoFromCard creates an AgentCardDetailInfo from an a2a.AgentCard
func AgentCardDetailInfoFromCard(card *a2a.AgentCard) *AgentCardDetailInfo {
	return &AgentCardDetailInfo{
		AgentCard:        card,
		RegistrationType: constant.A2A_ENDPOINT_TYPE_URL,
		LatestVersion:    false,
	}
}
