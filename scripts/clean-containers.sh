#!/usr/bin/env bash

cont=( $(docker ps -aq --filter name=cilium) )
[ "${cont}" ] && {
    docker rm -f "${cont[@]}"
}

exit 0
