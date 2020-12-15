set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"

IMAGE_NAME="kurtosistech/example-microservice"

git_branch="$(git rev-parse --abbrev-ref HEAD)"
docker_tag="$(echo "${git_branch}" | sed 's,[/:],_,g')"

docker build -t "${SUITE_IMAGE}:${docker_tag}" -f "${script_dirpath}/testsuite/Dockerfile" "${script_dirpath}"
