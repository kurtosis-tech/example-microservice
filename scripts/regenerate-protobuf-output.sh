# This script regenerates Go bindings corresponding to the .proto files that define the API container's API
# It requires the Golang Protobuf extension to the 'protoc' compiler, as well as the Golang gRPC extension

set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

# NOTE: Relies on Kurtosis devtools being installed
rpc_api_dirpath="${root_dirpath}/kurtosis-module/rpc_api"
if ! GO_MOD_FILEPATH="${root_dirpath}/go.mod" generate-protobuf-bindings.sh "${rpc_api_dirpath}" "${rpc_api_dirpath}/bindings" "golang"; then
    echo "Error: Could not generate Go bindings for the Kurtosis module Protobuf API" >&2
    exit 1
fi
