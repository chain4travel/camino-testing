set -euo pipefail
SCRIPT_DIRPATH="$(cd "$(dirname "${0}")" && pwd)"
ROOT_DIRPATH="$(dirname "${SCRIPT_DIRPATH}")"
PARALLELISM=4

DOCKER_REPO="c4tplatform"

# login to AWS for byzantine images
echo "$DOCKER_PASS" | docker login --username "$DOCKER_USERNAME" --password-stdin

# Use stable version of Everest for CI
CAMINO_IMAGE="$DOCKER_REPO/caminogo:v1.0.1"
# Use stable version of caminogo-byzantine (tbc)
BYZANTINE_IMAGE="$DOCKER_REPO/caminogo-byzantine:v0.0.0"

# Kurtosis doesn't currently support pulling from Docker repos that require authentication
# so we have to do the pull here
docker pull "${BYZANTINE_IMAGE}"
docker pull "${CAMINO_IMAGE}"

E2E_TEST_COMMAND="${ROOT_DIRPATH}/scripts/build_and_run.sh"

# Docker only allows you to have spaces in the variable if you escape them or use a Docker env file
CUSTOM_ENV_VARS_JSON_ARG="CUSTOM_ENV_VARS_JSON={\"CAMINO_IMAGE\":\"${CAMINO_IMAGE}\",\"BYZANTINE_IMAGE\":\"${BYZANTINE_IMAGE}\"}"

return_code=0
if ! bash "${E2E_TEST_COMMAND}" all --env "${CUSTOM_ENV_VARS_JSON_ARG}" --env "PARALLELISM=${PARALLELISM}"; then
    echo "Camino E2E tests failed"
    return_code=1
else
    echo "Camino E2E tests succeeded"
    return_code=0
fi

exit "${return_code}"
