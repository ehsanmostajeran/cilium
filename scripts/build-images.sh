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

cd ./debug-shell/
docker build -t debug_shell:latest .
cd ../
docker save -o ./images/debug_shell.ditar debug_shell:latest
chown $(whoami) ./images/debug_shell.ditar

$dir/build-cilium-image.sh

echo "Images successfully saved"

