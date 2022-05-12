// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package services

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/kurtosis-tech/kurtosis-go/lib/services"

	"github.com/chain4travel/camino-testing/camino/services/certs"
	"github.com/stretchr/testify/assert"
)

const (
	ipPlaceholder = "IP_PLACEHOLDER"
)

func TestNoDepsStartCommand(t *testing.T) {
	initializerCore := NewCaminoServiceInitializerCore(
		1,
		1,
		0,
		false,
		2*time.Second,
		make(map[string]string),
		[]string{},
		certs.NewStaticCaminoCertProvider(bytes.Buffer{}, bytes.Buffer{}),
		INFO,
	)

	expected := []string{
		caminogoBinary,
		"--public-ip=" + ipPlaceholder,
		"--network-id=local",
		"--http-port=9650",
		"--http-host=",
		"--staking-port=9651",
		"--log-level=info",
		"--snow-sample-size=1",
		"--snow-quorum-size=1",
		"--staking-enabled=false",
		"--tx-fee=0",
		fmt.Sprintf("--network-initial-timeout=%d", int64(2*time.Second)),
	}
	actual, err := initializerCore.GetStartCommand(make(map[string]string), ipPlaceholder, make([]services.Service, 0))
	assert.NoError(t, err, "An error occurred getting the start command")
	assert.Equal(t, expected, actual)
}

func TestWithDepsStartCommand(t *testing.T) {
	testNodeID := "node1"
	testDependencyIP := "1.2.3.4"

	bootstrapperNodeIDs := []string{
		testNodeID,
	}
	initializerCore := NewCaminoServiceInitializerCore(
		1,
		1,
		0,
		false,
		2*time.Second,
		make(map[string]string),
		bootstrapperNodeIDs,
		certs.NewStaticCaminoCertProvider(bytes.Buffer{}, bytes.Buffer{}),
		INFO,
	)

	expected := []string{
		caminogoBinary,
		"--public-ip=" + ipPlaceholder,
		"--network-id=local",
		"--http-port=9650",
		"--http-host=",
		"--staking-port=9651",
		"--log-level=info",
		"--snow-sample-size=1",
		"--snow-quorum-size=1",
		"--staking-enabled=false",
		"--tx-fee=0",
		fmt.Sprintf("--network-initial-timeout=%d", int64(2*time.Second)),
		fmt.Sprintf("--bootstrap-ips=%v:9651", testDependencyIP),
	}

	testDependency := CaminoService{
		ipAddr:      "1.2.3.4",
		jsonRPCPort: 9650,
		stakingPort: 9651,
	}
	testDependencySlice := []services.Service{
		testDependency,
	}
	actual, err := initializerCore.GetStartCommand(make(map[string]string), ipPlaceholder, testDependencySlice)
	assert.NoError(t, err, "An error occurred getting the start command")
	assert.Equal(t, expected, actual)
}
