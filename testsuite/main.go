// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package main

import (
	"flag"
	"fmt"
	"os"

	testsuite "github.com/chain4travel/camino-testing/testsuite/kurtosis"
	"github.com/kurtosis-tech/kurtosis-go/lib/client"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	// --------- Kurtosis-internal params --------------------------------------
	metadataFilepath := flag.String(
		"metadata-filepath",
		"",
		"The filepath of the file in which the test suite metadata should be written")
	testArg := flag.String(
		"test",
		"",
		"The name of the test to run")
	kurtosisApiIpArg := flag.String(
		"kurtosis-api-ip",
		"",
		"IP address of the Kurtosis API endpoint")
	logLevelArg := flag.String(
		"log-level",
		"",
		"String corresponding to Logrus log level that the test suite will output with",
	)
	servicesDirpathArg := flag.String(
		"services-relative-dirpath",
		"",
		"Dirpath, relative to the root of the suite execution volume, where directories for each service should be created")

	// ----------------------- Camino testing-custom params ---------------------------------
	caminogoImageArg := flag.String(
		"camino-go-image",
		"",
		"Name of Camino Go Docker image that will be used to launch Camino Go nodes")
	byzantineGoImageArg := flag.String(
		"byzantine-go-image",
		"",
		"Name of Byzantine Camino Go Docker image that will be used to launch Camino Go nodes with Byzantine behaviour")

	flag.Parse()

	level, err := logrus.ParseLevel(*logLevelArg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred parsing the log level string: %v\n", err)
		os.Exit(1)
	}
	logrus.SetLevel(level)

	logrus.Debugf("Byzantine image name: %s", *byzantineGoImageArg)
	testSuite := testsuite.CaminoTestSuite{
		ByzantineImageName: *byzantineGoImageArg,
		NormalImageName:    *caminogoImageArg,
	}
	exitCode := client.Run(testSuite, *metadataFilepath, *servicesDirpathArg, *testArg, *kurtosisApiIpArg)
	os.Exit(exitCode)
}
