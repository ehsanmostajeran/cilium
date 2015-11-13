#!/usr/bin/env bash
set -e
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

cd "${dir}/.."

vagrant snapshot go node1 all-installed
vagrant snapshot go node2 all-installed

vagrant ssh -c "cd cilium; docker load -i ./images/cilium.ditar" node1
vagrant ssh -c "cd cilium; docker load -i ./images/cilium.ditar" node2

vagrant snapshot delete node1 all-installed
vagrant snapshot delete node2 all-installed

vagrant snapshot take node1 all-installed
vagrant snapshot take node2 all-installed

exit 0
