#!/usr/bin/env bash
set -e
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

cd "${dir}/.."

mkdir -p ./images

docker_images=(\
"cilium-powerstrip-latest.ditar" \
cilium/powerstrip:latest \
"redis-latest.ditar" \
redis:latest \
"swarm-1.0.0.ditar" \
swarm:1.0.0 \
"elasticsearch-2.1.0.ditar" \
elasticsearch:2.1.0 \
"kibana-4.3.0.ditar" \
kibana:4.3.0 \
"logstash-2.1.0.ditar" \
logstash:2.1.0 \
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
gcr.io/google_containers/etcd:2.0.13 \
"hyperkube.ditar" \
gcr.io/google_containers/hyperkube:v1.0.7 \
"flannel.ditar" \
quay.io/coreos/flannel:0.5.3 \
"gb-redisslave.ditar" \
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

"${dir}/build-cilium-image.sh"

echo "Images successfully saved"

exit 0
