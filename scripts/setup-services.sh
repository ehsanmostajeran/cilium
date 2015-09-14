#!/bin/bash
set -e
dir=`dirname $0`

dnsID=$($dir/setup-dns.sh)
while true
do
	echo "Waiting for DNS container to be ready"
	running=$(DOCKER_HOST=localhost:2375 docker inspect --format='{{.State.Running}}' $dnsID)
	paused=$(DOCKER_HOST=localhost:2375 docker inspect --format='{{.State.Paused}}' $dnsID)
	restarting=$(DOCKER_HOST=localhost:2375 docker inspect --format='{{.State.Restarting}}' $dnsID)
	if [ "$running" == "true" ] && [ $paused == "false" ] && [ $restarting == "false" ]
	then
		break
	else
		DOCKER_HOST=localhost:2375 docker restart $dnsID
		sleep 1
	fi
done
#sleep 1
$dir/setup-haproxy.sh
