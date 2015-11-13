#!/usr/bin/env bash
set -e
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

dnsID=$("${dir}/setup-dns.sh")
while true
do
    echo "Waiting for DNS container to be ready"
    running=$(swarm-master docker inspect --format='{{.State.Running}}' "${dnsID}")
    paused=$(swarm-master docker inspect --format='{{.State.Paused}}' "${dnsID}")
    restarting=$(swarm-master docker inspect --format='{{.State.Restarting}}' "${dnsID}")
    if [ "${running}" == "true" ] && [ "${paused}" == "false" ] && [ "${restarting}" == "false" ]
    then
        break
    else
        swarm-master docker restart "${dnsID}"
        sleep 1
    fi
done

"${dir}/setup-haproxy.sh"

exit 0
