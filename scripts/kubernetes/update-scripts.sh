#!/bin/env bash

curl -LSsl -o master.sh.orig https://raw.githubusercontent.com/kubernetes/kubernetes/master/docs/getting-started-guides/docker-multinode/master.sh
curl -LSsl -o worker.sh.orig https://raw.githubusercontent.com/kubernetes/kubernetes/master/docs/getting-started-guides/docker-multinode/worker.sh

patch master.sh.orig master.sh.patch
mv master.sh.orig master.sh
chmod 700 master.sh
patch worker.sh.orig worker.sh.patch
mv worker.sh.orig worker.sh
chmod 700 worker.sh
