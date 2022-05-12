Camino End-to-End Tests
====================
This repo contains end-to-end tests for the Camino network and caminogo client using [the Kurtosis testing framework](https://github.com/kurtosis-tech/kurtosis-docs)

* [Requirements](#requirements)
* [Running Locally](#running-locally)
* [Developing Locally](#developing-locally)
    * [Architecture](#architecture)
    * [Adding A Test](#adding-a-test)
    * [Running Locally As A Developer](#running-locally-as-a-developer)
    * [Keeping Your Dev Environment Clean](#keeping-your-dev-environment-clean)

Requirements
------------
* Golang version 1.17.x
* [A Docker engine running in your environment](https://docs.docker.com/engine/install/)

Running Locally
---------------
1. Clone this repository
1. Run `scripts/build_and_run.sh all` and wait for it to finish (which will take a while)

This command will spin up multiple Docker containers during its operation, and you can examine container logs via either your Docker engine dashboard GUI or the `docker container ls` and `docker container logs`.

NOTE: The Camino E2E test suite defaults to running 4 tests in parallel to speed up test suite execution time. If your machine has less cores, you should reduce this parallelism to _at maximum_ the number of cores on your machine, else the extra context-switching will slow down test execution and potentially cause spurious failures. To set the paralleism, pass the `--env PARALLELISM=N` argument to `build_and_run.sh` (where "N" is the desired number of threads).

Once `build_and_run.sh all` has finished, you can now execute `build_and_run.sh run` to re-run the testing suite without needing to rebuild. To see full help information for the `build_and_run.sh` script, pass in the `help` action like so: `build_and_run.sh help`.

Developing Locally
------------------
This repo uses the [Kurtosis architecture](https://github.com/kurtosis-tech/kurtosis-docs), so you should first go through the tutorial there to familiarize yourself with the core Kurtosis concepts.

In this implementation of Kurtosis, we have:
* `NodeService` interface to represent the actions a test can take against a generic node exposing a staking socket and implementing the Kurtosis services.Service interface
* `CaminoService` interface to represent the actions a test can take against an caminogo client participating in the Camino network being tested
* `CaminoServiceInitializerCore` and `CaminoServiceAvailabilityChecker` for instantiating caminogo clients in test networks
    * `CaminoCertProvider` to allow controlling the cert that a caminogo node starts with, to allow for writing duplicate-node-ID tests
* `TestCaminoNetwork` to encapsulate a test Camino network of caminogo nodes of arbitrary size
* Several tests
* `CaminoTestSuite` to contain all the tests Kurtosis can run
* A `Dockerfile` for building the testsuite image under the `testsuite` package
* A `main.go` under the `testsuite` package for running the logic that the testsuite image will run

Additionally, for ease of writing tests, this repo also contains a Go client for interacting with the JSON RPC API of an Camino node (which will be moved to an external repo).

### Adding A Test
1. Create a new directory in `testsuite/tests` for your test
2. Create a struct that implements the `testsuite.Test` interface from Kurtosis
3. Fill in the interface's functions
4. Register the test in `CaminoTestSuite`'s `GetTests` method

### Running Your Code
The `scripts/build_and_run.sh all` will rebuild the testsuite Docker image and run the tests inside; rerun this every time that you make a change. You can also pass in extra Docker parameters using the `--env ARGNAME=argvalue` to modify the runtime behaviour of Kurtosis, e.g. `scripts/build_and_run.sh all --env PARALLELISM=2`. For the full list of arguments, see [the Kurtosis docs](https://github.com/kurtosis-tech/kurtosis-docs#details-1).

### Keeping Your Dev Environment Clean
Kurtosis intentionally doesn't delete containers and volumes, which means your local Docker environment will accumulate images, containers, and volumes; you can use [the script here](./scripts/clean_docker_environment.sh) to clean old containers and images. For further information, read [the Notes section of the Kurtosis README](https://github.com/kurtosis-tech/kurtosis-docs#abnormal-exit) for more details on how to keep your local environment clean while you develop.
