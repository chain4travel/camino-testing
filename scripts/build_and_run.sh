set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"

# ====================== CONSTANTS =======================================================
SUITE_IMAGE="c4tplatform/camino-testing"
CAMINO_IMAGE="c4tplatform/caminogo:v1.0.0"
BYZANTINE_IMAGE="c4tplatform/caminogo-byzantine:v0.0.0"
KURTOSIS_CORE_CHANNEL="master"
INITIALIZER_IMAGE="kurtosistech/kurtosis-core_initializer:${KURTOSIS_CORE_CHANNEL}"
API_IMAGE="kurtosistech/kurtosis-core_api:${KURTOSIS_CORE_CHANNEL}"

PARALLELISM=2

BUILD_ACTION="build"
RUN_ACTION="run"
BOTH_ACTION="all"
HELP_ACTION="help"

# ====================== ARG PARSING =======================================================
show_help() {
    echo "${0} <action> <extra Docker args...>"
    echo ""
    echo "  Actions:"
    echo "    help    Displays this messages"
    echo "    build   Executes only the build step, skipping the run step"
    echo "    run     Executes only the run step, skipping the build step"
    echo "    all     Executes both build and run steps"
    echo ""
    echo "  Example:"
    echo "    ${0} all --env PARALLELISM=4"
    echo ""
}

if [ "${#}" -eq 0 ]; then
    show_help
    exit 0
fi

action="${1:-}"
shift 1

do_build=true
do_run=true
case "${action}" in
    ${HELP_ACTION})
        show_help
        exit 0
        ;;
    ${BUILD_ACTION})
        do_build=true
        do_run=false
        ;;
    ${RUN_ACTION})
        do_build=false
        do_run=true
        ;;
    ${BOTH_ACTION})
        do_build=true
        do_run=true
        ;;
    *)
        echo "Error: First argument must be one of '${HELP_ACTION}', '${BUILD_ACTION}', '${RUN_ACTION}', or '${BOTH_ACTION}'" >&2
        exit 1
        ;;
esac

# ====================== MAIN LOGIC =======================================================
git_branch="$(git rev-parse --abbrev-ref HEAD)"
docker_tag="$(echo "${git_branch}" | sed 's,[/:],_,g')"

root_dirpath="$(dirname "${script_dirpath}")"
if "${do_build}"; then
    echo "Running unit tests..."
    if ! go test "${root_dirpath}/..."; then
        echo "Tests failed!"
        exit 1
    else
        echo "Tests succeeded"
    fi

    echo "Building Camino testing suite image..."
    docker build -t "${SUITE_IMAGE}:${docker_tag}" -f "${root_dirpath}/testsuite/Dockerfile" "${root_dirpath}"
fi

if "${do_run}"; then
    suite_execution_volume="camino-test-suite_${docker_tag}_$(date +%s)"
    docker volume create "${suite_execution_volume}"

    # Docker only allows you to have spaces in the variable if you escape them or use a Docker env file
    custom_env_vars_json_flag="CUSTOM_ENV_VARS_JSON={\"CAMINO_IMAGE\":\"${CAMINO_IMAGE}\",\"BYZANTINE_IMAGE\":\"${BYZANTINE_IMAGE}\"}"

    echo "${custom_env_vars_json_flag}"
    docker run \
        --mount "type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock" \
        --mount "type=volume,source=${suite_execution_volume},target=/suite-execution" \
        --env "${custom_env_vars_json_flag}" \
        --env "TEST_SUITE_IMAGE=${SUITE_IMAGE}:${docker_tag}" \
        --env "SUITE_EXECUTION_VOLUME=${suite_execution_volume}" \
        --env "KURTOSIS_API_IMAGE=${API_IMAGE}" \
        --env "PARALLELISM=${PARALLELISM}" \
        `# In Bash, this is how you feed arguments exactly as-is to a child script (since ${*} loses quoting and ${@} trips set -e if no arguments are passed)` \
        `# It basically says, "if and only if ${1} exists, evaluate ${@}"` \
        ${1+"${@}"} \
        "${INITIALIZER_IMAGE}"
fi
