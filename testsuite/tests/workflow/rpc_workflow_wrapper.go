// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package workflow

import (
	"time"

	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"

	caminoNetwork "github.com/chain4travel/camino-testing/camino/networks"
	caminoService "github.com/chain4travel/camino-testing/camino/services"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	regularNodeServiceID   networks.ServiceID = "validator-node"
	delegatorNodeServiceID networks.ServiceID = "delegator-node"

	networkAcceptanceTimeoutRatio                          = 0.3
	normalNodeConfigID            networks.ConfigurationID = "normal-config"
)

// StakingNetworkRPCWorkflowTest ...
type StakingNetworkRPCWorkflowTest struct {
	ImageName string
}

// Run implements the Kurtosis Test interface
func (test StakingNetworkRPCWorkflowTest) Run(network networks.Network, context testsuite.TestContext) {
	// =============================== SETUP CAMINO CLIENTS ======================================
	castedNetwork := network.(caminoNetwork.TestCaminoNetwork)
	networkAcceptanceTimeout := time.Duration(networkAcceptanceTimeoutRatio * float64(test.GetExecutionTimeout().Nanoseconds()))
	stakerClient, err := castedNetwork.GetCaminoClient(regularNodeServiceID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get staker client"))
	}

	delegatorClient, err := castedNetwork.GetCaminoClient(delegatorNodeServiceID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get delegator client"))
	}

	executor := NewRPCWorkflowTestExecutor(stakerClient, delegatorClient, networkAcceptanceTimeout)

	logrus.Infof("Set up RPCWorkFlowTest. Executing...")
	if err := executor.ExecuteTest(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "RPCWorkflow Test failed."))
	}
}

// GetNetworkLoader implements the Kurtosis Test interface
func (test StakingNetworkRPCWorkflowTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	// Define possible service configurations.
	serviceConfigs := map[networks.ConfigurationID]caminoNetwork.TestCaminoNetworkServiceConfig{
		normalNodeConfigID: *caminoNetwork.NewTestCaminoNetworkServiceConfig(
			true,
			caminoService.DEBUG,
			test.ImageName,
			2,
			2,
			2*time.Second,
			make(map[string]string),
		),
	}
	// Define which services use which configurations.
	desiredServices := map[networks.ServiceID]networks.ConfigurationID{
		regularNodeServiceID:   normalNodeConfigID,
		delegatorNodeServiceID: normalNodeConfigID,
	}
	// Return an Camino Test Network with this service:configuration mapping.
	return caminoNetwork.NewTestCaminoNetworkLoader(
		true,
		test.ImageName,
		caminoService.DEBUG,
		2,
		2,
		0,
		2*time.Second,
		serviceConfigs,
		desiredServices,
	)
}

// GetExecutionTimeout implements the Kurtosis Test interface
func (test StakingNetworkRPCWorkflowTest) GetExecutionTimeout() time.Duration {
	return 5 * time.Minute
}

// GetSetupBuffer implements the Kurtosis Test interface
func (test StakingNetworkRPCWorkflowTest) GetSetupBuffer() time.Duration {
	// TODO drop this down when the availability checker doesn't have a sleep (becuase we spin up a bunch of nodes before the test starts executing)
	return 6 * time.Minute
}
