#!/bin/bash
docker_file=/etc/default/docker
docker rm -f `docker ps -aq`
service docker stop
docker_opts=$(tail -n 1 $docker_file)
docker_opts=$(echo $docker_opts | grep "bip")

if [ -n "$docker_opts" ]; then
     echo "Removing docker options from file"
     head -n -1 $docker_file > ${docker_file}.tmp ; mv ${docker_file}.tmp ${docker_file}
fi
killall -9 docker
service docker start
echo "Clean up complete"
