#!/bin/bash
set -e
dir=`dirname $0`

cd $dir/..
vagrant ssh -c 'cd cilium/scripts/kubernetes/node1; . ./pre-setup.sh; sudo -E ../master.sh' node1
vagrant ssh -c 'cd cilium/scripts/kubernetes/node2; . ./pre-setup.sh; sudo -E ../worker.sh' node2
echo "Kubernetes dev cluster ready!"
