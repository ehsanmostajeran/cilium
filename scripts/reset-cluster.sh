#!/bin/bash
set -e
dir=`dirname $0`

$dir/clean-containers.sh

cd $dir/..
vagrant ssh -c 'cd cilium; make clean-containers' node1
vagrant ssh -c 'cd cilium; make clean-containers' node2

sudo make IP=192.168.50.1 infect

vagrant ssh -c "cd cilium; sudo MASTER_IP=192.168.50.1 IP=192.168.50.5 make infect" node1
vagrant ssh -c "cd cilium; sudo MASTER_IP=192.168.50.5 IP=192.168.50.6 make infect" node2

export ELASTIC_IP=192.168.50.1

$dir/../cilium-Linux-x86_64 -l debug -D
