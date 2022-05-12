// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package services

import (
	"context"
	"fmt"
	"time"

	"github.com/chain4travel/caminogo/api/info"
	"github.com/kurtosis-tech/kurtosis-go/lib/services"
)

// NewCaminoServiceAvailabilityChecker returns a new services.ServiceAvailabilityCheckerCore to
// check if an CaminoService is ready
func NewCaminoServiceAvailabilityChecker(timeout time.Duration) services.ServiceAvailabilityCheckerCore {
	return &CaminoServiceAvailabilityCheckerCore{
		timeout: timeout,
	}
}

// CaminoServiceAvailabilityCheckerCore implements services.ServiceAvailabilityCheckerCore
// that defines the criteria for an Camino service being available
type CaminoServiceAvailabilityCheckerCore struct {
	timeout                                                    time.Duration
	bootstrappedPChain, bootstrappedCChain, bootstrappedXChain bool
}

// IsServiceUp implements services.ServiceAvailabilityCheckerCore#IsServiceUp
// and returns true when the Camino healthcheck reports that the node is available
func (g CaminoServiceAvailabilityCheckerCore) IsServiceUp(toCheck services.Service, dependencies []services.Service) bool {
	// NOTE: we don't check the dependencies intentionally, because we don't need to - an Camino service won't report itself
	//  as up until its bootstrappers are up

	castedService := toCheck.(CaminoService)
	jsonRPCSocket := castedService.GetJSONRPCSocket()
	uri := fmt.Sprintf("http://%s:%d", jsonRPCSocket.GetIpAddr(), jsonRPCSocket.GetPort())
	client := info.NewClient(uri)

	if !g.bootstrappedPChain {
		if bootstrapped, err := client.IsBootstrapped(context.Background(), "P"); err != nil || !bootstrapped {
			return false
		}
	}

	if !g.bootstrappedCChain {
		if bootstrapped, err := client.IsBootstrapped(context.Background(), "C"); err != nil || !bootstrapped {
			return false
		}
	}

	if !g.bootstrappedXChain {
		if bootstrapped, err := client.IsBootstrapped(context.Background(), "X"); err != nil || !bootstrapped {
			return false
		}
	}

	time.Sleep(5 * time.Second)
	return true
}

// GetTimeout implements services.AvailabilityCheckerCore
func (g CaminoServiceAvailabilityCheckerCore) GetTimeout() time.Duration {
	return 90 * time.Second
}
