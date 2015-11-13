#!/usr/bin/env bash
set -e
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

cd "${dir}/.."
vagrant ssh -c 'cd cilium/scripts/kubernetes/node1; . ./pre-setup.sh; cp ../master.sh ./ && sudo -E ./master.sh' node1
vagrant ssh -c 'cd cilium/scripts/kubernetes/node2; . ./pre-setup.sh; cp ../worker.sh ./ && sudo -E ./worker.sh' node2
echo "Kubernetes dev cluster ready!"

exit 0
