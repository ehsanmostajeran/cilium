#!/bin/bash
set -e
dir=`dirname $0`

cd $dir/../
mkdir -p ./cilium-dev

cp Dockerfile.dev ./cilium-dev/Dockerfile
cp cilium-Linux-x86_64 ./cilium-dev/
cp -r backend ./cilium-dev/backend

cd "./cilium-dev"

docker build -t cilium/cilium .

cd ../
rm -fr ./cilium-dev

echo "Cilium development image successfully created"
