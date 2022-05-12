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
	"github.com/kurtosis-tech/kurtosis-go/lib/services"
)

// NodeService implements the Kurtosis generic services.Service interface that represents the minimum interface for a
// validator node
type NodeService interface {
	services.Service

	// GetStakingSocket returns the socket used for communication between nodes on the network
	GetStakingSocket() ServiceSocket
}
