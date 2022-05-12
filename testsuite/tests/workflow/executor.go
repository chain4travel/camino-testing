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
	"context"
	"time"

	"github.com/chain4travel/camino-testing/camino_client/apis"
	"github.com/chain4travel/camino-testing/testsuite/helpers"
	"github.com/chain4travel/camino-testing/testsuite/tester"
	"github.com/chain4travel/caminogo/api"
	"github.com/chain4travel/caminogo/ids"
	"github.com/chain4travel/caminogo/utils/constants"
	"github.com/chain4travel/caminogo/utils/units"
	"github.com/chain4travel/caminogo/vms/platformvm"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	genesisUsername   = "genesis"
	genesisPassword   = "MyNameIs!Jeff"
	stakerUsername    = "staker"
	stakerPassword    = "test34test!23"
	delegatorUsername = "delegator"
	delegatorPassword = "test34test!23"
	seedAmount        = 5 * units.KiloAvax
	stakeAmount       = 3 * units.KiloAvax
	delegatorAmount   = 3 * units.KiloAvax
)

type executor struct {
	ctx                           context.Context
	stakerClient, delegatorClient *apis.Client
	acceptanceTimeout             time.Duration
}

// NewRPCWorkflowTestExecutor ...
func NewRPCWorkflowTestExecutor(stakerClient, delegatorClient *apis.Client, acceptanceTimeout time.Duration) tester.CaminoTester {
	return &executor{
		stakerClient:      stakerClient,
		delegatorClient:   delegatorClient,
		acceptanceTimeout: acceptanceTimeout,
	}
}

// ExecuteTest ...
func (e *executor) ExecuteTest() error {
	genesisClient := helpers.NewRPCWorkFlowRunner(
		e.stakerClient,
		api.UserPass{Username: genesisUsername, Password: genesisPassword},
		e.acceptanceTimeout,
	)

	if _, err := genesisClient.ImportGenesisFunds(); err != nil {
		return stacktrace.Propagate(err, "Failed to fund genesis client.")
	}
	logrus.Debugf("Funded genesis client...")

	stakerNodeID, err := e.stakerClient.InfoAPI().GetNodeID(e.ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Could not get staker node ID.")
	}
	delegatorNodeID, err := e.delegatorClient.InfoAPI().GetNodeID(e.ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Could not get delegator node ID.")
	}
	highLevelStakerClient := helpers.NewRPCWorkFlowRunner(
		e.stakerClient,
		api.UserPass{Username: stakerUsername, Password: stakerPassword},
		e.acceptanceTimeout,
	)
	highLevelDelegatorClient := helpers.NewRPCWorkFlowRunner(
		e.delegatorClient,
		api.UserPass{Username: delegatorUsername, Password: delegatorPassword},
		e.acceptanceTimeout,
	)

	// ====================================== CREATE FUNDED ACCOUNTS ===============================
	stakerXChainAddress, stakerPChainAddress, err := highLevelStakerClient.CreateDefaultAddresses()
	if err != nil {
		return stacktrace.Propagate(err, "Could not create default addresses for staker client.")
	}
	delegatorXChainAddress, delegatorPChainAddress, err := highLevelDelegatorClient.CreateDefaultAddresses()
	if err != nil {
		return stacktrace.Propagate(err, "Could not create default addresses for delegator client.")
	}
	logrus.Infof("Created addresses for staker and delegator clients.")

	if err := genesisClient.FundXChainAddresses([]string{stakerXChainAddress, delegatorXChainAddress}, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Failed to fund X Chain Addresses from genesis client.")
	}

	if err := highLevelStakerClient.VerifyXChainAVABalance(stakerXChainAddress, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Unexpected X Chain balance for staker client.")
	}
	if err := highLevelDelegatorClient.VerifyXChainAVABalance(delegatorXChainAddress, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Unexpected X Chain Balance for delegator client.")
	}
	logrus.Infof("Funded X Chain Addresses for staker and delegator clients.")

	//  ====================================== ADD VALIDATOR ===============================
	err = highLevelStakerClient.TransferAvaXChainToPChain(stakerPChainAddress, seedAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not transfer AVAX from XChain to PChain account information")
	}
	if err := highLevelStakerClient.VerifyPChainBalance(stakerPChainAddress, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain balance after X -> P Transfer.")
	}
	if err := highLevelStakerClient.VerifyXChainAVABalance(stakerXChainAddress, 0); err != nil {
		return stacktrace.Propagate(err, "X Chain Balance not updated correctly after X -> P Transfer for validator")
	}
	err = highLevelStakerClient.AddValidatorToPrimaryNetwork(stakerNodeID, stakerPChainAddress, stakeAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not add staker %s to primary network.", stakerNodeID)
	}
	logrus.Infof("Transferred funds from X Chain to P Chain and added a new staker.")

	// ====================================== VERIFY NETWORK STATE ===============================
	currentValidators, err := e.stakerClient.PChainAPI().GetCurrentValidators(
		e.ctx, constants.PrimaryNetworkID, []ids.ShortID{})
	if err != nil {
		return stacktrace.Propagate(err, "Could not get current validators.")
	}
	actualNumValidators := len(currentValidators)
	logrus.Debugf("Number of current stakers: %d", actualNumValidators)
	expectedNumValidators := 6
	if actualNumValidators != expectedNumValidators {
		return stacktrace.NewError("Actual number of validators, %v, != expected number of validators, %v", actualNumValidators, expectedNumValidators)
	}
	actualNumDelegators := 0
	for _, iValidator := range currentValidators {
		if validator, ok := iValidator.(platformvm.APIPrimaryValidator); !ok {
			return stacktrace.Propagate(nil, "Could not convert validator.")
		} else {
			actualNumDelegators += len(validator.Delegators)
		}
	}

	logrus.Debugf("Number of current delegators: %d", actualNumDelegators)
	expectedNumDelegators := 0
	if actualNumDelegators != expectedNumDelegators {
		return stacktrace.NewError("Actual number of delegators, %v, != expected number of delegators, %v", actualNumDelegators, expectedNumDelegators)
	}
	expectedStakerBalance := seedAmount - stakeAmount
	if err := highLevelStakerClient.VerifyPChainBalance(stakerPChainAddress, expectedStakerBalance); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain Balance after adding  validator to the primary network")
	}
	logrus.Infof("Verified the staker was added to current validators and has the expected P Chain balance.")

	// ====================================== ADD DELEGATOR ======================================
	err = highLevelDelegatorClient.TransferAvaXChainToPChain(delegatorPChainAddress, seedAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not transfer AVAX from X Chain to P Chain account.")
	}
	if err := highLevelDelegatorClient.VerifyPChainBalance(delegatorPChainAddress, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain balance after X -> P Transfer for Delegator.")
	}
	if err := highLevelDelegatorClient.VerifyXChainAVABalance(delegatorXChainAddress, 0); err != nil {
		return stacktrace.Propagate(err, "Unexpected X Chain Balance after X -> P Transfer for Delegator")
	}

	err = highLevelDelegatorClient.AddDelegatorToPrimaryNetwork(stakerNodeID, delegatorPChainAddress, delegatorAmount)
	if err != nil {
		return stacktrace.Propagate(err, "Could not add delegator %s to the primary network.", delegatorNodeID)
	}
	expectedDelegatorBalance := seedAmount - delegatorAmount
	if err := highLevelDelegatorClient.VerifyPChainBalance(delegatorPChainAddress, expectedDelegatorBalance); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain Balance after adding a new delegator to the network.")
	}
	logrus.Infof("Added delegator to subnet and verified the expected P Chain balance.")

	// ====================================== TRANSFER TO X CHAIN ================================
	err = highLevelStakerClient.TransferAvaPChainToXChain(stakerXChainAddress, expectedStakerBalance)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to transfer AvaX from P Chain to X Chain.")
	}
	if err := highLevelStakerClient.VerifyPChainBalance(stakerPChainAddress, 0); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain Balance after P -> X Transfer.")
	}
	if err := highLevelStakerClient.VerifyXChainAVABalance(stakerXChainAddress, expectedStakerBalance); err != nil {
		return stacktrace.Propagate(err, "Unexpected X Chain Balance after P -> X Transfer.")
	}
	logrus.Infof("Transferred leftover staker funds back to X Chain and verified X and P balances.")

	err = highLevelDelegatorClient.TransferAvaPChainToXChain(delegatorXChainAddress, expectedStakerBalance)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to transfer AVAX from P Chain to X Chain.")
	}
	if err := highLevelDelegatorClient.VerifyPChainBalance(delegatorPChainAddress, 0); err != nil {
		return stacktrace.Propagate(err, "Unexpected P Chain Balance after P -> X Transfer.")
	}
	if err := highLevelDelegatorClient.VerifyXChainAVABalance(delegatorXChainAddress, expectedDelegatorBalance); err != nil {
		return stacktrace.Propagate(err, "Unexpected X Chain Balance after P -> X Transfer.")
	}
	logrus.Infof("Transferred leftover delegator funds back to X Chain and verified X and P balances.")

	return nil
}
