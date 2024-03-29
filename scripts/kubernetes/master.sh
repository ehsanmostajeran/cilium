#!/bin/bash

# Copyright 2015 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# A scripts to install k8s worker node.
# Author @wizard_cxy @reouser

set -e

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# See if there's a different CILIUM_ROOT path set
if [ -z ${CILIUM_ROOT} ]; then
    CILIUM_ROOT=$( cd "$( dirname "${dir}/../../../../" )" && pwd )
    echo "CILIUM_ROOT not set, using default: ${CILIUM_ROOT}"
else
    echo "CILIUM_ROOT is set to: ${CILIUM_ROOT}"
fi

# See if there's a different DOCKER_ENDPOINT set
if [ -z ${DOCKER_ENDPOINT} ]; then
    DOCKER_ENDPOINT="unix:///var/run/docker.sock"
    echo "DOCKER_ENDPOINT not set, using default: ${DOCKER_ENDPOINT}"
else
    echo "Docker endpoint is set to: ${DOCKER_ENDPOINT}"
fi

# Make sure docker daemon is running
if ( ! ps -ef | grep "/usr/bin/docker" | grep -v 'grep' &> /dev/null ); then
    echo "Docker is not running on this machine!"
    exit 1
fi

# Make sure k8s version env is properly set
if [ -z ${K8S_VERSION} ]; then
    K8S_VERSION="1.0.7"
    echo "K8S_VERSION is not set, using default: ${K8S_VERSION}"
else
    echo "k8s version is set to: ${K8S_VERSION}"
fi

# Run as root
if [ "$(id -u)" != "0" ]; then
    echo >&2 "Please run as root"
    exit 1
fi

# Check if a command is valid
command_exists() {
    command -v "$@" > /dev/null 2>&1
}

lsb_dist=""

# Detect the OS distro, we support ubuntu, debian, mint, centos, fedora dist
detect_lsb() {
    case "$(uname -m)" in
        *64)
            ;;
         *)
            echo "Error: We currently only support 64-bit platforms."       
            exit 1
            ;;
    esac

    if command_exists lsb_release; then
        lsb_dist="$(lsb_release -si)"
    fi
    if [ -z ${lsb_dist} ] && [ -r /etc/lsb-release ]; then
        lsb_dist="$(. /etc/lsb-release && echo "$DISTRIB_ID")"
    fi
    if [ -z ${lsb_dist} ] && [ -r /etc/debian_version ]; then
        lsb_dist='debian'
    fi
    if [ -z ${lsb_dist} ] && [ -r /etc/fedora-release ]; then
        lsb_dist='fedora'
    fi
    if [ -z ${lsb_dist} ] && [ -r /etc/os-release ]; then
        lsb_dist="$(. /etc/os-release && echo "$ID")"
    fi

    lsb_dist="$(echo ${lsb_dist} | tr '[:upper:]' '[:lower:]')"

    case "${lsb_dist}" in
        amzn|centos|debian|ubuntu)
            ;;
        *)
            echo "Error: We currently only support ubuntu|debian|amzn|centos."
            exit 1
            ;;
    esac
}


# Start the bootstrap daemon
bootstrap_daemon() {
    sudo -b docker -d -H unix:///var/run/docker-bootstrap.sock -p /var/run/docker-bootstrap.pid --iptables=false --ip-masq=false --bridge=none --graph=/var/lib/docker-bootstrap 2> /var/log/docker-bootstrap.log 1> /dev/null
    
    sleep 5
}

# Start k8s components in containers
DOCKER_CONF=""

start_k8s(){
    if [ -f ${CILIUM_ROOT}/images/etcd.ditar ]; then
        sudo docker -H unix:///var/run/docker-bootstrap.sock load -i ${CILIUM_ROOT}/images/etcd.ditar
    fi
    if [ -f ${CILIUM_ROOT}/images/flannel.ditar ]; then
        sudo docker -H unix:///var/run/docker-bootstrap.sock load -i ${CILIUM_ROOT}/images/flannel.ditar
    fi
    # Start etcd 
    docker -H unix:///var/run/docker-bootstrap.sock run --restart=always --net=host -d gcr.io/google_containers/etcd:2.0.13 /usr/local/bin/etcd --addr=127.0.0.1:4001 --bind-addr=0.0.0.0:4001 --data-dir=/var/etcd/data

    sleep 5
    # Set flannel net config
    docker -H unix:///var/run/docker-bootstrap.sock run --net=host gcr.io/google_containers/etcd:2.0.13 etcdctl set /coreos.com/network/config '{ "Network": "10.1.0.0/16", "Backend": {"Type": "vxlan"}}'

    # iface may change to a private network interface, eth0 is for default
    flannelCID=$(docker -H unix:///var/run/docker-bootstrap.sock run --restart=always -d --net=host --privileged -v /dev/net:/dev/net quay.io/coreos/flannel:0.5.3 /opt/bin/flanneld -iface="eth0")

    sleep 8

    # Copy flannel env out and source it on the host
    docker -H unix:///var/run/docker-bootstrap.sock cp ${flannelCID}:/run/flannel/subnet.env .
    source subnet.env

    # Configure docker net settings, then restart it
    case "${lsb_dist}" in
        amzn)
            DOCKER_CONF="/etc/sysconfig/docker"
            echo "OPTIONS=\"\$OPTIONS --mtu=${FLANNEL_MTU} --bip=${FLANNEL_SUBNET}\"" | sudo tee -a ${DOCKER_CONF}
            ifconfig docker0 down
            yum -y -q install bridge-utils && brctl delbr docker0 && service docker restart
            ;;
        centos)
            DOCKER_CONF="/etc/sysconfig/docker"
            echo "OPTIONS=\"\$OPTIONS --mtu=${FLANNEL_MTU} --bip=${FLANNEL_SUBNET}\"" | sudo tee -a ${DOCKER_CONF}
            if ! command_exists ifconfig; then
                yum -y -q install net-tools
            fi
            ifconfig docker0 down
            yum -y -q install bridge-utils && brctl delbr docker0 && systemctl restart docker
            ;;
        ubuntu|debian)
            DOCKER_CONF="/etc/default/docker"
            echo "DOCKER_OPTS=\"\$DOCKER_OPTS --mtu=${FLANNEL_MTU} --bip=${FLANNEL_SUBNET}\"" | sudo tee -a ${DOCKER_CONF}
            ifconfig docker0 down
            apt-get install bridge-utils
            brctl delbr docker0
            service docker stop
            while [ `ps aux | grep /usr/bin/docker | grep -v grep | wc -l` -gt 0 ]; do
                echo "Waiting for docker to terminate"
                sleep 1
            done
            service docker start
            ;;
        *)
            echo "Unsupported operations system ${lsb_dist}"
            exit 1
            ;;
    esac

    # sleep a little bit
    sleep 5

    # Start cilium
    ${CILIUM_ROOT}/entrypoint.sh infect

    # Start kubelet & proxy, then start master components as pods
    docker run \
        --net=host \
        --pid=host \
        --privileged \
        --restart=always \
        -d \
        -v /sys:/sys:ro \
        -v /var/run:/var/run:rw \
        -v /:/rootfs:ro \
        -v /dev:/dev \
        -v /var/lib/docker/:/var/lib/docker:rw \
        -v /var/lib/kubelet/:/var/lib/kubelet:rw \
        gcr.io/google_containers/hyperkube:v${K8S_VERSION} \
        /hyperkube kubelet \
        --v=2 --address=0.0.0.0 --enable-server \
        --config=/etc/kubernetes/manifests-multi \
        --cluster-dns=10.0.0.10 \
        --cluster-domain=cluster.local \
        --containerized \
        --docker-endpoint=${DOCKER_ENDPOINT}

    docker run \
        -d \
        --net=host \
        --privileged \
        gcr.io/google_containers/hyperkube:v${K8S_VERSION} \
        /hyperkube proxy --master=http://127.0.0.1:8080 --v=2   
}

echo "Detecting your OS distro ..."
detect_lsb

echo "Starting bootstrap docker ..."
bootstrap_daemon

echo "Starting k8s ..."
start_k8s

echo "Master done!"
