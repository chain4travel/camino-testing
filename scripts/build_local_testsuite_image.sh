#!/bin/bash

set -euo pipefail

# Note: this script will build a docker image by cloning a remote version of camino-testing and caminogo into a temporary
# location and using that version's Dockerfile to build the image.
SCRIPT_DIRPATH=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
ROOT_DIRPATH="$(dirname "${SCRIPT_DIRPATH}")"
CAMINO_PATH="$GOPATH/src/github.com/chain4travel/caminogo"
E2E_COMMIT="$(git --git-dir="$ROOT_DIRPATH/.git" rev-parse --short HEAD)"
CAMINO_COMMIT="$(git --git-dir="$CAMINO_PATH/.git" rev-parse --short HEAD)"

export GOPATH="$SCRIPT_DIRPATH/.build_image_gopath"
WORKPREFIX="$GOPATH/src/github.com/chain4travel"
DOCKER="${DOCKER:-docker}"


CAMINO_REMOTE="https://github.com/chain4travel/caminogo-internal.git"
E2E_REMOTE="https://github.com/chain4travel/camino-testing.git"


# Clone the remotes and checkout the desired branch/commits
CAMINO_CLONE="$WORKPREFIX/caminogo"
E2E_CLONE="$WORKPREFIX/camino-testing"

# Create the WORKPREFIX directory if it does not exist yet
if [[ ! -d "$WORKPREFIX" ]]; then
    mkdir -p "$WORKPREFIX"
fi

# Configure git credential helper
git config --global credential.helper cache

if [[ ! -d "$CAMINO_CLONE" ]]; then
    git clone "$CAMINO_REMOTE" "$CAMINO_CLONE"
else
    git -C "$CAMINO_CLONE" fetch origin
fi

git -C "$CAMINO_CLONE" checkout "$CAMINO_COMMIT"

if [[ ! -d "$E2E_CLONE" ]]; then
    git clone "$E2E_REMOTE" "$E2E_CLONE"
else
    git -C "$E2E_CLONE" fetch origin
fi

git -C "$E2E_CLONE" checkout "$E2E_COMMIT"


DOCKER_ORG="c4tplatform"
REPO_BASE="camino-testing"
CONTROLLER_REPO="${REPO_BASE}_controller"

CONTROLLER_TAG="$DOCKER_ORG/$CONTROLLER_REPO-$E2E_COMMIT-$CAMINO_COMMIT"

"${DOCKER}" build -t "${CONTROLLER_TAG}" "${WORKPREFIX}" -f "$ROOT_DIRPATH/testsuite/local.Dockerfile"
