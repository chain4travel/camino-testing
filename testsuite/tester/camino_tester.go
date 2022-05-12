// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package tester

// CaminoTester is the interface for a ready to execute test
type CaminoTester interface {
	ExecuteTest() error
}
