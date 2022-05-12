// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package connected

import (
	"context"
	"time"

	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"

	caminoNetwork "github.com/chain4travel/camino-testing/camino/networks"
	caminoService "github.com/chain4travel/camino-testing/camino/services"
	"github.com/chain4travel/camino-testing/camino_client/apis"
	"github.com/chain4travel/camino-testing/testsuite/helpers"
	"github.com/chain4travel/camino-testing/testsuite/verifier"
	"github.com/chain4travel/caminogo/api"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	stakerUsername = "staker"
	stakerPassword = "test34test!23"
	seedAmount     = uint64(50000000000000)
	stakeAmount    = uint64(30000000000000)

	normalNodeConfigID networks.ConfigurationID = "normal-config"

	networkAcceptanceTimeoutRatio                    = 0.3
	nonBootValidatorServiceID     networks.ServiceID = "validator-service"
	nonBootNonValidatorServiceID  networks.ServiceID = "non-validator-service"
)

// StakingNetworkFullyConnectedTest adds nodes to the network and verifies that the network stays fully connected
type StakingNetworkFullyConnectedTest struct {
	ImageName           string
	FullyConnectedDelay time.Duration
	Verifier            verifier.NetworkStateVerifier
}

// Run implements the Kurtosis Test interface
func (test StakingNetworkFullyConnectedTest) Run(network networks.Network, context testsuite.TestContext) {
	castedNetwork := network.(caminoNetwork.TestCaminoNetwork)
	networkAcceptanceTimeout := time.Duration(networkAcceptanceTimeoutRatio * float64(test.GetExecutionTimeout().Nanoseconds()))

	stakerIDs := castedNetwork.GetAllBootServiceIDs()
	allServiceIDs := make(map[networks.ServiceID]bool)
	for stakerID := range stakerIDs {
		allServiceIDs[stakerID] = true
	}
	// Add our custom nodes
	allServiceIDs[nonBootValidatorServiceID] = true
	allServiceIDs[nonBootNonValidatorServiceID] = true

	allNodeIDs, allCaminoClients := getNodeIDsAndClients(test.Verifier.Ctx(), context, castedNetwork, allServiceIDs)
	logrus.Infof("Verifying that the network is fully connected...")
	if err := test.Verifier.VerifyNetworkFullyConnected(allServiceIDs, stakerIDs, allNodeIDs, allCaminoClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
	}
	logrus.Infof("Network is fully connected.")

	logrus.Infof("Adding additional staker to the network...")
	nonBootValidatorClient := allCaminoClients[nonBootValidatorServiceID]
	highLevelExtraStakerClient := helpers.NewRPCWorkFlowRunner(
		nonBootValidatorClient,
		api.UserPass{Username: stakerUsername, Password: stakerPassword},
		networkAcceptanceTimeout)
	if _, err := highLevelExtraStakerClient.ImportGenesisFundsAndStartValidating(seedAmount, stakeAmount); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to add extra staker."))
	}

	logrus.Infof("Sleeping %v seconds before verifying that the network has fully connected to the new staker...", test.FullyConnectedDelay.Seconds())
	// Give time for the new validator to propagate via gossip
	time.Sleep(70 * time.Second)

	logrus.Infof("Verifying that the network is fully connected...")
	stakerIDs[nonBootValidatorServiceID] = true
	/*
		After gossip, we expect the peers list to look like:
		1) No node has itself in its peers list
		2) The validators will have ALL other nodes in the network (propagated via gossip)
		3) The non-validators will have all the validators in the network (propagated via gossip)
	*/
	if err := test.Verifier.VerifyNetworkFullyConnected(allServiceIDs, stakerIDs, allNodeIDs, allCaminoClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying that the network is fully connected after gossip"))
	}
	logrus.Infof("The network is fully connected.")
}

// GetNetworkLoader implements the Kurtosis Test interface
func (test StakingNetworkFullyConnectedTest) GetNetworkLoader() (networks.NetworkLoader, error) {
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
	desiredServices := map[networks.ServiceID]networks.ConfigurationID{
		nonBootValidatorServiceID:    normalNodeConfigID,
		nonBootNonValidatorServiceID: normalNodeConfigID,
	}
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
func (test StakingNetworkFullyConnectedTest) GetExecutionTimeout() time.Duration {
	return 5 * time.Minute
}

// GetSetupBuffer implements the Kurtosis Test interface
func (test StakingNetworkFullyConnectedTest) GetSetupBuffer() time.Duration {
	// TODO drop this when the availabilityChecker doesn't have a sleep (because we spin up a bunch of nodes before running the test)
	return 6 * time.Minute
}

// ================ Helper functions =========================
/*
This helper function will grab node IDs and Camino clients
*/
func getNodeIDsAndClients(
	ctx context.Context,
	testContext testsuite.TestContext,
	network caminoNetwork.TestCaminoNetwork,
	allServiceIDs map[networks.ServiceID]bool,
) (allNodeIDs map[networks.ServiceID]string, allCaminoClients map[networks.ServiceID]*apis.Client) {
	allCaminoClients = make(map[networks.ServiceID]*apis.Client)
	allNodeIDs = make(map[networks.ServiceID]string)
	for serviceID := range allServiceIDs {
		client, err := network.GetCaminoClient(serviceID)
		if err != nil {
			testContext.Fatal(stacktrace.Propagate(err, "An error occurred getting the Camino client for service with ID %v", serviceID))
		}
		allCaminoClients[serviceID] = client
		nodeID, err := client.InfoAPI().GetNodeID(ctx)

		if err != nil {
			testContext.Fatal(stacktrace.Propagate(err, "An error occurred getting the Camino node ID for service with ID %v", serviceID))
		}
		allNodeIDs[serviceID] = nodeID
	}
	return
}
