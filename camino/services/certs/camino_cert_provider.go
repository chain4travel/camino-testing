// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package certs

import "bytes"

// CaminoCertProvider defines an interface representing a cert provider for an Camino service
// (used in the duplicate node ID test, which requires that multiple Camino services start
// with the same cert)
type CaminoCertProvider interface {
	// GetCertAndKey generates a cert and accompanying private key
	// Returns:
	// 	certPemBytes: The bytes of the generated cert
	// 	keyPemBytes: The bytes of the private key generated with the cert
	GetCertAndKey() (certPemBytes bytes.Buffer, keyPemBytes bytes.Buffer, err error)
}
