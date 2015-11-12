#!/bin/bash

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [ -z $ELASTIC_IP ]; then
    ELASTIC_IP=192.168.50.1
fi

docker run \
    --name cilium-kibana \
    -e ELASTICSEARCH_URL=http://${ELASTIC_IP}:9200 \
    -p 5601:5601 \
    -d kibana:4.2.0

docker cp $dir/../external-deps/marvel-2.0.0.tar.gz cilium-kibana:/tmp
docker exec -ti cilium-kibana kibana plugin --install marvel --url file:///tmp/marvel-2.0.0.tar.gz
docker restart cilium-kibana
