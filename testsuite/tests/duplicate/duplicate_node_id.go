// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package duplicate

import (
	"context"
	"time"

	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"

	caminoNetwork "github.com/chain4travel/camino-testing/camino/networks"
	caminoService "github.com/chain4travel/camino-testing/camino/services"
	"github.com/chain4travel/camino-testing/camino_client/apis"
	"github.com/chain4travel/camino-testing/testsuite/verifier"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	normalNodeConfigID networks.ConfigurationID = "normal-config"
	sameCertConfigID   networks.ConfigurationID = "same-cert-config"

	vanillaNodeServiceID networks.ServiceID = "vanilla-node"
	badServiceID1        networks.ServiceID = "bad-service-1"
	badServiceID2        networks.ServiceID = "bad-service-2"
)

// DuplicateNodeIDTest adds a node with a duplicate nodeID and ensures that the network handles the duplicate
// appropriately and settles on the remaining node when the duplicate is removed.
type DuplicateNodeIDTest struct {
	ImageName string
	Verifier  verifier.NetworkStateVerifier
}

// Run implements the Kurtosis Test interface
func (test DuplicateNodeIDTest) Run(network networks.Network, context testsuite.TestContext) {
	castedNetwork := network.(caminoNetwork.TestCaminoNetwork)

	bootServiceIDs := castedNetwork.GetAllBootServiceIDs()

	allServiceIDs := make(map[networks.ServiceID]bool)
	for bootServiceID := range bootServiceIDs {
		allServiceIDs[bootServiceID] = true
	}
	allServiceIDs[vanillaNodeServiceID] = true

	allNodeIDs, allCaminoClients := getNodeIDsAndClients(test.Verifier.Ctx(), context, castedNetwork, allServiceIDs)
	if err := test.Verifier.VerifyNetworkFullyConnected(allServiceIDs, bootServiceIDs, allNodeIDs, allCaminoClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
	}

	// We'll need these later
	originalServiceIDs := make(map[networks.ServiceID]bool)
	for serviceID := range allServiceIDs {
		originalServiceIDs[serviceID] = true
	}

	logrus.Debugf("Service IDs before adding any nodes: %v", allServiceIDs)
	logrus.Debugf("Camino node IDs before adding any nodes: %v", allNodeIDs)

	// Add the first dupe node ID (should look normal from a network perspective
	logrus.Info("Adding first node with soon-to-be-duplicated node ID...")
	checker1, err := castedNetwork.AddService(sameCertConfigID, badServiceID1)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to create first dupe node ID service with ID %v", badServiceID1))
	}
	if err := checker1.WaitForStartup(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred waiting for first dupe node ID service with ID %v to start", badServiceID1))
	}
	allServiceIDs[badServiceID1] = true

	badServiceClient1, err := castedNetwork.GetCaminoClient(badServiceID1)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred getting the Camino client for the first dupe node ID service with ID %v", badServiceID1))
	}
	allCaminoClients[badServiceID1] = badServiceClient1

	badServiceNodeID1, err := badServiceClient1.InfoAPI().GetNodeID(test.Verifier.Ctx())
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get node ID from first dupe node ID service with ID %v", badServiceID1))
	}
	allNodeIDs[badServiceID1] = badServiceNodeID1

	logrus.Info("Successfully added first node with soon-to-be-duplicated ID")

	// Verify that the new node got accepted by everyone
	logrus.Infof("Verifying that the new node with service ID %v was accepted by all bootstrappers...", badServiceID1)
	if err := test.Verifier.VerifyNetworkFullyConnected(allServiceIDs, bootServiceIDs, allNodeIDs, allCaminoClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
	}
	logrus.Infof("New node with service ID %v was accepted by all bootstrappers", badServiceID1)

	// Now, add a second node with the same ID
	logrus.Infof("Adding second node with service ID %v which will be a duplicated node ID...", badServiceID2)
	checker2, err := castedNetwork.AddService(sameCertConfigID, badServiceID2)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Failed to create second dupe node ID service with ID %v", badServiceID2))
	}
	if err := checker2.WaitForStartup(); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred waiting for second dupe node ID service to start"))
	}
	allServiceIDs[badServiceID2] = true

	badServiceClient2, err := castedNetwork.GetCaminoClient(badServiceID2)
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred getting the Camino client for the second dupe node ID service with ID %v", badServiceID2))
	}
	allCaminoClients[badServiceID2] = badServiceClient2

	badServiceNodeID2, err := badServiceClient2.InfoAPI().GetNodeID(test.Verifier.Ctx())
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get node ID from first dupe node ID service with ID %v", badServiceID2))
	}
	allNodeIDs[badServiceID2] = badServiceNodeID2
	logrus.Info("Second node added, causing duplicate node ID")

	// At this point, it's undefined what happens with the two nodes with duplicate IDs; verify that the original nodes
	//  in the network operate normally amongst themselves
	logrus.Info("Connection behaviour to nodes with duplicate IDs is undefined, so verifying that the original nodes connect as expected...")
	for serviceID := range originalServiceIDs {
		acceptableNodeIDs := make(map[string]bool)

		// All original nodes should have the boot nodes (though a boot node won't have itself)
		for bootServiceID := range bootServiceIDs {
			if serviceID != bootServiceID {
				bootNodeID := allNodeIDs[bootServiceID]
				acceptableNodeIDs[bootNodeID] = true
			}
		}

		if _, found := bootServiceIDs[serviceID]; found {
			// Boot nodes should have the original node, one of the duplicates, and MAY have the duplicate nodes
			acceptableNodeIDs[allNodeIDs[vanillaNodeServiceID]] = true
			acceptableNodeIDs[badServiceNodeID1] = true
			acceptableNodeIDs[badServiceNodeID2] = true
			if err := test.Verifier.VerifyExpectedPeers(serviceID, allCaminoClients[serviceID], acceptableNodeIDs, len(originalServiceIDs)-1, true); err != nil {
				context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
			}
		} else {
			// The original non-boot node should have exactly the boot nodes
			if err := test.Verifier.VerifyExpectedPeers(serviceID, allCaminoClients[serviceID], acceptableNodeIDs, len(bootServiceIDs), false); err != nil {
				context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
			}
		}
	}
	logrus.Info("Verified that original nodes are still connected to each other")

	// Now, kill the first dupe node to leave only the second (who everyone should connect with)
	logrus.Info("Removing first node with duplicate ID...")
	if err := castedNetwork.RemoveService(badServiceID1); err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not remove the first service with duped node ID"))
	}
	delete(allServiceIDs, badServiceID1)
	delete(allCaminoClients, badServiceID1)
	delete(allNodeIDs, badServiceID1)
	logrus.Info("Successfully removed first node with duplicate ID, leaving only the second")

	// Now that the first duped node is gone, verify that the original node is still connected to just boot nodes and
	//  the second duped-ID node is now accepted by the boot nodes
	logrus.Info("Verifying that the network has connected to the second node with a previously-duplicated node ID...")
	if err := test.Verifier.VerifyNetworkFullyConnected(allServiceIDs, bootServiceIDs, allNodeIDs, allCaminoClients); err != nil {
		context.Fatal(stacktrace.Propagate(err, "An error occurred verifying the network's state"))
	}
	logrus.Info("Verified that the network has settled on the second node with previously-duplicated ID")
}

// GetNetworkLoader implements the Kurtosis Test interface
func (test DuplicateNodeIDTest) GetNetworkLoader() (networks.NetworkLoader, error) {
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
		sameCertConfigID: *caminoNetwork.NewTestCaminoNetworkServiceConfig(
			false,
			caminoService.DEBUG,
			test.ImageName,
			2,
			2,
			2*time.Second,
			make(map[string]string),
		),
	}
	desiredServices := map[networks.ServiceID]networks.ConfigurationID{
		vanillaNodeServiceID: normalNodeConfigID,
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
func (test DuplicateNodeIDTest) GetExecutionTimeout() time.Duration {
	return 5 * time.Minute
}

// GetSetupBuffer implements the Kurtosis Test interface
func (test DuplicateNodeIDTest) GetSetupBuffer() time.Duration {
	// TODO drop this when the availabilityChecker doesn't have a sleep (because we spin up a bunch of nodes before execution)
	return 6 * time.Minute
}

// ================ Helper functions ==================================
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
