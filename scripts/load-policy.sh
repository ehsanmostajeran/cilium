#!/bin/bash

dir="$(dirname "$0")"
export ELASTIC_IP=192.168.50.5
$dir/../cilium-Linux-x86_64 -l debug -F
temp_dir=$(mktemp -d)
docker_IP=$(ip -f inet -o addr show docker0|cut -d\  -f 7 | cut -d/ -f 1)
cp -R "$dir/../policy/." "$temp_dir"
sed -i "s/172.17.0.1/$docker_IP/g" "$temp_dir/1-cluster.yml"
$dir/../cilium-Linux-x86_64 -f "$temp_dir" -l debug
$dir/../cilium-Linux-x86_64 -f  $dir/../examples/compose/app-policy.yml -l debug
$dir/../cilium-Linux-x86_64 -f  $dir/../examples/kubernetes/policy/app-policy.yml -l debug
