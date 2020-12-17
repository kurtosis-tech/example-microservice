set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"

API_IMAGE_NAME="kurtosistech/example-microservices_api"
DATASTORE_IMAGE_NAME="kurtosistech/example-microservices_datastore"

echo "Building API image..."
docker build -t "${API_IMAGE_NAME}" -f "${script_dirpath}/api/Dockerfile" "${script_dirpath}"
echo "API image built"

echo "Building datastore image..."
docker build -t "${DATASTORE_IMAGE_NAME}" -f "${script_dirpath}/datastore/Dockerfile" "${script_dirpath}"
echo "Datastore image built"
