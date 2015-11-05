#!/bin/bash
docker_IP=$(ip -f inet -o addr show docker0|cut -d\  -f 7 | cut -d/ -f 1)
docker run \
    --name cilium-kibana \
    -e ELASTICSEARCH_URL=http://${docker_IP}:9200 \
    -p 5601:5601 \
    -d kibana:4.1.1
