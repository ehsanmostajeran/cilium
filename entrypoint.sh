#!/usr/bin/env bash
set -e

address="https://raw.githubusercontent.com/cilium-team/cilium/master"

dependencies=( \
    "docker" \
	"docker --version" \
	"1.8.0" \
	"Open vSwitch" \
	"ovs-ofctl --version" \
	"2.3.2" \
)

show_help(){
    echo "Usage: entrypoint.sh <command>"
    echo ""
    echo "  check"
    echo "            checks all requirements for this node"
    echo "  prepare"
    echo "            downloads all images used by cilium node"
    echo "  infect"
    echo "            infects this host with cilium's components The IP"
    echo "            environment variable should be set to the reachable"
    echo "            node's IP. If set, the NET_IP environment variable should"
    echo "            be set to the network's address of the reachable node's"
    echo "            IP. If set, the MASTER_IP should point to one of the"
    echo "            already infected nodes. On the first node it"
    echo "            automatically starts 3rd party services for the cilium"
    echo "            cluster such as the loadbalancer and the DNS"
    echo "  infect-kubernetes"
    echo "            infects this host with cilium's components The IP"
    echo "            environment variable should be set to the reachable"
    echo "            node's IP. If set, the NET_IP environment variable should"
    echo "            be set to the network's address of the reachable node's"
    echo "            IP. If set, the MASTER_IP should point to master of the"
    echo "            kubernetes cluster."
    echo "  start-kibana"
    echo "            starts kibana container with dashboard The IP environment"
    echo "            variable should be set to the reachable node's IP"
}

check_version(){
    cmd="${1}"
    want="${2}"
    [[ $(eval "${cmd}") =~ ([0-9][.][0-9][.][0-9]*) ]] && version="${BASH_REMATCH[1]}"
    local got=$(echo -e "${version}\n${want}" | sed '/^$/d' | sort -nr | head -1)
    if [[ "${got}" = "${version}" ]]; then
        echo "OK!"
    else
        echo "ERROR: got ${version}, version ${want} or higher required"
    fi
}

check_requirements(){
    echo "Checking dependencies..."
    for ((i=0; i<"${#dependencies[@]}"; i+=3)); do
        res=$(check_version "${dependencies["$((i+1))"]}" "${dependencies["$((i+2))"]}")
        echo "${dependencies["$((i))"]}: $res"
        if [ "${res}" != "OK!" ]
        then
            echo "ERROR: Please install the right version of "${dependencies["$((i))"]}""
            eval "${1}=1"
            return
        fi
    done
    docker_IP=$(ip -f inet -o addr show docker0|cut -d\  -f 7 | cut -d/ -f 1)
    if [[ -z "${docker_IP}" ]]; then
        echo "ERROR: Unable to find docker0 IP address."
        echo "Are you sure you have docker daemon running?"
    fi

    echo "SUCCESS: All dependencies are available with the right version!"
    eval "${1}=0"
}

prepare(){
    docker_images=(\
        "cilium-powerstrip-latest.ditar" \
        cilium/powerstrip:latest \
        "swarm-1.0.0.ditar" \
        swarm:1.0.0 \
        "elasticsearch-2.1.0.ditar" \
        elasticsearch:2.1.0 \
        "haproxy-rest.ditar" \
        tnolet/haproxy-rest:latest \
        "consul.ditar" \
        progrium/consul:latest \
        "cilium-dns.ditar" \
        cilium/docker-dns-rest:1.0-rr-with-del \
        "docker-collector.ditar" \
        cilium/docker-collector:latest \
        "cilium.ditar" \
        cilium/cilium:latest \
    )

    echo "Pulling necessary images from DockerHub..."
    for ((i=0; i<"${#docker_images[@]}"; i+=2)); do
        docker pull "${docker_images["$((i+1))"]}"
    done
}

infect(){
    echo "Infecting node with cilium..."
    echo "MASTER_IP=${MASTER_IP}"
    echo "IP=${IP}"
    echo "Downloading scripts..."
    tmp_dir=$(mktemp -d)
    scripts_tmp="${tmp_dir}/scripts"
    backend_tmp="${tmp_dir}/backend"
    external_deps_temp="${tmp_dir}/external-deps"
    mkdir -p "${scripts_tmp}"
    mkdir -p "${backend_tmp}"
    mkdir -p "${external_deps_temp}/docker-collector-configs"
    curl -Ssl -o "${backend_tmp}/setup.sh" "${address}/backend/setup.sh"
    chmod +x "${backend_tmp}/setup.sh"
    curl -Ssl -o "${backend_tmp}/config.sh" "${address}/backend/config.sh"
    chmod +x "${backend_tmp}/config.sh"
    curl -Ssl -o "${backend_tmp}/utils.sh" "${address}/backend/utils.sh"
    chmod +x "${backend_tmp}/utils.sh"
    curl -Ssl -o "${scripts_tmp}/infect-node.sh" "${address}/scripts/infect-node.sh"
    chmod +x "${scripts_tmp}/infect-node.sh"
    curl -Ssl -o "${external_deps_temp}/marvel-agent-2.0.0.zip" "${address}/external-deps/marvel-agent-2.0.0.zip"
    curl -Ssl -o "${external_deps_temp}/license-2.0.0.zip" "${address}/external-deps/license-2.0.0.zip"
    curl -Ssl -o "${external_deps_temp}/docker-collector-configs/templates.json" "${address}/external-deps/docker-collector-configs/templates.json"
    curl -Ssl -o "${external_deps_temp}/docker-collector-configs/configs.json" "${address}/external-deps/docker-collector-configs/configs.json"
    echo "Downloads completed..."
    "${scripts_tmp}/infect-node.sh"
}

infect_kubernetes(){
    echo "Infecting node with cilium..."
    echo "MASTER_IP=${MASTER_IP}"
    echo "IP=${IP}"
    echo "Downloading scripts..."
    export KUBERNETES=true
    export DOCKER_ENDPOINT="tcp://${IP}:2375"
    tmp_dir="$(mktemp -d)"
    tmp_dir_cilium="${tmp_dir}/cilium"
    scripts_tmp="${tmp_dir_cilium}/scripts"
    scripts_k8s_tmp="${tmp_dir_cilium}/scripts/kubernetes"
    mkdir -p "${tmp_dir_cilium}"
    mkdir -p "${scripts_tmp}"
    mkdir -p "${scripts_k8s_tmp}"
    curl -Ssl -o "${tmp_dir}/entrypoint.sh" "${address}/entrypoint.sh"
    chmod +x "${tmp_dir}/entrypoint.sh"
    CILIUM_ROOT=${tmp_dir}
    if [[ -z "${MASTER_IP}" ]]; then
        curl -Ssl -o "${scripts_k8s_tmp}/master.sh" "${address}/scripts/kubernetes/master.sh"
        chmod +x "${scripts_k8s_tmp}/master.sh"
        echo "Downloads completed..."
        "${scripts_k8s_tmp}/master.sh"
    else
        curl -Ssl -o "${scripts_k8s_tmp}/worker.sh" "${address}/scripts/kubernetes/worker.sh"
        chmod +x "${scripts_k8s_tmp}/worker.sh"
        echo "Downloads completed..."
        "${scripts_k8s_tmp}/worker.sh"
    fi
}

start_services(){
    echo "Downloading scripts..."
    tmp_dir=$(mktemp -d)
    scripts_tmp="${tmp_dir}/scripts"
    mkdir -p "${scripts_tmp}"
    curl -Ssl -o "${scripts_tmp}/setup-haproxy.sh" "${address}/scripts/setup-haproxy.sh"
    chmod +x "${scripts_tmp}/setup-haproxy.sh"
    curl -Ssl -o "${scripts_tmp}/setup-dns.sh" "${address}/scripts/setup-dns.sh"
    chmod +x "${scripts_tmp}/setup-dns.sh"
    curl -Ssl -o "${scripts_tmp}/setup-services.sh" "${address}/scripts/setup-services.sh"
    chmod +x "${scripts_tmp}/setup-services.sh"
    echo "Downloads completed..."
    "${scripts_tmp}/setup-services.sh"
    echo "Done"
}

requote() { sed 's/[^\/]/&/g; s/\//\\\//g' <<< "${1}"; }

store_policy(){
    echo "Using network address (NET_IP): ${NET_IP}"
    echo "Downloading basic policies..."
    tmp_dir=$(mktemp -d)
    policies_tmp="${tmp_dir}/policy"
    mkdir -p "${policies_tmp}"
    curl -Ssl -o "${policies_tmp}/1-cluster.yml" "${address}/policy/1-cluster.yml"
    curl -Ssl -o "${policies_tmp}/2-dns-config.yml" "${address}/policy/2-dns-config.yml"
    curl -Ssl -o "${policies_tmp}/3-haproxy-config.yml" "${address}/policy/3-haproxy-config.yml"
    curl -Ssl -o "${policies_tmp}/4-debug-shell.yml" "${address}/policy/4-debug-shell.yml"
    net_ip=$(requote "${NET_IP}")
    sed -i "s/192.168.50.0\/24/${net_ip}/g" "${policies_tmp}/1-cluster.yml"
    docker_IP=$(ip -f inet -o addr show docker0|cut -d\  -f 7 | cut -d/ -f 1)
    sed -i "s/172.17.0.1/${docker_IP}/g" "${policies_tmp}/1-cluster.yml"
    docker run --rm \
           -e ELASTIC_IP="${IP}" \
           -v "${policies_tmp}":/opt/cilium/policies/ \
           cilium/cilium -f /opt/cilium/policies
}

start_kibana(){
    tmp_dir=$(mktemp -d)
    scripts_tmp="${tmp_dir}/scripts"
    backend_tmp="${tmp_dir}/backend"
    external_deps_temp="${tmp_dir}/external-deps"
    curl -Ssl -o "${external_deps_temp}/marvel-2.0.0.tar.gz" "${address}/external-deps/marvel-2.0.0.tar.gz"
    echo "Starting Kibana..."
    docker run \
           --name "cilium-kibana" \
           -e ELASTICSEARCH_URL="http://${IP}:9200" \
           -p 5601:5601 \
           -d kibana:4.2.0
    docker cp "${external_deps_temp}/marvel-2.0.0.tar.gz" cilium-kibana:/tmp
    docker exec -ti cilium-kibana kibana plugin --install marvel --url file:///tmp/marvel-2.0.0.tar.gz
    docker restart cilium-kibana
}

entry(){
    case "${1}" in
        check)
            check_requirements ret
            if [[ "${ret}" == 0 ]]; then
                exit 0
            else
                exit 1
            fi
            ;;
        prepare)
            prepare
            exit 0
            ;;
        infect)
            if [[ -z "${NET_IP}" ]] && [[ -z "${MASTER_IP}" ]]; then
                echo "ERROR: NET_IP or MASTER_IP is empty"
                show_help
                exit 0
            fi
            if [[ -z "${IP}" ]]; then
                echo "ERROR: IP is empty"
                show_help
                exit 0
            fi
            check_requirements ret
            if [[ "${ret}" == 0 ]]; then
                infect
                if [[ -n "${NET_IP}" ]]; then
                    store_policy
                    start_services
                    echo "First node infected!"
                else
                    echo "Node infected!"
                fi
                exit 0
            else
                exit 1
            fi
            show_help
            exit 0
            ;;
        infect-kubernetes)
            if [[ -z "${NET_IP}" ]] && [[ -z "${MASTER_IP}" ]]; then
                echo "ERROR: NET_IP or MASTER_IP is empty"
                show_help
                exit 0
            fi
            if [[ -z "${IP}" ]]; then
                echo "ERROR: IP is empty"
                show_help
                exit 0
            fi
            check_requirements ret
            if [[ "${ret}" == 0 ]]; then
                infect_kubernetes
                if [[ -n "${NET_IP}" ]]; then
                    echo "Master node infected!"
                else
                    echo "Worker node infected!"
                fi
                exit 0
            else
                exit 1
            fi
            show_help
            exit 0
            ;;
        start-kibana)
            if [[ -z "${IP}" ]]; then
                echo "ERROR: IP is empty"
                show_help
                exit 0
            fi
            start_kibana
            exit 0
            ;;
        *)
            show_help
            exit -1
            ;;
    esac
}

entry "${@}"
