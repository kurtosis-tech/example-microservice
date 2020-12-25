set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"
root_dirpath"$(dirname "${script_dirpath}")"

API_IMAGE_NAME="kurtosistech/example-microservices_api"
DATASTORE_IMAGE_NAME="kurtosistech/example-microservices_datastore"

# ========================================================================================================
#                                           Arg Parsing
# ========================================================================================================
function print_help_and_exit() {
    echo "Usage: $(basename "${0}") (build|publish)"
    echo ""
    echo "  Builds and/or publishes the example microservice Docker images"
    echo ""
    echo "  build       Builds the example microservice Docker images"
    echo "  publish     Builds the example microservice Docker images, and publishes the microservice images to the Kurtosis Dockerhub"
    echo ""
    exit 1  # Exit with error so CI fails if this gets accidentally called
}

if [ "${#}" -ne 1 ]; then
    print_help_and_exit
fi

do_build=false
do_publish=false

arg="${1}"
case "${arg}" in
    build)
        do_build=true
        ;;
    publish)
        do_build=true
        do_publish=true
        ;;
    *)
        echo "Error: Unrecognized argument '${arg}'" >&2
        print_help_and_exit
        ;;
esac

# ========================================================================================================
#                                           Main Code
# ========================================================================================================
if "${do_build}"; then
    if ! [ -f "${root_dirpath}"/.dockerignore ]; then
        echo "Error: No .dockerignore file found in root; this is required so Docker caching works properly" >&2
        exit 1
    fi

    echo "Building API image..."
    docker build -t "${API_IMAGE_NAME}" -f "${root_dirpath}/api/Dockerfile" "${root_dirpath}"
    echo "API image built"

    echo "Building datastore image..."
    docker build -t "${DATASTORE_IMAGE_NAME}" -f "${root_dirpath}/datastore/Dockerfile" "${root_dirpath}"
    echo "Datastore image built"
fi

if "${do_publish}"; then
    docker push "${API_IMAGE_NAME}"
    docker push "${DATASTORE_IMAGE_NAME}"
fi
