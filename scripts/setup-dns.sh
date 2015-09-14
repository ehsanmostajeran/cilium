#!/bin/bash

DOCKER_HOST=localhost:2375 docker run \
	-d -p 80:80 -p 53:53/udp \
	--name cilium-dns \
	-l "com.intent.service=svc_dns" \
	cilium/docker-dns-rest:1.0-rr-with-del
