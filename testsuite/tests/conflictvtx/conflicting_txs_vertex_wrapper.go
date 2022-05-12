// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package conflictvtx

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
	normalNodeConfigID          networks.ConfigurationID = "normal-config"
	byzantineConfigID           networks.ConfigurationID = "byzantine-config"
	byzantineUsername                                    = "byzantine_camino"
	byzantinePassword                                    = "byzant1n3!"
	byzantineBehavior                                    = "byzantine-behavior"
	conflictingTxVertexBehavior                          = "conflicting-txs-vertex"
	stakerUsername                                       = "staker_camino"
	stakerPassword                                       = "test34test!23"
	byzantineNodeServiceID                               = "byzantine-node"
	normalNodeServiceID                                  = "virtuous-node"
	seedAmount                                           = int64(50000000000000)
	stakeAmount                                          = int64(30000000000000)
)

// StakingNetworkConflictingTxsVertexTest creates a byzantine node to issue conflicting transactions into a single
// vertex. It then checks to ensure that the byzantine node has accepted these transactions, while the virtuous nodes
// drop the vertex.
type StakingNetworkConflictingTxsVertexTest struct {
	ByzantineImageName string
	NormalImageName    string
}

// Run implements the Kurtosis Test interface
func (test StakingNetworkConflictingTxsVertexTest) Run(network networks.Network, context testsuite.TestContext) {
	castedNetwork := network.(caminoNetwork.TestCaminoNetwork)

	byzantineClient, err := castedNetwork.GetCaminoClient(byzantineNodeServiceID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to get byzantine client."))
	}
	virtuousClient, err := castedNetwork.GetCaminoClient(normalNodeServiceID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to get virtuous client."))
	}
	executor := NewConflictingTxsVertexExecutor(virtuousClient, byzantineClient)
	logrus.Infof("Executing conflicting transaction vertex test...")
	if err := executor.ExecuteTest(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Conflicting Transactions Vertex Test failed."))
	}
}

// GetNetworkLoader implements the Kurtosis Test interface
func (test StakingNetworkConflictingTxsVertexTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	// Provision a byzantine and normal node
	desiredServices := map[networks.ServiceID]networks.ConfigurationID{}
	desiredServices[byzantineNodeServiceID] = byzantineConfigID
	desiredServices[normalNodeServiceID] = normalNodeConfigID

	return getByzantineNetworkLoader(desiredServices, test.ByzantineImageName, test.NormalImageName)
}

// GetExecutionTimeout implements the Kurtosis Test interface
func (test StakingNetworkConflictingTxsVertexTest) GetExecutionTimeout() time.Duration {
	return 2 * time.Minute
}

// GetSetupBuffer implements the Kurtosis Test interface
func (test StakingNetworkConflictingTxsVertexTest) GetSetupBuffer() time.Duration {
	return 2 * time.Minute
}

// =============== Helper functions =============================

/*
Args:
	desiredServices: Mapping of service_id -> configuration_id for all services *in addition to the boot nodes* that the user wants
*/
func getByzantineNetworkLoader(desiredServices map[networks.ServiceID]networks.ConfigurationID, byzantineImageName string, normalImageName string) (networks.NetworkLoader, error) {
	serviceConfigs := map[networks.ConfigurationID]caminoNetwork.TestCaminoNetworkServiceConfig{
		normalNodeConfigID: *caminoNetwork.NewTestCaminoNetworkServiceConfig(
			true,
			caminoService.DEBUG,
			normalImageName,
			2,
			2,
			2*time.Second,
			make(map[string]string),
		),
		byzantineConfigID: *caminoNetwork.NewTestCaminoNetworkServiceConfig(
			true,
			caminoService.DEBUG,
			byzantineImageName,
			2,
			2,
			2*time.Second,
			map[string]string{byzantineBehavior: conflictingTxVertexBehavior},
		),
	}
	logrus.Debugf("Byzantine Image Name: %s", byzantineImageName)
	logrus.Debugf("Normal Image Name: %s", normalImageName)

	return caminoNetwork.NewTestCaminoNetworkLoader(
		true,
		normalImageName,
		caminoService.DEBUG,
		2,
		2,
		1000000,
		2*time.Second,
		serviceConfigs,
		desiredServices,
	)
}
