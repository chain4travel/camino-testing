// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package services

// CaminoService implements CaminoService
type CaminoService struct {
	ipAddr      string
	stakingPort int
	jsonRPCPort int
}

// GetStakingSocket implements CaminoService
func (service CaminoService) GetStakingSocket() ServiceSocket {
	return *NewServiceSocket(service.ipAddr, service.stakingPort)
}

// GetJSONRPCSocket implements CaminoService
func (service CaminoService) GetJSONRPCSocket() ServiceSocket {
	return *NewServiceSocket(service.ipAddr, service.jsonRPCPort)
}
