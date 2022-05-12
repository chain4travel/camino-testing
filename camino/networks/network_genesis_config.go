// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package networks

// NetworkGenesisConfig encapusulates genesis information describing
// a network
type NetworkGenesisConfig struct {
	Stakers         []StakerIdentity
	FundedAddresses FundedAddress
}

// FundedAddress encapsulates a pre-funded address
type FundedAddress struct {
	Address    string
	PrivateKey string
}

// StakerIdentity contains a staker's identifying information
type StakerIdentity struct {
	NodeID     string
	PrivateKey string
	TLSCert    string
}
