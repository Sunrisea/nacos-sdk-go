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

package ai

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/maintainer_client/core"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/maintainer_client/remote"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/nacos_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/common/security"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// AiMaintainerClient implements IAiMaintainerClient.
type AiMaintainerClient struct {
	*core.CoreMaintainerClient
}

func NewAiMaintainerClient(nc nacos_client.INacosClient, provider security.RamCredentialProvider) (*AiMaintainerClient, error) {
	coreClient, err := core.NewCoreMaintainerClient(nc, provider)
	if err != nil {
		return nil, err
	}
	return &AiMaintainerClient{CoreMaintainerClient: coreClient}, nil
}

func toJsonString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func (c *AiMaintainerClient) buildAiResource(namespace, resourceName string) security.RequestResource {
	return security.BuildAdminResource(namespace, "DEFAULT_GROUP", resourceName)
}

// --- MCP Server ---

func (c *AiMaintainerClient) ListMcpServer(param vo.ListMcpServerParam) (model.Page[model.McpServerBasicInfo], error) {
	params := remote.TransformParam(param)
	params["search"] = "accurate"
	res := c.buildAiResource(param.NamespaceId, param.McpName)
	return remote.DoGet[model.Page[model.McpServerBasicInfo]](c.Proxy, constant.AdminAiMcpPath+"/list", params, res)
}

func (c *AiMaintainerClient) SearchMcpServer(param vo.SearchMcpServerParam) (model.Page[model.McpServerBasicInfo], error) {
	params := remote.TransformParam(param)
	params["search"] = "blur"
	res := c.buildAiResource(param.NamespaceId, param.McpName)
	return remote.DoGet[model.Page[model.McpServerBasicInfo]](c.Proxy, constant.AdminAiMcpPath+"/list", params, res)
}

func (c *AiMaintainerClient) GetMcpServerDetail(param vo.GetMcpServerDetailParam) (model.McpServerDetailInfo, error) {
	params := remote.TransformParam(param)
	res := c.buildAiResource(param.NamespaceId, param.McpName)
	return remote.DoGet[model.McpServerDetailInfo](c.Proxy, constant.AdminAiMcpPath, params, res)
}

func (c *AiMaintainerClient) CreateMcpServer(param vo.CreateMcpServerParam) (string, error) {
	params := map[string]string{
		"namespaceId": param.NamespaceId,
		"mcpName":     param.McpName,
	}
	if param.ServerSpec != nil {
		params["serverSpecification"] = toJsonString(param.ServerSpec)
	}
	if param.ToolSpec != nil {
		params["toolSpecification"] = toJsonString(param.ToolSpec)
	}
	if param.EndpointSpec != nil {
		params["endpointSpecification"] = toJsonString(param.EndpointSpec)
	}
	res := c.buildAiResource(param.NamespaceId, param.McpName)
	return remote.DoPost[string](c.Proxy, constant.AdminAiMcpPath, params, res)
}

func (c *AiMaintainerClient) UpdateMcpServer(param vo.UpdateMcpServerParam) (bool, error) {
	params := map[string]string{
		"namespaceId":      param.NamespaceId,
		"mcpName":          param.McpName,
		"latest":           strconv.FormatBool(param.IsLatest),
		"overrideExisting": strconv.FormatBool(param.OverrideExisting),
	}
	if param.ServerSpec != nil {
		params["serverSpecification"] = toJsonString(param.ServerSpec)
	}
	if param.ToolSpec != nil {
		params["toolSpecification"] = toJsonString(param.ToolSpec)
	}
	if param.EndpointSpec != nil {
		params["endpointSpecification"] = toJsonString(param.EndpointSpec)
	}
	res := c.buildAiResource(param.NamespaceId, param.McpName)
	resp, err := c.Proxy.ReqApi(constant.AdminAiMcpPath, params, http.MethodPut, res)
	if err != nil {
		return false, err
	}
	_, err = remote.ParseResult[string](resp)
	return err == nil, err
}

func (c *AiMaintainerClient) DeleteMcpServer(param vo.DeleteMcpServerParam) (bool, error) {
	params := remote.TransformParam(param)
	res := c.buildAiResource(param.NamespaceId, param.McpName)
	resp, err := c.Proxy.ReqApi(constant.AdminAiMcpPath, params, http.MethodDelete, res)
	if err != nil {
		return false, err
	}
	_, err = remote.ParseResult[string](resp)
	return err == nil, err
}

// --- A2A Agent ---

func (c *AiMaintainerClient) RegisterAgent(param vo.RegisterAgentParam) (bool, error) {
	params := map[string]string{}
	if param.AgentCard != nil {
		params["agentCard"] = toJsonString(param.AgentCard)
	}
	if param.NamespaceId != "" {
		params["namespaceId"] = param.NamespaceId
	}
	if param.AgentName != "" {
		params["agentName"] = param.AgentName
	}
	if param.RegistrationType != "" {
		params["registrationType"] = param.RegistrationType
	}
	res := c.buildAiResource(param.NamespaceId, param.AgentName)
	resp, err := c.Proxy.ReqApi(constant.AdminAiA2aPath, params, http.MethodPost, res)
	if err != nil {
		return false, err
	}
	_, err = remote.ParseResult[string](resp)
	return err == nil, err
}

func (c *AiMaintainerClient) GetAgentCard(param vo.GetMaintainerAgentCardParam) (model.AgentCardDetailInfo, error) {
	params := remote.TransformParam(param)
	res := c.buildAiResource(param.NamespaceId, param.AgentName)
	return remote.DoGet[model.AgentCardDetailInfo](c.Proxy, constant.AdminAiA2aPath, params, res)
}

func (c *AiMaintainerClient) UpdateAgentCard(param vo.UpdateAgentCardParam) (bool, error) {
	params := map[string]string{}
	if param.AgentCard != nil {
		params["agentCard"] = toJsonString(param.AgentCard)
	}
	if param.NamespaceId != "" {
		params["namespaceId"] = param.NamespaceId
	}
	if param.AgentName != "" {
		params["agentName"] = param.AgentName
	}
	params["setAsLatest"] = strconv.FormatBool(param.SetAsLatest)
	if param.RegistrationType != "" {
		params["registrationType"] = param.RegistrationType
	}
	res := c.buildAiResource(param.NamespaceId, param.AgentName)
	resp, err := c.Proxy.ReqApi(constant.AdminAiA2aPath, params, http.MethodPut, res)
	if err != nil {
		return false, err
	}
	_, err = remote.ParseResult[string](resp)
	return err == nil, err
}

func (c *AiMaintainerClient) DeleteAgent(param vo.DeleteAgentParam) (bool, error) {
	params := remote.TransformParam(param)
	res := c.buildAiResource(param.NamespaceId, param.AgentName)
	resp, err := c.Proxy.ReqApi(constant.AdminAiA2aPath, params, http.MethodDelete, res)
	if err != nil {
		return false, err
	}
	_, err = remote.ParseResult[string](resp)
	return err == nil, err
}

func (c *AiMaintainerClient) ListAllVersionOfAgent(param vo.ListAgentVersionParam) ([]model.AgentVersionDetail, error) {
	params := remote.TransformParam(param)
	res := c.buildAiResource(param.NamespaceId, param.AgentName)
	return remote.DoGet[[]model.AgentVersionDetail](c.Proxy, constant.AdminAiA2aPath+"/version/list", params, res)
}

func (c *AiMaintainerClient) SearchAgentCards(param vo.SearchAgentParam) (model.Page[model.AgentCardVersionInfo], error) {
	params := remote.TransformParam(param)
	params["search"] = "blur"
	res := c.buildAiResource(param.NamespaceId, "")
	return remote.DoGet[model.Page[model.AgentCardVersionInfo]](c.Proxy, constant.AdminAiA2aPath+"/list", params, res)
}

func (c *AiMaintainerClient) ListAgentCards(param vo.ListAgentCardsParam) (model.Page[model.AgentCardVersionInfo], error) {
	params := remote.TransformParam(param)
	params["search"] = "accurate"
	res := c.buildAiResource(param.NamespaceId, "")
	return remote.DoGet[model.Page[model.AgentCardVersionInfo]](c.Proxy, constant.AdminAiA2aPath+"/list", params, res)
}
