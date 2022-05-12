// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package spamchits

import (
	"context"
	"strconv"
	"time"

	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"

	caminoNetwork "github.com/chain4travel/camino-testing/camino/networks"
	caminoService "github.com/chain4travel/camino-testing/camino/services"
	"github.com/chain4travel/camino-testing/testsuite/helpers"
	"github.com/chain4travel/caminogo/api"
	"github.com/chain4travel/caminogo/ids"
	"github.com/chain4travel/caminogo/vms/platformvm"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	normalNodeConfigID     networks.ConfigurationID = "normal-config"
	byzantineConfigID      networks.ConfigurationID = "byzantine-config"
	byzantineUsername                               = "byzantine_camino"
	byzantinePassword                               = "byzant1n3!"
	stakerUsername                                  = "staker_camino"
	stakerPassword                                  = "test34test!23"
	normalNodeServiceID    networks.ServiceID       = "normal-node"
	byzantineNodePrefix    string                   = "byzantine-node-"
	numberOfByzantineNodes                          = 4
	seedAmount                                      = uint64(50000000000000)
	stakeAmount                                     = uint64(30000000000000)

	networkAcceptanceTimeoutRatio = 0.3
	byzantineBehavior             = "byzantine-behavior"
	chitSpammerBehavior           = "chit-spammer"
)

// StakingNetworkUnrequestedChitSpammerTest tests that a node is able to continue to work normally
// while the network is spammed with chit messages from byzantine peers
type StakingNetworkUnrequestedChitSpammerTest struct {
	ctx                context.Context
	ByzantineImageName string
	NormalImageName    string
}

func NewStakingNetworkUnrequestedChitSpammerTest(
	byzantineImageName string,
	normalImageName string,
) StakingNetworkUnrequestedChitSpammerTest {
	return StakingNetworkUnrequestedChitSpammerTest{
		ctx:                context.Background(),
		ByzantineImageName: byzantineImageName,
		NormalImageName:    normalImageName,
	}
}

// Run implements the Kurtosis Test interface
func (test StakingNetworkUnrequestedChitSpammerTest) Run(network networks.Network, context testsuite.TestContext) {
	castedNetwork := network.(caminoNetwork.TestCaminoNetwork)
	networkAcceptanceTimeout := time.Duration(networkAcceptanceTimeoutRatio * float64(test.GetExecutionTimeout().Nanoseconds()))

	// ============= ADD SET OF BYZANTINE NODES AS VALIDATORS ON THE NETWORK ===================
	logrus.Infof("Adding byzantine chit spammer nodes as validators...")
	for i := 0; i < numberOfByzantineNodes; i++ {
		byzClient, err := castedNetwork.GetCaminoClient(networks.ServiceID(byzantineNodePrefix + strconv.Itoa(i)))
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to get byzantine client."))
		}
		highLevelByzClient := helpers.NewRPCWorkFlowRunner(
			byzClient,
			api.UserPass{Username: byzantineUsername, Password: byzantinePassword},
			networkAcceptanceTimeout)
		_, err = highLevelByzClient.ImportGenesisFundsAndStartValidating(seedAmount, stakeAmount)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed add client as a validator."))
		}
		currentValidators, err := byzClient.PChainAPI().GetCurrentValidators(test.ctx, ids.Empty, []ids.ShortID{})
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Could not get current validators."))
		}
		currentNumDelegators := 0
		for _, iValidator := range currentValidators {
			if validator, ok := iValidator.(platformvm.APIPrimaryValidator); !ok {
				context.Fatal(stacktrace.Propagate(err, "Could not convert validator."))
			} else {
				currentNumDelegators += len(validator.Delegators)
			}
		}
		logrus.Infof("Current Validators: %d, Delegators: %d", len(currentValidators), currentNumDelegators)
	}

	// =================== ADD NORMAL NODE AS A VALIDATOR ON THE NETWORK =======================
	logrus.Infof("Adding normal node as a staker...")
	availabilityChecker, err := castedNetwork.AddService(normalNodeConfigID, normalNodeServiceID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to add normal node with high quorum and sample to network."))
	}
	if err = availabilityChecker.WaitForStartup(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to wait for startup of normal node."))
	}
	normalClient, err := castedNetwork.GetCaminoClient(normalNodeServiceID)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to get staker client."))
	}
	highLevelNormalClient := helpers.NewRPCWorkFlowRunner(
		normalClient,
		api.UserPass{Username: stakerUsername, Password: stakerPassword},
		networkAcceptanceTimeout)
	_, err = highLevelNormalClient.ImportGenesisFundsAndStartValidating(seedAmount, stakeAmount)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to add client as a validator."))
	}

	logrus.Infof("Added normal node as a staker. Sleeping an additional 10 seconds to ensure it joins current validators...")
	time.Sleep(10 * time.Second)

	// ============= VALIDATE NETWORK STATE DESPITE BYZANTINE BEHAVIOR =========================
	logrus.Infof("Validating network state...")
	actualValidators, err := normalClient.PChainAPI().GetCurrentValidators(test.ctx, ids.Empty, []ids.ShortID{})
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get current validators."))
	}
	actualNumValidators := len(actualValidators)
	expectedNumvalidators := 10
	logrus.Debugf("Number of current validators: %d, expected number of validators: %d", actualNumValidators, expectedNumvalidators)
	if actualNumValidators != expectedNumvalidators {
		context.AssertTrue(actualNumValidators == expectedNumvalidators, stacktrace.NewError("Actual number of validators, %v, != expected number of validators, %v", actualNumValidators, expectedNumvalidators))
	}
	actualNumDelegators := 0
	for _, iValidator := range actualValidators {
		if validator, ok := iValidator.(platformvm.APIPrimaryValidator); !ok {
			context.Fatal(stacktrace.Propagate(err, "Could not convert validator."))
		} else {
			actualNumDelegators += len(validator.Delegators)
		}
	}
	expectedNumDelegators := 0
	logrus.Debugf("Number of current delegators: %d, expected number of delegators: %d", actualNumDelegators, expectedNumDelegators)
	if actualNumDelegators != expectedNumDelegators {
		context.AssertTrue(actualNumDelegators == expectedNumDelegators, stacktrace.NewError("Actual number of delegators, %v, != expected number of delegators, %v", actualNumDelegators, expectedNumDelegators))
	}
}

// GetNetworkLoader implements the Kurtosis Test interface
func (test StakingNetworkUnrequestedChitSpammerTest) GetNetworkLoader() (networks.NetworkLoader, error) {
	// Define normal node and byzantine node configurations
	serviceConfigs := map[networks.ConfigurationID]caminoNetwork.TestCaminoNetworkServiceConfig{
		byzantineConfigID: *caminoNetwork.NewTestCaminoNetworkServiceConfig(
			true,
			caminoService.DEBUG,
			test.ByzantineImageName,
			2,
			2,
			2*time.Second,
			map[string]string{
				byzantineBehavior: chitSpammerBehavior,
			},
		),
		normalNodeConfigID: *caminoNetwork.NewTestCaminoNetworkServiceConfig(
			true,
			caminoService.DEBUG,
			test.NormalImageName,
			6,
			8,
			2*time.Second,
			make(map[string]string),
		),
	}

	// Define the map from service->configuration for the network
	serviceIDConfigMap := map[networks.ServiceID]networks.ConfigurationID{}
	for i := 0; i < numberOfByzantineNodes; i++ {
		serviceIDConfigMap[networks.ServiceID(byzantineNodePrefix+strconv.Itoa(i))] = byzantineConfigID
	}
	logrus.Debugf("Byzantine Image Name: %s", test.ByzantineImageName)
	logrus.Debugf("Normal Image Name: %s", test.NormalImageName)

	return caminoNetwork.NewTestCaminoNetworkLoader(
		true,
		test.NormalImageName,
		caminoService.DEBUG,
		2,
		2,
		0,
		2*time.Second,
		serviceConfigs,
		serviceIDConfigMap,
	)
}

// GetExecutionTimeout implements the Kurtosis Test interface
func (test StakingNetworkUnrequestedChitSpammerTest) GetExecutionTimeout() time.Duration {
	// TODO drop this when the availabilityChecker doesn't have a sleep, because we spin up a *bunch* of byzantine
	// nodes during test execution
	return 10 * time.Minute
}

// GetSetupBuffer implements the Kurtosis Test interface
func (test StakingNetworkUnrequestedChitSpammerTest) GetSetupBuffer() time.Duration {
	return 4 * time.Minute
}
