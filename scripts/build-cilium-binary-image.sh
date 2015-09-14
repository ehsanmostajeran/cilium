#!/bin/bash
set -e
dir=`dirname $0`

cd $dir/../
mkdir -p ./cilium-binary
rm -fr ./cilium-binary/Godeps
rm -fr ./cilium-binary/scripts
rm -fr ./cilium-binary/cilium

cp Dockerfile.binary ./cilium-binary/Dockerfile
cp -r ./Godeps/ ./cilium-binary/Godeps
cp -r ./scripts/ ./cilium-binary/scripts
cp -r ./cilium/ ./cilium-binary/cilium

cd ./cilium-binary/

docker build -t cilium/cilium-binary-image .
ID=$(docker create cilium/cilium-binary-image)
docker cp $ID:/go/src/github.com/cilium-team/cilium/cilium-Linux-x86_64 ./
docker rm -f $ID
docker rmi cilium/cilium-binary-image

cp ./cilium-Linux-x86_64 ../cilium-Linux-x86_64

cd ../
rm -fr ./cilium-binary/

echo "Cilium binary created successfully"
