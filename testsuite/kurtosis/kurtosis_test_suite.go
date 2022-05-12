// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package kurtosis

import (
	"time"

	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"

	"github.com/chain4travel/camino-testing/testsuite/tests/bombard"
	"github.com/chain4travel/camino-testing/testsuite/tests/conflictvtx"
	"github.com/chain4travel/camino-testing/testsuite/tests/connected"
	"github.com/chain4travel/camino-testing/testsuite/tests/duplicate"
	"github.com/chain4travel/camino-testing/testsuite/tests/spamchits"
	"github.com/chain4travel/camino-testing/testsuite/tests/workflow"
	"github.com/chain4travel/camino-testing/testsuite/verifier"
)

const (
	// The number of bits to make each test network, which dictates the max number of services a test can spin up
	// Here we choose 8 bits = 256 max services per test
	networkWidthBits = 8
)

// CaminoTestSuite implements the Kurtosis TestSuite interface
type CaminoTestSuite struct {
	ByzantineImageName string
	NormalImageName    string
}

// GetTests implements the Kurtosis TestSuite interface
func (a CaminoTestSuite) GetTests() map[string]testsuite.Test {
	result := make(map[string]testsuite.Test)

	if a.ByzantineImageName != "" {
		result["stakingNetworkChitSpammerTest"] = spamchits.NewStakingNetworkUnrequestedChitSpammerTest(
			a.ByzantineImageName,
			a.NormalImageName,
		)
		result["conflictingTxsVertexTest"] = conflictvtx.StakingNetworkConflictingTxsVertexTest{
			ByzantineImageName: a.ByzantineImageName,
			NormalImageName:    a.NormalImageName,
		}
	}
	result["stakingNetworkBombardXChainTest"] = bombard.StakingNetworkBombardTest{
		ImageName:         a.NormalImageName,
		NumTxs:            1000,
		TxFee:             1000000,
		AcceptanceTimeout: 10 * time.Second,
	}
	result["stakingNetworkFullyConnectedTest"] = connected.StakingNetworkFullyConnectedTest{
		ImageName: a.NormalImageName,
		Verifier:  verifier.NewNetworkStateVerifier(),
	}
	result["stakingNetworkDuplicateNodeIDTest"] = duplicate.DuplicateNodeIDTest{
		ImageName: a.NormalImageName,
		Verifier:  verifier.NewNetworkStateVerifier(),
	}
	result["StakingNetworkRPCWorkflowTest"] = workflow.StakingNetworkRPCWorkflowTest{
		ImageName: a.NormalImageName,
	}

	return result
}

func (a CaminoTestSuite) GetNetworkWidthBits() uint32 {
	return networkWidthBits
}
