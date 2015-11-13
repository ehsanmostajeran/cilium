#!/usr/bin/env bash
set -e
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [ -z "${ELASTIC_IP}" ]; then
    ELASTIC_IP="192.168.50.1"
fi

endpoint=$(curl -s -XGET "http://${ELASTIC_IP}:9200/cilium-state/endpoint/_search?q=${@}")
group=$(echo "${endpoint}" | grep -e '"group":[0-9]*')

[ -z "${group}" ] &&  exit 1
echo "${group}" | sed 's/^.*"group":\([0-9]*\).*$/\1/'

exit 0
