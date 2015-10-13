#!/bin/bash

echo "Setting up HA-Proxy..."
swarm-master docker run \
    -d -p 10001:10001 -p 1988:1988 \
    -p 5000:5000 \
    --name cilium-loadbalancer \
    -l "com.intent.service=svc_loadbalancer" \
    tnolet/haproxy-rest
