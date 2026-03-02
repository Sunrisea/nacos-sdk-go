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

package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/common/http_agent"
	"github.com/nacos-group/nacos-sdk-go/v2/common/nacos_server"
	"github.com/nacos-group/nacos-sdk-go/v2/common/security"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/util"
)

type MaintainerHttpProxy struct {
	nacosServer *nacos_server.NacosServer
	clientCfg   constant.ClientConfig
}

func NewMaintainerHttpProxy(ctx context.Context, serverCfgs []constant.ServerConfig,
	clientCfg constant.ClientConfig, httpAgent http_agent.IHttpAgent,
	provider security.RamCredentialProvider) (*MaintainerHttpProxy, error) {

	nacosServer, err := nacos_server.NewNacosServerWithRamCredentialProvider(
		ctx, serverCfgs, clientCfg, httpAgent, clientCfg.TimeoutMs, clientCfg.Endpoint,
		nil, provider)
	if err != nil {
		return nil, err
	}

	return &MaintainerHttpProxy{
		nacosServer: nacosServer,
		clientCfg:   clientCfg,
	}, nil
}

// ReqApi sends a form-urlencoded request (GET/DELETE/POST/PUT).
func (p *MaintainerHttpProxy) ReqApi(api string, params map[string]string,
	method string, resource security.RequestResource) (string, error) {
	return p.nacosServer.ReqAdminApi(api, params, nil, nil, method, resource, p.clientCfg.TimeoutMs)
}

// ReqApiWithHeaders sends a form-urlencoded request with custom headers.
func (p *MaintainerHttpProxy) ReqApiWithHeaders(api string, params map[string]string,
	headers map[string]string, method string, resource security.RequestResource) (string, error) {
	return p.nacosServer.ReqAdminApi(api, params, headers, nil, method, resource, p.clientCfg.TimeoutMs)
}

// ReqApiWithJsonBody sends a JSON body request (POST/PUT).
// Params are sent as query string, body is JSON-serialized from the given object.
func (p *MaintainerHttpProxy) ReqApiWithJsonBody(api string, params map[string]string,
	body interface{}, method string, resource security.RequestResource) (string, error) {

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal JSON body")
	}
	return p.nacosServer.ReqAdminApi(api, params, nil, bodyBytes, method, resource, p.clientCfg.TimeoutMs)
}

// ParseResult parses the Nacos v3 unified response format into RestResult[T].
// Returns the Data field of type T if code == 0, otherwise returns an error.
func ParseResult[T any](response string) (T, error) {
	var result model.RestResult[T]
	var zero T
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return zero, errors.Wrap(err, "failed to unmarshal response")
	}
	if result.Code != 0 {
		return zero, fmt.Errorf("server returned error code %d: %s", result.Code, result.Message)
	}
	return result.Data, nil
}

// ParseBoolResult parses a response where Data is a boolean value.
func ParseBoolResult(response string) (bool, error) {
	return ParseResult[bool](response)
}

// ParseStringResult parses a response where Data is a string value.
func ParseStringResult(response string) (string, error) {
	return ParseResult[string](response)
}

// --- Convenience methods that combine request + parse ---

// DoGet sends a GET request and parses the result.
func DoGet[T any](p *MaintainerHttpProxy, api string, params map[string]string,
	resource security.RequestResource) (T, error) {
	var zero T
	resp, err := p.ReqApi(api, params, http.MethodGet, resource)
	if err != nil {
		return zero, err
	}
	return ParseResult[T](resp)
}

// DoPost sends a form-urlencoded POST request and parses the result.
func DoPost[T any](p *MaintainerHttpProxy, api string, params map[string]string,
	resource security.RequestResource) (T, error) {
	var zero T
	resp, err := p.ReqApi(api, params, http.MethodPost, resource)
	if err != nil {
		return zero, err
	}
	return ParseResult[T](resp)
}

// DoPostJson sends a JSON body POST request and parses the result.
func DoPostJson[T any](p *MaintainerHttpProxy, api string, params map[string]string,
	body interface{}, resource security.RequestResource) (T, error) {
	var zero T
	resp, err := p.ReqApiWithJsonBody(api, params, body, http.MethodPost, resource)
	if err != nil {
		return zero, err
	}
	return ParseResult[T](resp)
}

// DoPut sends a form-urlencoded PUT request and parses the result.
func DoPut[T any](p *MaintainerHttpProxy, api string, params map[string]string,
	resource security.RequestResource) (T, error) {
	var zero T
	resp, err := p.ReqApi(api, params, http.MethodPut, resource)
	if err != nil {
		return zero, err
	}
	return ParseResult[T](resp)
}

// DoPutJson sends a JSON body PUT request and parses the result.
func DoPutJson[T any](p *MaintainerHttpProxy, api string, params map[string]string,
	body interface{}, resource security.RequestResource) (T, error) {
	var zero T
	resp, err := p.ReqApiWithJsonBody(api, params, body, http.MethodPut, resource)
	if err != nil {
		return zero, err
	}
	return ParseResult[T](resp)
}

// DoDelete sends a DELETE request and parses the result.
func DoDelete[T any](p *MaintainerHttpProxy, api string, params map[string]string,
	resource security.RequestResource) (T, error) {
	var zero T
	resp, err := p.ReqApi(api, params, http.MethodDelete, resource)
	if err != nil {
		return zero, err
	}
	return ParseResult[T](resp)
}

// BuildResource is a helper to build a RequestResource for admin API calls.
func BuildResource(namespace, group, resource string) security.RequestResource {
	return security.BuildAdminResource(namespace, group, resource)
}

// BuildEmptyResource builds an empty RequestResource for admin API calls
// that don't require specific resource authorization.
func BuildEmptyResource() security.RequestResource {
	return security.BuildAdminResource("", "", "")
}

// TransformParam converts a struct to map[string]string using param tags.
func TransformParam(obj interface{}) map[string]string {
	return util.TransformObject2Param(obj)
}
