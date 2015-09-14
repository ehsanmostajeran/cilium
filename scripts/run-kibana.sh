#!/bin/bash
docker run \
    --name cilium-kibana \
    -e ELASTICSEARCH_URL=http://172.17.42.1:9200 \
    -p 5601:5601 \
    -d kibana:4.1.1
