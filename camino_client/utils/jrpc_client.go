// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package utils

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	rpc "github.com/gorilla/rpc/v2/json2"
)

// ============= RPC Requester ===================
const (
	defaultMinID = 1
	defaultMaxID = 1000000
)

// CaminoRPCRequester ...
type CaminoRPCRequester interface {
	SendJSONRPCRequest(endpoint string, method string, params interface{}, reply interface{}) error
}

type jsonRPCRequester struct {
	uri    string
	client http.Client
}

// NewCaminoRPCRequester ...
func NewCaminoRPCRequester(uri string, requestTimeout time.Duration) CaminoRPCRequester {
	return &jsonRPCRequester{
		uri: uri,
		client: http.Client{
			Timeout: requestTimeout,
		},
	}
}

// SendJSONRPCRequest ...
func (requester jsonRPCRequester) SendJSONRPCRequest(endpoint string, method string, params interface{}, reply interface{}) error {
	// Golang has a nasty & subtle behaviour where duplicated '//' in the URL is treated as GET, even if it's POST
	// https://stackoverflow.com/questions/23463601/why-golang-treats-my-post-request-as-a-get-one
	endpoint = strings.TrimLeft(endpoint, "/")

	requestBodyBytes, err := rpc.EncodeClientRequest(method, params)
	if err != nil {
		return fmt.Errorf("problem marshaling request to endpoint '%v' with method '%v' and params '%v': %w", endpoint, method, params, err)
	}

	url := fmt.Sprintf("%v/%v", requester.uri, endpoint)
	logrus.Tracef("Sending request to %s:\n%s\n", url, requestBodyBytes)
	resp, err := requester.client.Post(url, "application/json", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return fmt.Errorf("problem while making JSON RPC POST request to %s: %s", url, err)
	}
	defer resp.Body.Close()
	statusCode := resp.StatusCode

	// Return an error for any non successful status code
	if statusCode < 200 || statusCode > 299 {
		return fmt.Errorf("received status code '%v'", statusCode)
	}

	return rpc.DecodeClientResponse(resp.Body, reply)
}

// EndpointRequester ...
type EndpointRequester interface {
	SendRequest(method string, params interface{}, reply interface{}) error
}

type caminoEndpointRequester struct {
	requester      CaminoRPCRequester
	endpoint, base string
}

// NewEndpointRequester ...
func NewEndpointRequester(uri, endpoint, base string, requestTimeout time.Duration) EndpointRequester {
	return &caminoEndpointRequester{
		requester: NewCaminoRPCRequester(uri, requestTimeout),
		endpoint:  endpoint,
		base:      base,
	}
}

func (e *caminoEndpointRequester) SendRequest(method string, params interface{}, reply interface{}) error {
	return e.requester.SendJSONRPCRequest(
		e.endpoint,
		fmt.Sprintf("%s.%s", e.base, method),
		params,
		reply,
	)
}
