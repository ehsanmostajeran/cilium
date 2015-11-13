#!/usr/bin/env bash

ip_regex="(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.)\
{3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])"

swarm-master() {

    info=$(DOCKER_HOST=127.0.0.1:2373 docker info)
    primary=$(echo "${info}" | grep -Eo "Primary: ${ip_regex}")

    if [[ -z "${primary}" ]]; then
        primary=$(echo "${info}" | grep "primary")
        if [[ -z "${primary}" ]]; then
            echo "Unable to find swarm master"
            exit -1
        fi
        export DOCKER_HOST="127.0.0.1:2375"
    else
        export DOCKER_HOST="$(echo "${primary}" | grep -oE "${ip_regex}"):2375"
    fi

    if [ "${1}" == "-q" ]; then
        shift
        quiet=true
    fi

    if [ -z "${quiet}" ]; then
        >&2 echo "Using DOCKER_HOST=${DOCKER_HOST}"
        >&2 echo ""
    fi

    ${@}

}

swarm-master ${@}
