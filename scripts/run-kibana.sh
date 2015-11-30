#!/usr/bin/env bash
set -e
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

marvel="marvel-2.1.0.tar.gz"
kibana_tag="4.3.0"

cd "${dir}/.."

if [ -z ${ELASTIC_IP} ]; then
    ELASTIC_IP="192.168.50.1"
fi

docker run \
       --name "cilium-kibana" \
       -e ELASTICSEARCH_URL="http://${ELASTIC_IP}:9200" \
       -p 5601:5601 \
       -d \
       "kibana:${kibana_tag}"

docker cp \
       "./external-deps/${marvel}" \
       cilium-kibana:/tmp

sleep 2s

docker exec \
       -ti \
       cilium-kibana \
       kibana plugin --install marvel --url "file:///tmp/${marvel}"

docker restart \
       cilium-kibana

exit 0
