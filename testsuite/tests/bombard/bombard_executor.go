// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package bombard

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/chain4travel/camino-testing/camino_client/apis"
	"github.com/chain4travel/camino-testing/testsuite/helpers"
	"github.com/chain4travel/camino-testing/testsuite/tester"
	"github.com/chain4travel/caminogo/api"
	"github.com/chain4travel/caminogo/ids"
	"github.com/chain4travel/caminogo/utils/constants"
	"github.com/chain4travel/caminogo/utils/crypto"
	"github.com/chain4travel/caminogo/utils/formatting"
	"github.com/chain4travel/caminogo/vms/components/avax"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

// NewBombardExecutor returns a new bombard test bombardExecutor
func NewBombardExecutor(clients []*apis.Client, numTxs, txFee uint64, acceptanceTimeout time.Duration) tester.CaminoTester {
	return &bombardExecutor{
		ctx:               context.Background(),
		normalClients:     clients,
		numTxs:            numTxs,
		acceptanceTimeout: acceptanceTimeout,
		txFee:             txFee,
	}
}

type bombardExecutor struct {
	ctx               context.Context
	normalClients     []*apis.Client
	acceptanceTimeout time.Duration
	numTxs            uint64
	txFee             uint64
}

func createRandomString() string {
	return fmt.Sprintf("rand:%d", rand.Int())
}

// ExecuteTest implements the CaminoTester interface
func (e *bombardExecutor) ExecuteTest() error {
	genesisClient := e.normalClients[0]
	secondaryClients := make([]*helpers.RPCWorkFlowRunner, len(e.normalClients)-1)
	xChainAddrs := make([]string, len(e.normalClients)-1)
	for i, client := range e.normalClients[1:] {
		secondaryClients[i] = helpers.NewRPCWorkFlowRunner(
			client,
			api.UserPass{Username: createRandomString(), Password: createRandomString()},
			e.acceptanceTimeout,
		)
		xChainAddress, _, err := secondaryClients[i].CreateDefaultAddresses()
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create default addresses for client: %d", i)
		}
		xChainAddrs[i] = xChainAddress
	}

	genesisUser := api.UserPass{Username: createRandomString(), Password: createRandomString()}
	highLevelGenesisClient := helpers.NewRPCWorkFlowRunner(
		genesisClient,
		genesisUser,
		e.acceptanceTimeout,
	)

	if _, err := highLevelGenesisClient.ImportGenesisFunds(); err != nil {
		return stacktrace.Propagate(err, "Failed to fund genesis client.")
	}
	addrs, err := genesisClient.XChainAPI().ListAddresses(e.ctx, genesisUser)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get genesis client's addresses")
	}
	if len(addrs) != 1 {
		return stacktrace.NewError("Found unexecpted number of addresses for genesis client: %d", len(addrs))
	}
	genesisAddress := addrs[0]
	logrus.Infof("Imported genesis funds at address: %s", genesisAddress)

	// Fund X Chain Addresses enough to issue [numTxs]
	seedAmount := (e.numTxs + 1) * e.txFee
	if err := highLevelGenesisClient.FundXChainAddresses(xChainAddrs, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Failed to fund X Chain Addresses for Clients")
	}
	logrus.Infof("Funded X Chain Addresses with seedAmount %v.", seedAmount)

	codec, err := createXChainCodec()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to initialize codec.")
	}
	utxoLists := make([][]*avax.UTXO, len(secondaryClients))
	for i, client := range secondaryClients {
		// Each address should have [e.txFee] remaining after sending [numTxs] and paying the fixed fee each time
		if err := client.VerifyXChainAVABalance(xChainAddrs[i], seedAmount); err != nil {
			return stacktrace.Propagate(err, "Failed to verify X Chain Balane for Client: %d", i)
		}
		formattedUTXOs, _, err := genesisClient.XChainAPI().GetUTXOs(e.ctx, []string{xChainAddrs[i]}, 10, "", "")
		if err != nil {
			return err
		}
		utxos := make([]*avax.UTXO, len(formattedUTXOs))
		for i, formattedUTXO := range formattedUTXOs {
			utxo := &avax.UTXO{}
			_, err := codec.Unmarshal(formattedUTXO, utxo)
			if err != nil {
				return stacktrace.Propagate(err, "Failed to unmarshal utxo bytes.")
			}
			utxos[i] = utxo
		}
		utxoLists[i] = utxos
		logrus.Infof("Decoded %d UTXOs", len(utxos))

	}
	logrus.Infof("Verified X Chain Balances and retrieved UTXOs.")

	// Create a string of consecutive transactions for each secondary client to send
	privateKeys := make([]*crypto.PrivateKeySECP256K1R, len(secondaryClients))
	txLists := make([][][]byte, len(secondaryClients))
	txIDLists := make([][]ids.ID, len(secondaryClients))
	for i, client := range e.normalClients[1:] {
		utxo := utxoLists[i][0]
		pkStr, err := client.XChainAPI().ExportKey(e.ctx, secondaryClients[i].User(), xChainAddrs[i])
		if err != nil {
			return stacktrace.Propagate(err, "Failed to export key.")
		}

		if !strings.HasPrefix(pkStr, constants.SecretKeyPrefix) {
			return fmt.Errorf("private key missing %s prefix", constants.SecretKeyPrefix)
		}
		trimmedPrivateKey := strings.TrimPrefix(pkStr, constants.SecretKeyPrefix)
		formattedPrivateKey, err := formatting.Decode(formatting.CB58, trimmedPrivateKey)
		if err != nil {
			return fmt.Errorf("problem parsing private key: %w", err)
		}

		factory := crypto.FactorySECP256K1R{}
		skIntf, err := factory.ToPrivateKey(formattedPrivateKey)
		sk := skIntf.(*crypto.PrivateKeySECP256K1R)
		privateKeys[i] = sk

		logrus.Infof("Creating string of %d transactions", e.numTxs)
		txs, txIDs, err := CreateConsecutiveTransactions(utxo, e.numTxs, seedAmount, e.txFee, sk)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create transaction list.")
		}
		txLists[i] = txs
		txIDLists[i] = txIDs
	}

	wg := sync.WaitGroup{}
	issueTxsAsync := func(runner *helpers.RPCWorkFlowRunner, txList [][]byte) {
		if err := runner.IssueTxList(txList); err != nil {
			panic(err)
		}
		wg.Done()
	}

	startTime := time.Now()
	logrus.Infof("Beginning to issue transactions...")
	for i, client := range secondaryClients {
		wg.Add(1)
		issueTxsAsync(client, txLists[i])
	}
	wg.Wait()

	duration := time.Since(startTime)
	logrus.Infof("Finished issuing transaction lists in %v seconds.", duration.Seconds())
	for _, txIDs := range txIDLists {
		if err := highLevelGenesisClient.AwaitXChainTxs(txIDs...); err != nil {
			stacktrace.Propagate(err, "Failed to confirm transactions.")
		}
	}

	logrus.Infof("Confirmed all issued transactions.")

	return nil
}
