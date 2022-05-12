// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package networks

import (
	"bytes"
	"fmt"
	"time"

	"github.com/kurtosis-tech/kurtosis-go/lib/networks"
	"github.com/kurtosis-tech/kurtosis-go/lib/services"

	"strconv"
	"strings"

	caminoService "github.com/chain4travel/camino-testing/camino/services"
	"github.com/chain4travel/camino-testing/camino/services/certs"
	"github.com/chain4travel/camino-testing/camino_client/apis"
	"github.com/chain4travel/camino-testing/utils/constants"

	"github.com/palantir/stacktrace"
)

const (
	// The prefix for boot node configuration IDs, with an integer appended to specify each one
	bootNodeConfigIDPrefix string = "boot-node-config-"

	// The prefix for boot node service IDs, with an integer appended to specify each one
	bootNodeServiceIDPrefix string = "boot-node-"
)

// ========================================================================================================
//                                    Camino Test Network
// ========================================================================================================
const (
	containerStopTimeoutSeconds = 30
)

// TestCaminoNetwork wraps Kurtosis' ServiceNetwork that is meant to be the interface tests use for interacting with Camino
// networks
type TestCaminoNetwork struct {
	networks.Network

	svcNetwork *networks.ServiceNetwork
}

// GetCaminoClient returns the API Client for the node with the given service ID
func (network TestCaminoNetwork) GetCaminoClient(serviceID networks.ServiceID) (*apis.Client, error) {
	node, err := network.svcNetwork.GetService(serviceID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving service node with ID %v", serviceID)
	}
	caminoService := node.Service.(caminoService.CaminoService)
	jsonRPCSocket := caminoService.GetJSONRPCSocket()
	uri := fmt.Sprintf("http://%s:%d", jsonRPCSocket.GetIpAddr(), jsonRPCSocket.GetPort())
	return apis.NewClient(uri, constants.DefaultRequestTimeout), nil
}

// GetAllBootServiceIDs returns the service IDs of all the boot nodes in the network
func (network TestCaminoNetwork) GetAllBootServiceIDs() map[networks.ServiceID]bool {
	result := make(map[networks.ServiceID]bool)
	for i := 0; i < len(DefaultLocalNetGenesisConfig.Stakers); i++ {
		bootID := networks.ServiceID(bootNodeServiceIDPrefix + strconv.Itoa(i))
		result[bootID] = true
	}
	return result
}

// AddService adds a service to the test Camino network, using the given configuration
// Args:
// 		configurationID: The ID of the configuration to use for the service being added
// 		serviceID: The ID to give the service being added
// Returns:
// 		An availability checker that will return true when teh newly-added service is available
func (network TestCaminoNetwork) AddService(configurationID networks.ConfigurationID, serviceID networks.ServiceID) (*services.ServiceAvailabilityChecker, error) {
	availabilityChecker, err := network.svcNetwork.AddService(configurationID, serviceID, network.GetAllBootServiceIDs())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding service with service ID %v, configuration ID %v", serviceID, configurationID)
	}
	return availabilityChecker, nil
}

// RemoveService removes the service with the given service ID from the network
// Args:
// 	serviceID: The ID of the service to remove from the network
func (network TestCaminoNetwork) RemoveService(serviceID networks.ServiceID) error {
	if err := network.svcNetwork.RemoveService(serviceID, containerStopTimeoutSeconds); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing service with ID %v", serviceID)
	}
	return nil
}

// ========================================================================================================
//                                    Camino Service Config
// ========================================================================================================

// TestCaminoNetworkServiceConfig is Camino-specific layer of abstraction atop Kurtosis' service configurations that makes it a
// bit easier for users to define network service configurations specifically for Camino nodes
type TestCaminoNetworkServiceConfig struct {
	// Whether the certs used by Camino services created with this configuration will be different or not (which is used
	//  for testing how the network performs using duplicate node IDs)
	varyCerts bool

	// The log level that the Camino service should use
	serviceLogLevel caminoService.CaminoLogLevel

	// The image name that Camino services started from this configuration should use
	// Used primarily for Byzantine tests but can also test heterogenous Camino versions, for example.
	imageName string

	// The Snow protocol quroum size that Camino services started from this configuration should have
	snowQuorumSize int

	// The Snow protocol sample size that Camino services started from this configuration should have
	snowSampleSize int

	networkInitialTimeout time.Duration

	// TODO Make these named parameters, so we don't have an arbitrary bag of extra CLI args!
	// A list of extra CLI args that should be passed to the Camino services started with this configuration
	additionalCLIArgs map[string]string
}

// NewTestCaminoNetworkServiceConfig creates a new Camino network service config with the given parameters
// Args:
// 		varyCerts: True if the Camino services created with this configuration will have differing certs (and therefore
// 			differing node IDs), or the same cert (used for a test to see how the Camino network behaves with duplicate node
// 			IDs)
// 		serviceLogLevel: The log level that Camino services started with this configuration will use
// 		imageName: The name of the Docker image that Camino services started with this configuration will use
// 		snowQuroumSize: The Snow protocol quorum size that Camino services started with this configuration will use
// 		snowSampleSize: The Snow protocol sample size that Camino services started with this configuration will use
// 		cliArgs: A key-value mapping of extra CLI args that will be passed to Camino services started with this configuration
func NewTestCaminoNetworkServiceConfig(
	varyCerts bool,
	serviceLogLevel caminoService.CaminoLogLevel,
	imageName string,
	snowQuorumSize int,
	snowSampleSize int,
	networkInitialTimeout time.Duration,
	additionalCLIArgs map[string]string) *TestCaminoNetworkServiceConfig {
	return &TestCaminoNetworkServiceConfig{
		varyCerts:             varyCerts,
		serviceLogLevel:       serviceLogLevel,
		imageName:             imageName,
		snowQuorumSize:        snowQuorumSize,
		snowSampleSize:        snowSampleSize,
		networkInitialTimeout: networkInitialTimeout,
		additionalCLIArgs:     additionalCLIArgs,
	}
}

// ========================================================================================================
//                                Camino Test Network Loader
// ========================================================================================================

// TestCaminoNetworkLoader implements Kurtosis' NetworkLoader interface that's used for creating the test network
// of Camino services
type TestCaminoNetworkLoader struct {
	// The Docker image that should be used for the Camino boot nodes
	bootNodeImage string

	// The log level that the Camino boot nodes should use
	bootNodeLogLevel caminoService.CaminoLogLevel

	// Whether the nodes that get added to the network (boot node and otherwise) will have staking enabled
	isStaking bool

	// A registry of the service configurations available for use in this network
	serviceConfigs map[networks.ConfigurationID]TestCaminoNetworkServiceConfig

	// A mapping of (service ID) -> (service config ID) for the services that the network will initialize with
	desiredServiceConfig map[networks.ServiceID]networks.ConfigurationID

	// The Snow quorum size that the bootstrapper nodes of the network will use
	bootstrapperSnowQuorumSize int

	// The Snow sample size that the bootstrapper nodes of the network will use
	bootstrapperSnowSampleSize int

	// The fixed transaction fee for the network
	txFee uint64

	// The initial timeout for the network
	networkInitialTimeout time.Duration
}

// NewTestCaminoNetworkLoader creates a new loader to create a TestCaminoNetwork with the specified parameters, transparently handling the creation
// of bootstrapper nodes.
// NOTE: Bootstrapper nodes will be created automatically, and will show up in the ServiceAvailabilityChecker map that gets returned
// upon initialization.
// Args:
// 	isStaking: Whether the network will have staking enabled
// 	bootNodeImage: The Docker image that should be used to launch the boot nodes
// 	bootNodeLogLevel: The log level that the boot nodes will launch with
// 	bootstrapperSnowQuorumSize: The Snow consensus sample size used for nodes in the network
// 	bootstrapperSnowSampleSize: The Snow consensus quorum size used for nodes in the network
// 	serviceConfigs: A mapping of service config ID -> config info that the network will provide to the test for use
// 	desiredServiceConfigs: A map of service_id -> config_id, one per node, that this network will initialize with
func NewTestCaminoNetworkLoader(
	isStaking bool,
	bootNodeImage string,
	bootNodeLogLevel caminoService.CaminoLogLevel,
	bootstrapperSnowQuorumSize int,
	bootstrapperSnowSampleSize int,
	txFee uint64,
	networkInitialTimeout time.Duration,
	serviceConfigs map[networks.ConfigurationID]TestCaminoNetworkServiceConfig,
	desiredServiceConfigs map[networks.ServiceID]networks.ConfigurationID) (*TestCaminoNetworkLoader, error) {
	// Defensive copy
	serviceConfigsCopy := make(map[networks.ConfigurationID]TestCaminoNetworkServiceConfig)
	for configID, configParams := range serviceConfigs {
		if strings.HasPrefix(string(configID), bootNodeConfigIDPrefix) {
			return nil, stacktrace.NewError("Config ID %v cannot be used because prefix %v is reserved for boot node configurations. Choose a configuration id that does not begin with %v.",
				configID,
				bootNodeConfigIDPrefix,
				bootNodeConfigIDPrefix)
		}
		serviceConfigsCopy[configID] = configParams
	}

	// Defensive copy
	desiredServiceConfigsCopy := make(map[networks.ServiceID]networks.ConfigurationID)
	for serviceID, configID := range desiredServiceConfigs {
		if strings.HasPrefix(string(serviceID), bootNodeServiceIDPrefix) {
			return nil, stacktrace.NewError("Service ID %v cannot be used because prefix %v is reserved for boot node services. Choose a service id that does not begin with %v.",
				serviceID,
				bootNodeServiceIDPrefix,
				bootNodeServiceIDPrefix)
		}
		desiredServiceConfigsCopy[serviceID] = configID
	}

	return &TestCaminoNetworkLoader{
		bootNodeImage:              bootNodeImage,
		bootNodeLogLevel:           bootNodeLogLevel,
		isStaking:                  isStaking,
		serviceConfigs:             serviceConfigsCopy,
		desiredServiceConfig:       desiredServiceConfigsCopy,
		bootstrapperSnowQuorumSize: bootstrapperSnowQuorumSize,
		bootstrapperSnowSampleSize: bootstrapperSnowSampleSize,
		txFee:                      txFee,
		networkInitialTimeout:      networkInitialTimeout,
	}, nil
}

// ConfigureNetwork defines the netwrok's service configurations to be used
func (loader TestCaminoNetworkLoader) ConfigureNetwork(builder *networks.ServiceNetworkBuilder) error {
	localNetGenesisStakers := DefaultLocalNetGenesisConfig.Stakers
	bootNodeIDs := make([]string, 0, len(localNetGenesisStakers))
	for _, staker := range DefaultLocalNetGenesisConfig.Stakers {
		bootNodeIDs = append(bootNodeIDs, staker.NodeID)
	}

	// Add boot node configs
	for i := 0; i < len(DefaultLocalNetGenesisConfig.Stakers); i++ {
		configID := networks.ConfigurationID(bootNodeConfigIDPrefix + strconv.Itoa(i))

		certString := localNetGenesisStakers[i].TLSCert
		keyString := localNetGenesisStakers[i].PrivateKey

		certBytes := bytes.NewBufferString(certString)
		keyBytes := bytes.NewBufferString(keyString)

		initializerCore := caminoService.NewCaminoServiceInitializerCore(
			loader.bootstrapperSnowSampleSize,
			loader.bootstrapperSnowQuorumSize,
			loader.txFee,
			loader.isStaking,
			loader.networkInitialTimeout,
			make(map[string]string), // No additional CLI args for the default network
			bootNodeIDs[0:i],        // Only the node IDs of the already-started nodes
			certs.NewStaticCaminoCertProvider(*keyBytes, *certBytes),
			loader.bootNodeLogLevel,
		)
		availabilityCheckerCore := caminoService.CaminoServiceAvailabilityCheckerCore{}

		if err := builder.AddConfiguration(configID, loader.bootNodeImage, initializerCore, availabilityCheckerCore); err != nil {
			return stacktrace.Propagate(err, "An error occurred adding bootstrapper node with config ID %v", configID)
		}
	}

	// Add user-custom configs
	for configID, configParams := range loader.serviceConfigs {
		certProvider := certs.NewRandomCaminoCertProvider(configParams.varyCerts)
		imageName := configParams.imageName

		initializerCore := caminoService.NewCaminoServiceInitializerCore(
			configParams.snowSampleSize,
			configParams.snowQuorumSize,
			loader.txFee,
			loader.isStaking,
			configParams.networkInitialTimeout,
			configParams.additionalCLIArgs,
			bootNodeIDs,
			certProvider,
			configParams.serviceLogLevel,
		)
		availabilityCheckerCore := caminoService.CaminoServiceAvailabilityCheckerCore{}
		if err := builder.AddConfiguration(configID, imageName, initializerCore, availabilityCheckerCore); err != nil {
			return stacktrace.Propagate(err, "An error occurred adding Camino node configuration with ID %v", configID)
		}
	}
	return nil
}

// InitializeNetwork implements networks.NetworkLoader that initializes the Camino test network to the state specified at
// construction time, spinning up the correct number of bootstrapper nodes and subsequently the user-requested nodes.
// NOTE: The resulting services.ServiceAvailabilityChecker map will contain more IDs than the user requested as it will
// 		contain boot nodes. The IDs that these boot nodes are an unspecified implementation detail.
func (loader TestCaminoNetworkLoader) InitializeNetwork(network *networks.ServiceNetwork) (map[networks.ServiceID]services.ServiceAvailabilityChecker, error) {
	availabilityCheckers := make(map[networks.ServiceID]services.ServiceAvailabilityChecker)

	// Add the bootstrapper nodes
	bootstrapperServiceIDs := make(map[networks.ServiceID]bool)
	for i := 0; i < len(DefaultLocalNetGenesisConfig.Stakers); i++ {
		configID := networks.ConfigurationID(bootNodeConfigIDPrefix + strconv.Itoa(i))
		serviceID := networks.ServiceID(bootNodeServiceIDPrefix + strconv.Itoa(i))
		checker, err := network.AddService(configID, serviceID, bootstrapperServiceIDs)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error occurred when adding boot node with ID %v and config ID %v", serviceID, configID)
		}

		// TODO the first node should have zero dependencies and the rest should
		// have only the first node as a dependency
		bootstrapperServiceIDs[serviceID] = true
		availabilityCheckers[serviceID] = *checker
	}

	// Additional user defined nodes
	for serviceID, configID := range loader.desiredServiceConfig {
		checker, err := network.AddService(configID, serviceID, bootstrapperServiceIDs)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error occurred when adding non-boot node with ID %v and config ID %v", serviceID, configID)
		}
		availabilityCheckers[serviceID] = *checker
	}
	return availabilityCheckers, nil
}

// WrapNetwork implements a networks.NetworkLoader function and wraps the underlying networks.ServiceNetwork with the TestCaminoNetwork
func (loader TestCaminoNetworkLoader) WrapNetwork(network *networks.ServiceNetwork) (networks.Network, error) {
	return TestCaminoNetwork{
		svcNetwork: network,
	}, nil
}
