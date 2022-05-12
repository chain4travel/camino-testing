// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package constants

import (
	"github.com/chain4travel/caminogo/ids"
)

var (
	// XChainID ...
	XChainID ids.ID
	// PlatformChainID ...
	PlatformChainID ids.ID
	// CChainID ...
	CChainID ids.ID
	// AvaxAssetID ...
	AvaxAssetID ids.ID
)

func init() {
	XChainID, _ = ids.FromString("2eNy1mUFdmaxXNj1eQHUe7Np4gju9sJsEtWQ4MX3ToiNKuADed")
	PlatformChainID = ids.Empty
	CChainID, _ = ids.FromString("WKNkfmNxgqpKPe9Q12UCoTuGYXX5JbQn2tf2WTpNTJeQrezqa")
	AvaxAssetID, _ = ids.FromString("2fombhL7aGPwj3KH4bfrmJwW6PVnMobf9Y2fn9GwxiAAJyFDbe")
}
