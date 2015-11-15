#!/usr/bin/env bash
set -e
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [ -z "${ELASTIC_IP}" ]; then
    ELASTIC_IP=192.168.50.1
fi

str_search="${1}"

size=100
if [ "${2}" != "" ]; then
    size="${2}"
fi

echo "Logs results:"
curl -XGET "http://${ELASTIC_IP}:9200/cilium-log*/_search?q=${str_search}&pretty&size=${size}"
echo ""
echo "State results:"
curl -XGET "http://${ELASTIC_IP}:9200/cilium-state/_search?q=${str_search}&pretty&size=${size}"
echo ""
echo "Configs results:"
curl -XGET "http://${ELASTIC_IP}:9200/cilium-configs/_search?q=${str_search}&pretty&size=${size}"

exit 0
