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

// StaticCaminoCertProvider implements CaminoCertProvider and provides the same cert every time
type StaticCaminoCertProvider struct {
	key  bytes.Buffer
	cert bytes.Buffer
}

// NewStaticCaminoCertProvider creates an instance of StaticCaminoCertProvider using the given key and cert
// Args:
// 	key: The private key that the StaticCaminoCertProvider will return on every call to GetCertAndKey
// 	cert: The cert that will be returned on every call to GetCertAndKey
func NewStaticCaminoCertProvider(key bytes.Buffer, cert bytes.Buffer) *StaticCaminoCertProvider {
	return &StaticCaminoCertProvider{key: key, cert: cert}
}

// GetCertAndKey returns the same cert and key that was configured at the time of construction
func (s StaticCaminoCertProvider) GetCertAndKey() (certPemBytes bytes.Buffer, keyPemBytes bytes.Buffer, err error) {
	return s.cert, s.key, nil
}
