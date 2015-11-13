#!/usr/bin/env bash
set -e
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

cd "${dir}/.."

if [ -z "${ELASTIC_IP}" ]; then
    ELASTIC_IP="192.168.50.1"
fi

export ELASTIC_IP

./cilium-Linux-x86_64 -l debug -F

temp_dir=$(mktemp -d)
if [ -z "${DOCKER_IP}" ]; then
    DOCKER_IP=$(ip -f inet -o addr show docker0|cut -d\  -f 7 | cut -d/ -f 1)
    if [ -z "${DOCKER_IP}" ]; then
        >&2 echo "Unable to retrive DOCKER_IP, you can set it manually"
    fi
fi

cp -r ./policy/. "${temp_dir}"

sed -i "s/172.17.0.1/${DOCKER_IP}/g" "${temp_dir}/1-cluster.yml"

./cilium-Linux-x86_64 -l debug -f "${temp_dir}"
./cilium-Linux-x86_64 -l debug -f  ./examples/compose/app-policy.yml
./cilium-Linux-x86_64 -l debug -f  ./examples/kubernetes/policy/app-policy.yml

exit 0
