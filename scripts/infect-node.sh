#!/usr/bin/env bash
set -e
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

swarm_tag="1.0.0"
es_version="2.1.0"

cilium_port="8081"
consul_port="8500"
k8s_port="8080"
logstash_port="8080"
pwr_bef_daemon_port="2371"
pwr_bef_kub_master="8083"
pwr_bef_swarm_port="2375"
swarm_master_port="2373"

is_master="${NET_IP}"
workdir="$(mktemp -d)/cilium"
consul_dir="${workdir}/consul"
pwrstrip_dir="${workdir}/powerstrip/swarm/local"
logstash_dir="${workdir}/logstash"
mkdir -p "${pwrstrip_dir}"
mkdir -p "${consul_dir}"
mkdir -p "${logstash_dir}"
chmod a+rx "${logstash_dir}"
adapter_file_docker_swarm="${pwrstrip_dir}/adapters-local-1.yml"
adapter_file_docker_daemon="${pwrstrip_dir}/adapters-local-2.yml"
adapter_file_kubernetes="${pwrstrip_dir}/adapters-local-3.yml"
logstash_cilium_conf_file="${logstash_dir}/logstash-cilium.conf"
logstash_dcolle_conf_file="${logstash_dir}/logstash-docker-collector.conf"
configs_dir=$(cd "${dir}/../external-deps/docker-collector-configs" && pwd)

list_IPs() {
    local localip interface ip
    local ifs=() ips=();

    while read -r interface ip; do
        ifs+=( "${interface}" )  ips+=( "${ip}" );
    done < <(ip -o -4 addr | awk '!/^[0-9]*: ?lo|link\/ether/ {gsub("/", " "); print $2" "$4}')

    echo "You have the following addresses in your machine:"
    for i in "${!ifs[@]}"; do
        printf ' * %5s %s\n' "${ifs[i]}" "${ips[i]}";
    done

    read -ep 'Which IP address do you want to use? ' -i "${ips[0]}" localip
    eval "${1}=${localip}"
}

[ -z "${IP}" ] && {
    list_IPs IP
}

"${dir}/../backend/setup.sh"

# Local consul instance
start_consul() {
    docker run \
           -d \
           --name "cilium-consul" \
           -v "${consul_dir}:/data" \
           -p "${IP}:8300:8300" \
           -p "${IP}:8301:8301" \
           -p "${IP}:8301:8301/udp" \
           -p "${IP}:8302:8302" \
           -p "${IP}:8302:8302/udp" \
           -p "${IP}:8400:8400" \
           -p "${IP}:${consul_port}:8500" \
           progrium/consul \
           -server -advertise "${IP}" "${@}"
}

# Local elastic instance
start_elasticsearch() {
    docker run \
           -d \
           --name "cilium-elastic" \
           -p 9200:9200 \
           -p 9300:9300 \
           --net "host" \
           -l "com.intent.service=gov_db" \
           -l "com.intent.logical-name=cilium-elastic" \
           -e ES_HEAP_SIZE=1g \
           elasticsearch:${es_version} \
           elasticsearch \
           -Des.cluster.name="cilium-elastic" \
           -Des.network.bind_host="${IP}" \
           -Des.transport.publish_host="${IP}" \
           -Des.http.publish_host="${IP}" "${@}"

    docker cp \
           "${dir}/../external-deps/license-${es_version}.zip" cilium-elastic:/tmp

    docker exec \
           cilium-elastic \
           /usr/share/elasticsearch/bin/plugin install \
           "file:///tmp/license-${es_version}.zip"

    docker cp \
           "${dir}/../external-deps/marvel-agent-${es_version}.zip" cilium-elastic:/tmp

    docker exec \
           cilium-elastic \
           /usr/share/elasticsearch/bin/plugin install \
           "file:///tmp/marvel-agent-${es_version}.zip"

    docker restart cilium-elastic
}

# Local cilium instance
start_cilium() {
    docker run \
           -d \
           --name "cilium" \
           --net "host" \
           --pid "host" \
           --privileged \
           -v /tmp/cilium-logs \
           -v /var/run/openvswitch/db.sock:/var/run/openvswitch/db.sock \
           -v /var/run/openvswitch/lxc-br0.mgmt:/var/run/openvswitch/lxc-br0.mgmt \
           -e HOST_IP="${IP}" \
           -e ELASTIC_IP="${IP}" \
           -e DOCKER_HOST="tcp://${IP}:${swarm_master_port}" \
           cilium/cilium \
           -l=debug -e=false -P "${cilium_port}"
}

create_logstash_conf_files() {

    cat > "${logstash_cilium_conf_file}" << EOF
input {
  file {
    type => "cilium-log"
    path => "/tmp/cilium-logs/cilium-log-*.log"
    start_position => beginning
  }
}

filter {
  if [type] == "cilium-log" {
    grok {
      match => [ "message", "\A%{TIMESTAMP_ISO8601:timestamp}%{SPACE}%{HOSTNAME:node}%{SPACE}%{WORD}%{SPACE}%{NOTSPACE}%{SPACE}%{WORD:level}%{SPACE}%{WORD}%{SPACE}%{GREEDYDATA:logMessage}" ]
    }
    date {
      match => [ "timestamp", "ISO8601" ]
      target => [ "@timestamp" ]
    }
    mutate {
      remove_field => [ "message", "timestamp", "host", "tags", "path", "_score", "_type"]
    }
  }
}

output {
  if [type] == "cilium-log" {
    elasticsearch {
      hosts => ["${IP}:9200"]
      index => "cilium-log-%{+YYYY-MM-dd}"
    }
  }
}
EOF

chmod a+r "${logstash_cilium_conf_file}"

    cat > "${logstash_dcolle_conf_file}" << EOF
input {
  tcp {
    type => "cilium-docker-collector"
    port => "${logstash_port}"
  }
}

filter {
  if [type] == "cilium-docker-collector" {
    json {
      source => "message"
    }
    date {
      match => [ "UpdatedAt", "ISO8601" ]
      target => [ "@timestamp" ]
    }
    mutate {
      remove_field => [ "message", "host", "tags", "path", "_score", "_type"]
    }
  }
}

output {
  if [type] == "cilium-docker-collector" {
    elasticsearch {
      hosts => ["${IP}:9200"]
      index => "cilium-docker-collector-%{+YYYY-MM-dd}"
    }
  }
}
EOF

chmod a+r "${logstash_dcolle_conf_file}"

}

# Start logstash
start_logstash() {
    docker run \
           -d \
           --name "cilium-logstash" \
           --volumes-from "cilium" \
           -v "${logstash_dir}:/config" \
           logstash:${es_version} \
           -f /config

    sleep 2s
}

# Create adapter files
create_adapter_files() {

    cat > "${adapter_file_docker_swarm}" << EOF
version: 1
endpoints:
  "POST /containers/create":
    pre: [cilium]
  "POST /*/containers/create":
    pre: [cilium]
adapters:
  cilium: http://${IP}:${cilium_port}/docker/swarm/cilium-adapter
EOF

    cat > "${adapter_file_docker_daemon}" << EOF
version: 1
endpoints:
  "POST /containers/*/start":
    post: [cilium]
  "POST /containers/*/restart":
    post: [cilium]
  "POST /containers/create":
    pre: [cilium]
  "POST /*/containers/*/start":
    post: [cilium]
  "POST /*/containers/*/restart":
    post: [cilium]
  "POST /*/containers/create":
    pre: [cilium]
adapters:
  cilium: http://${IP}:${cilium_port}/docker/daemon/cilium-adapter
EOF

    cat > "${adapter_file_kubernetes}" << EOF
version: 1
endpoints:
  "POST /api/v1/namespaces/*/pods":
    pre: [cilium]
  "POST /api/v1/namespaces/*/services":
    pre: [cilium]
  "POST /api/v1/namespaces/*/replicationcontrollers":
    pre: [cilium]
adapters:
  cilium: http://${IP}:${cilium_port}/kubernetes/master/cilium-adapter
EOF

}

# Local powerstrip before s
start_powerstrip() {
    if [ -z "${KUBERNETES}" ]; then
        docker run \
               -d \
               --name "cilium-powerstrip-pre-swarm" \
               -e DOCKER_HOST="${IP}:${swarm_master_port}" \
               -v "${adapter_file_docker_swarm}:/etc/powerstrip/adapters.yml" \
               -v /var/run/docker.sock:/var/run/docker.sock \
               -p "${pwr_bef_swarm_port}:2375" \
               cilium/powerstrip:latest
    else
        if [ -n "${is_master}" ]; then
            docker run \
                   -d \
                   --name "cilium-powerstrip-pre-k8s-master" \
                   -e KUBE_SERVER="tcp://${IP}:${k8s_port}" \
                   -v "${adapter_file_kubernetes}:/etc/powerstrip/adapters.yml" \
                   -v /var/run/docker.sock:/var/run/docker.sock \
                   -p "${pwr_bef_kub_master}:8080" \
                   cilium/powerstrip:kubernetes
        fi
        docker run \
               -d \
               --name "cilium-powerstrip-pre-pwr-daemon" \
               -e DOCKER_HOST="${IP}:${pwr_bef_daemon_port}" \
               -v "${adapter_file_docker_swarm}:/etc/powerstrip/adapters.yml" \
               -v /var/run/docker.sock:/var/run/docker.sock \
               -p "${pwr_bef_swarm_port}:2375" \
               cilium/powerstrip:latest
    fi

    # powerstrip before docker daemon
    docker run \
           -d \
           --name "cilium-powerstrip-pre-daemon" \
           -v "${adapter_file_docker_daemon}:/etc/powerstrip/adapters.yml" \
           -v /var/run/docker.sock:/var/run/docker.sock \
           -p "${pwr_bef_daemon_port}:2375" \
           cilium/powerstrip:latest

    sleep 3s
}

start_swarm() {
    # Run a swarm agent on each node
    docker run \
           -d \
           --name "cilium-swarm-agent" \
           -l "com.intent.service=gov_swarm_events" \
           "swarm:${swarm_tag}" \
           join --advertise="${IP}:${pwr_bef_daemon_port}" \
           "consul://${IP}:${consul_port}/ciliumnodes"

    # Given swarm agent time to start up
    sleep 3s

    # Run a swarm master on each node
    docker run \
           -d \
           -p "${swarm_master_port}:2375" \
           --name "cilium-swarm-master" \
           "swarm:${swarm_tag}" \
           manage --replication \
           --advertise "${IP}:${swarm_master_port}" \
           "consul://${IP}:${consul_port}/ciliumnodes"

    # Given swarm master time to start up or we might end up without an event manager
    sleep 5s
}

# Local cilium swarm event handler
start_cilium_swarm_event_handler (){
    docker run \
           -d \
           --name "cilium-swarm-event-handler" \
           --net "host" \
           --pid "host" \
           --privileged \
           -e HOST_IP="${IP}" \
           -e ELASTIC_IP="${IP}" \
           -e DOCKER_HOST="tcp://${IP}:${swarm_master_port}" \
           -v /var/run/openvswitch/db.sock:/var/run/openvswitch/db.sock \
           -v /var/run/openvswitch/lxc-br0.mgmt:/var/run/openvswitch/lxc-br0.mgmt \
           cilium/cilium \
           -l=debug -o
}

# Local docker-collector instance
start_docker_collector() {
    docker run \
           -d \
           --name "cilium-docker-collector" \
           -h "$(hostname)" \
           --pid "host" \
           --privileged \
           -e ELASTIC_IP="${IP}" \
           -e LOGSTASH_IP="logstash" \
           -e LOGSTASH_PORT="${logstash_port}" \
           -v /var/run/docker.sock:/var/run/docker.sock \
           -v "${configs_dir}:/docker-collector/configs" \
           --link cilium-logstash:logstash \
           cilium/docker-collector:latest \
           -f '(k8s_.*)|(cilium.*)' \
           -i 'cilium-docker-collector' \
           -l debug
}

if [ -z "${MASTER_IP}" ]; then
    start_consul -bootstrap-expect 1
    start_elasticsearch
    MASTER_IP="${IP}"
else
    start_consul -join "${MASTER_IP}"
    start_elasticsearch -Des.discovery.zen.ping.unicast.hosts="${MASTER_IP}"
fi

start_cilium

create_logstash_conf_files

start_logstash

create_adapter_files

start_powerstrip

start_swarm

start_cilium_swarm_event_handler

start_docker_collector

echo "==========================================================================="
echo ""
echo " Successfully infected node ${IP}"
echo ""
echo " Master IP: ${MASTER_IP}"
echo ""
echo " Further nodes can be infected with: MASTER_IP=${MASTER_IP} infect"
echo ""
echo "==========================================================================="

exit 0
