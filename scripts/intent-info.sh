#!/usr/bin/env bash
set -e
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [ -z "${ELASTIC_IP}" ]; then
    ELASTIC_IP=192.168.50.1
fi

curl -XGET "http://${ELASTIC_IP}:9200/_search?q=${1}&pretty"

exit 0
