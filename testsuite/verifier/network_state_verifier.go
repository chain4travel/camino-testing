// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package verifier

import (
	"context"

	"github.com/chain4travel/camino-testing/camino_client/apis"
	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

// NetworkStateVerifier contains logic for verifying the state of the network
// We attach these functions to a struct even though the struct doesn't have state to avoid a utils class (which
// inevitably becomes a mess of unconnected logic), and to categorize the functions around a common purpose.
type NetworkStateVerifier struct {
	ctx context.Context
}

func NewNetworkStateVerifier() NetworkStateVerifier {
	return NetworkStateVerifier{ctx: context.Background()}
}

func (s *NetworkStateVerifier) Ctx() context.Context { return s.ctx }

// VerifyNetworkFullyConnected asserts that the network is fully connected
// Meaning:
// 		1) The stakers have all the other nodes in the network besides themselves in their peer list
// 		2) All non-stakers have all the stakers in their peer list
// Args:
// 	allServiceIDs: All the service IDs in the network, and the IDs that will be iterated over to check
// 	stakerServiceIDs: The service IDs of nodes that we expect to be fully connected - i.e. any node that's actually
// 		staking. Most of the time this will be just the bootstrappers, but if we add more stakers then this set will
// 		expand beyond the bootstrappers.
// 	allNodeIDs: The mapping of servcie_id -> node_id
func (verifier NetworkStateVerifier) VerifyNetworkFullyConnected(
	allServiceIDs map[networks.ServiceID]bool,
	stakerServiceIDs map[networks.ServiceID]bool,
	allNodeIDs map[networks.ServiceID]string,
	allAvalalancheClients map[networks.ServiceID]*apis.Client,
) error {
	logrus.Tracef("All node IDs in network being verified: %v", allNodeIDs)
	for serviceID := range allServiceIDs {
		_, isStaker := stakerServiceIDs[serviceID]

		acceptableNodeIDs := make(map[string]bool)
		for comparisonID := range allServiceIDs {
			// Nodes will never have themselves in their peer list
			if serviceID == comparisonID {
				continue
			}
			_, isComparisonStaker := stakerServiceIDs[comparisonID]

			// Staker nodes will have all other nodes in their peer list
			// Non-stakers will only have the stakers
			if isStaker || (!isStaker && isComparisonStaker) {
				comparisonNodeID := allNodeIDs[comparisonID]
				acceptableNodeIDs[comparisonNodeID] = true
			}
		}

		logrus.Infof("Expecting serviceID %v to have the following peer node IDs, %v", serviceID, acceptableNodeIDs)
		if err := verifier.VerifyExpectedPeers(serviceID, allAvalalancheClients[serviceID], acceptableNodeIDs, len(acceptableNodeIDs), false); err != nil {
			return stacktrace.Propagate(err, "An error occurred verifying the expected peers list")
		}
	}
	return nil
}

// VerifyExpectedPeers verifies that a node's actual peers match the expected value
// Args:
// 		serviceID: Service ID of the node whose peers are being examined
// 		client: camino client for the node being examined
// 		acceptableNodeIDs: A "set" of acceptable node IDs where, if a peer doesn't have this ID, the test will be failed
// 		expectedNumPeers: The number of peers we expect this node to have
// 		atLeast: If true, indicates that the number of peers must be AT LEAST the expected number of peers; if false, must be exact
func (verifier NetworkStateVerifier) VerifyExpectedPeers(
	serviceID networks.ServiceID,
	client *apis.Client,
	acceptableNodeIDs map[string]bool,
	expectedNumPeers int,
	atLeast bool) error {
	peers, err := client.InfoAPI().Peers(verifier.ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get peers from service with ID %v", serviceID)
	}

	actualNumPeers := len(peers)
	var condition bool
	var operatorAsserted string
	if atLeast {
		condition = actualNumPeers >= expectedNumPeers
		operatorAsserted = ">="
	} else {
		condition = actualNumPeers == expectedNumPeers
		operatorAsserted = "=="
	}

	if !condition {
		return stacktrace.NewError(
			"Service ID %v actual num peers, %v, is not %v expected num peers, %v",
			serviceID,
			actualNumPeers,
			operatorAsserted,
			expectedNumPeers,
		)
	}

	// Verify that IDs of the peers we have are in our list of acceptable IDs
	for _, peer := range peers {
		_, found := acceptableNodeIDs[peer.ID]
		if !found {
			return stacktrace.NewError("Service ID %v has a peer with node ID %s that we don't recognize", serviceID, peer.ID)
		}
	}
	return nil
}
