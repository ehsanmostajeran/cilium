#!/bin/bash
set -e
dir=`dirname $0`

echo "... cilium.ditar"
set +e
docker rmi --no-prune cilium/cilium
set -e
docker build -t cilium/cilium .
mkdir -p ./images
docker save -o ./images/cilium.ditar cilium/cilium:latest
chown $(whoami) ./images/cilium.ditar
