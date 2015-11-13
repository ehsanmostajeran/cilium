#!/usr/bin/env bash
set -e
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

image_tag="cilium-binary-image"

cd "${dir}/.."
mkdir -p ./cilium-binary
rm -fr ./cilium-binary/Godeps
rm -fr ./cilium-binary/cilium

cp Dockerfile.binary ./cilium-binary/Dockerfile
cp -r ./Godeps ./cilium-binary/Godeps
cp -r ./cilium ./cilium-binary/cilium

cd ./cilium-binary

docker build -t "cilium/${image_tag}" .
docker_ID=$(docker create cilium/cilium-binary-image)
docker cp "${docker_ID}":/go/src/github.com/cilium-team/cilium/cilium-Linux-x86_64 ./
docker rm -f "${docker_ID}"
docker rmi "cilium/${image_tag}"

cp ./cilium-Linux-x86_64 ../cilium-Linux-x86_64

cd ..
rm -fr ./cilium-binary

echo "Cilium binary successfully created"

exit 0
