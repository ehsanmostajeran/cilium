#!/bin/bash
set -e
dir=`dirname $0`

docker_images=(\
"cilium-powerstrip-latest.ditar" \
cilium/powerstrip:latest \
"redis-latest.ditar" \
redis:latest \
"swarm-0.4.0.ditar" \
swarm:0.4.0 \
"elasticsearch-1.7.1.ditar" \
elasticsearch:1.7.1 \
"kibana-4.1.1.ditar" \
kibana:4.1.1 \
"haproxy-rest.ditar" \
tnolet/haproxy-rest:latest \
"consul.ditar" \
progrium/consul:latest \
"cilium-dns.ditar" \
cilium/docker-dns-rest:1.0-rr-with-del \
"docker-collector.ditar" \
cilium/docker-collector:latest \
"compose-example.ditar" \
cilium/compose-example:latest \
"debug_shell.ditar" \
cilium/debug_shell:latest \
"powerstrip-kubernetes.ditar" \
cilium/powerstrip:kubernetes \
"etcd.ditar" \
gcr.io/google_containers/etcd:2.0.12 \
"hyperkube.ditar" \
gcr.io/google_containers/hyperkube:v1.0.3 \
"flannel.ditar" \
quay.io/coreos/flannel:0.5.3 \
"gb-redisslave.ditar"
gcr.io/google_samples/gb-redisslave:v1 \
)

echo "Pulling necessary images from DockerHub..."
for ((i=0; i<"${#docker_images[@]}"; i+=2)); do
    docker pull "${docker_images["$((i+1))"]}"
done

echo "Saving images from this machine as tar so they could be deployed on swarm nodes"
for ((i=0; i<"${#docker_images[@]}"; i+=2)); do
    echo "... ${docker_images["$((i))"]}"
    docker save -o "./images/${docker_images["$((i))"]}" "${docker_images["$((i+1))"]}"
    chown $(whoami) "./images/${docker_images["$((i))"]}"
done

$dir/build-cilium-image.sh

echo "Images successfully saved"

