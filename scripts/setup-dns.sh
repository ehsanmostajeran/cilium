#!/usr/bin/env bash

>&2 echo "Setting up DNS..."

swarm-master docker run \
	     --name "cilium-dns" \
	     -d \
	     -p 80:80 -p 53:53/udp \
	     -l "com.intent.service=svc_dns" \
	     cilium/docker-dns-rest:1.0-rr-with-del

exit 0
