#!/usr/bin/env bash
set -e
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

image_tag="cilium:latest"

cd "${dir}/.."

echo "... cilium.ditar"
set +e
docker rmi --no-prune "cilium/${image_tag}"
set -e
"${dir}/build-cilium-dev-image.sh"

mkdir -p ./images
echo "Saving cilium image to 'images' directory"
docker save -o ./images/cilium.ditar "cilium/${image_tag}"
chown $(whoami) ./images/cilium.ditar

exit 0
