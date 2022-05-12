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
	"context"
	"fmt"
	"time"

	"github.com/chain4travel/camino-testing/camino_client/apis"
	"github.com/chain4travel/camino-testing/testsuite/tester"
	"github.com/chain4travel/caminogo/snow/choices"
	"github.com/chain4travel/caminogo/utils/formatting"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

type executor struct {
	ctx             context.Context
	virtuousClient  *apis.Client
	byzantineClient *apis.Client
}

// NewConflictingTxsVertexExecutor ...
func NewConflictingTxsVertexExecutor(virtuousClient, byzantineClient *apis.Client) tester.CaminoTester {
	return &executor{
		ctx:             context.Background(),
		virtuousClient:  virtuousClient,
		byzantineClient: byzantineClient,
	}
}

// ExecuteTest implements CaminoTester interface
func (e *executor) ExecuteTest() error {
	byzantineXChainAPI := e.byzantineClient.XChainAPI()

	// TODO switch to test vectors or come up with method to reliably generate conflicting transactions
	// how to create a conflicting transaction???
	// test vector create asset tx and conflicting transactions
	// Note: these hardcoded transactions are based on everest-deployment with a txFee of 1,000,000 nAVAX

	createAssetTx, err := formatting.Decode(formatting.CB58, "11111kH8MvvvKhHX48mhBwKNrd4iqNeWsUyXFuuXaaymx3TE5jQEGn4h2mn211eCZwckWeNyMGHMCuHFbSBUF3q6Dz3btR95g17xWe7QraryeR6zKmJi9YuH2cbAoDCVBzqHaHEDVY2mArzpbE2wLLcVRPqnrneB4K1EqjepXSzvLJumGhej6eGCvqfQSTegV9jPjQrcxzWSFDbbvP9e532NzxpP84P4Rhy7oC9R2ngMcz846xsso44YvGorT3gRNHBCVXqKvi1epJ6EsjmskGEFr9xaQm32kkUKf2KUeM9EiKzwn6DDBusAPoLAw97mdjEuoxogue5xvwo4bHDQaL8zsZ68Gm1ETLaFdc1qXwhR6Y6Pd6b4MRgJTEG4cF4dGz18RvheicgGrwQDuASY51xjM1ijeVjmwyXGoo4k248fVY3rLgRirtnGcVfsuAyxdcb4x7ZqcRDNAjhriQCEV9m7R2Bm7XLcL2FkuXNLHFbXgsLEZ8L7LiC5fsmf4")
	if err != nil {
		return stacktrace.Propagate(err, "Problem parsing create asset tx")
	}

	conflictingTx1, err := formatting.Decode(formatting.CB58, "11111111Wc1cPy8nDRbzTRYwGxcqaSC5eCqSmpK855uz6HP8FQZMXSja7nmTgfnmAT5M1PMQUEf89WfM3LtWCathAK2mfpynPtrtTagPktae4fXHL6Ucn3hpQou6A5ijtw8hJPfzcX3JZKbTPbB4qw8D3jX8k8D9avVA7nwdv9eyuMgJnBSHdLpDfnyBgytkk7uicy3NR6dHndwnWB7euhZ55yzCHRBAxSrqiirz7EBmxScw1b5baU4eRypoGgeeMkwMjx6ifuYwEBg5cQX2z3chRLBodDypTn4QLLjxXQscP9DHX9sYQmRiuxVgymyvYPj2vd99W5sSLoqjCdzVm8unp2Q4Rx9JFNRrhza8bUCh8eDxZUVV7bxG3d3bBkpaAW2DmD3ZUwbuFLZvE5XSvpapcz8tzsHnPzQa5EA2HrJLczfg2WMnger3441En1Yg9Aj5FAbEYbo6AzTZrGhx4tgXXNnfGuQLqXCQvkc5SsGEAeT73Kcg7VWUVNT5FthCRSAxBTZ3YxNnBddgfcg4Zne1TqBCAD8n6WfBQnT8MZnDvk3XV5SxmjzqMcdxvew1KUD26LaZHZvqg8SbaP26EmHACJuJ4YprSP2ESPZVrBZwmcbWRcaD6LgnmmV3u7NGTxye36i7KW3cAxUs9HCNSmLR17MQmgukZX4eQ6H2LNeCwQdkyyS5aukKLxbeF1Q6zqYFZpdaSDKVzPnmDKmPDaEFsqCDzXLMy6UXXzyp6ebHSM9mdryPPWHjSijsmKWjggN4BgiVQrKVUM8pyx4NiVDTRufyzgXWUmAZC3hm3cvujSozW5hZAfTp3Bd4o6m7YRh6fnE1QXArY8H5S")
	if err != nil {
		return stacktrace.Propagate(err, "Problem parsing first send tx")
	}

	conflictingTx2, err := formatting.Decode(formatting.CB58, "11111111Wc1cPy8nDRbzTRYwGxcqaSC5eCqSmpK855uz6HP8FQZMXSja7nmTgfnmAT5M1PMQUEf89WfM3LtWCathAK2mfpynPtrtTagPktae4fXHL6Ucn3iKKvR7GZ1jMEJZg5PAN1Kyq5wN1qAqnm6PCtomhS7bTkshbS4veE8QgztEWuzZ2bgNSvEaY8LvRG3XaetE5c9aj8676ruodp6AMFtXULZLp94naenxLhkKmNPiKzjxawn6KfPoUsfRvP9HYyFtLC7WrpuHyzXM2m9ysWv4zZym93SzMrjfs7cy5BKwufRV6uZQ5EMJCmY8y2wDmVfD3WT19bz9853jrTBTcq1BTMChYQaoBEn96Xk9ik1NQWm3Qwq3aGES97z1yRcYmXKvRKqcKivYqCrkRFXixpVYebBJv715ikm1uebLw7TSybP8GWRayAFKPE3s5dMJShu3gXBeqiXesrKygm7e2cDaASjFcg2J58eKfovETAX81b2qcDTyHNBPjsvJidxyDdYF6mMQauXwSrseLJAD14v4GnwjuMwc6gbaWfQwRrgtE11coHX2fzQbYRQuWEad5fvz9NwnwRHwqj4xqGrZkYfWWQ61avSHNo4WvnYP9uSCvvwUuU3HoBs4tUhZYe8eUuQuxGiydS717Rth2t9HKHig6wrbdARxmDhVYErordxdjjpXgMU2xPQdqvchnGgak7m6CjXq2SJh5HUaGemf5T8kh4EZry6CT4ytAMyEMRhphSjrtQmbd1AHZM74z7QezhXNwgbwgvDD2oiC6Qxzn9w7xWhRF5f87aR2y6dJQu9umX9khwoW2GprnmDdgsexHfiuSzW1GPjNV")
	if err != nil {
		return stacktrace.Propagate(err, "Problem parsing second send tx")
	}

	logrus.Infof("Issuing conflicting transactions to a byzantine node...")
	nonConflictID, err := byzantineXChainAPI.IssueTx(e.ctx, createAssetTx)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to issue first transaction to byzantine node.")
	}
	conflictID1, err := byzantineXChainAPI.IssueTx(e.ctx, conflictingTx1)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to issue second transaction to byzantine node.")
	}
	conflictID2, err := byzantineXChainAPI.IssueTx(e.ctx, conflictingTx2)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to issue third transaction to byzantine node.")
	}

	logrus.Infof("Issued three transactions to Byzantine Node with IDs: %s, %s, %s", nonConflictID, conflictID1, conflictID2)

	// Confirm the byzantine node Accepted the transactions
	// Note: The byzantine behavior is to batch the pending transactions into a vertex as soon as it detects a conflict.
	// It should try to accept each transaction before PushQuery-ing the vertex to other nodes to signal to this test
	// controller that the vertex was successfully issued
	status, err := byzantineXChainAPI.GetTxStatus(e.ctx, nonConflictID)
	if err != nil {
		return stacktrace.Propagate(err, fmt.Sprintf("Failed to get status of Transaction: %s", nonConflictID))
	}
	if status != choices.Accepted {
		return stacktrace.Propagate(err, fmt.Sprintf("Transaction: %s was not accepted, status: %s", nonConflictID, status))
	}

	logrus.Infof("Status of non-conflict transactions on byzantine node is: %s", status)

	conflictStatus1, err := byzantineXChainAPI.GetTxStatus(e.ctx, conflictID1)
	if err != nil {
		return stacktrace.Propagate(err, fmt.Sprintf("Failed to get status of Transaction: %s", conflictID1))
	}

	logrus.Infof("Status of conflict tx1: %s on byzantine node is: %s", conflictID1, conflictStatus1)

	conflictStatus2, err := byzantineXChainAPI.GetTxStatus(e.ctx, conflictID2)
	if err != nil {
		return stacktrace.Propagate(err, fmt.Sprintf("Failed to get status of Transaction: %s", conflictID2))
	}

	logrus.Infof("Status of conflict tx2: %s on byzantine node is: %s", conflictID2, conflictStatus2)

	// Byzantine node should try to accept both conflicting transactions, but will fail to accept one due to the missing UTXO
	// after the other consumes it.
	if conflictStatus1 != choices.Accepted && conflictStatus2 != choices.Accepted {
		return fmt.Errorf("Byzantine node did not accept either of the conflicting transactions, status1: %s. status2: %s", conflictStatus1, conflictStatus2)
	}

	// The issued vertex should be dropped completely, so the virtuous nodes should drop the vertex
	// and never issue the transactions into consensus.
	// Note: since the transactions will be parsed in the process, we expect the status to be "Processing" not "Rejected" or "Unknown"
	// We issue a vertex from the virtuous node to see if it builds on invalid vertex
	// This is meant to remove the need to wait an arbitrary amount of time to see if the vertex gets accepted
	// and instead confirm the valid transaction as a measure of the time to finality before checking if
	// the transactions that should have been dropped were in fact dropped successfully.
	// TODO move to test vector
	virtuousXChainAPI := e.virtuousClient.XChainAPI()
	virtuousCreateAssetTx, err := formatting.Decode(formatting.CB58, "11111111JvKptXTaQ2uHBi6XcCA7ZzcsegZxXU9tcjx16oorLZYvar6DEstkNaKtqvTuLvCrAM8rySf5rKGqMXwb9kvaKSeruScX1kZXbR5kuUAKSDsNkseR4SCMmgUvpH5jBXouhX7MyL24XvHPANVNw8gk9oqNgUmhtpXF29XmFkpY3QZWBnXeK6ZFbAQN6JXxqPMYWgUCCrc5Ddqvyy3eWQhGCud57W9AsGDAQZDphYWWLEwPbAzLbWipkuyXQxEZKWR3BQYJWQAVDyh1QxLdD8oREKpQ61qU8k7kiDzBmyU4jWTq6JvSMVmGuBCRS56pAbJxhJBEGcacAVbzyXwKUvmfxZ3QkfpDQTbbRxRXjwSoYEExN37wDAUmGZ7UFim9SF8QwgD42PC3Stxu2cDBdeU6eJVvg1Ba64ks41kCmnopJAQHix38C37EeGXXFDEtz5mYipQUk68YZusXhcyFoFZe4xAwB1fQmj3a1jByTwujxetvkeVVRDthCAYUgYsymyGA")
	if err != nil {
		return stacktrace.Propagate(err, "Failed to parse virtuous create asset tx")
	}

	// Ignore the TxID of this because it should be accepted immediately after entering consensus
	virtuousTxID, err := virtuousXChainAPI.IssueTx(e.ctx, virtuousCreateAssetTx)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to issue virtuous create asset transaction after issuing illegal vertex from byzantine node.")
	}

	for {
		status, err := virtuousXChainAPI.GetTxStatus(e.ctx, virtuousTxID)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to get virtuous transactions status from virtuous node")
		}
		if status == choices.Accepted {
			logrus.Infof("Accepted virtuous transaction with ID: %s", virtuousTxID)
			break
		} else {
			logrus.Infof("Waiting for transaction with ID: %s to be accepted", virtuousTxID)
			time.Sleep(2 * time.Second)
		}
	}

	// Once the virtuous transaction was accepted, check to see if the non-conflicting transaction
	// in an illegal vertex was accepted
	status, err = virtuousXChainAPI.GetTxStatus(e.ctx, nonConflictID)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get transaction status for non-conflicting transaction.")
	}
	logrus.Infof("Status of CreateAssetTx: %s is %s", nonConflictID, status)
	// If the transaction was Accepted, the test should fail because virtuous nodes should not issue the vertex and
	// the underlying transactions into consensus
	if status == choices.Accepted {
		return stacktrace.Propagate(err, fmt.Sprintf("Expected status of non-conflicting transaction issued in bad vertex to be Processing, but found %s", status))
	}
	return nil
}
