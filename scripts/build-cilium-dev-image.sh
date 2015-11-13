#!/usr/bin/env bash
set -e
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

image_tag="cilium:latest"

cd "${dir}/.."

mkdir -p ./cilium-dev

cp Dockerfile.dev ./cilium-dev/Dockerfile
cp cilium-Linux-x86_64 ./cilium-dev/
cp -r backend ./cilium-dev/backend

cd ./cilium-dev

docker build -t "cilium/${image_tag}" .

cd ..
rm -fr ./cilium-dev

echo "Cilium development image successfully created"

exit 0
