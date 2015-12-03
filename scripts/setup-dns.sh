#!/usr/bin/env bash

>&2 echo "Setting up DNS..."

swarm-master -q docker run \
	     --name "cilium-dns" \
	     -d \
	     -p 80:80 -p 53:53/udp \
	     -l "com.intent.service=svc_dns" \
	     cilium/docker-dns-rest:latest

sleep 3s

exit 0
