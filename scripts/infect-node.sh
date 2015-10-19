#!/usr/bin/env bash
set -e
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

SWARM_VERSION="0.4.0"
PWR_BEF_DAEMON_PORT="2371"
SWARM_MASTER_PORT="2373"
PWR_BEF_SWARM_PORT="2375"
CONSUL_PORT="8500"
PWR_BEF_KUB_MASTER="8083"
CILIUM_PORT="8081"
K8S_PORT="8080"
IS_MASTER=$NET_IP
WORKDIR="$(mktemp -d)/cilium"
CONSUL_DIR="$WORKDIR/consul"
POWRSTRIP_DIR="$WORKDIR/powerstrip/swarm/local"
mkdir -p "$POWRSTRIP_DIR"
mkdir -p "$CONSUL_DIR"
ADAPTER_FILE1="$POWRSTRIP_DIR/adapters-local-1.yml"
ADAPTER_FILE2="$POWRSTRIP_DIR/adapters-local-2.yml"
ADAPTER_FILE3="$POWRSTRIP_DIR/adapters-local-3.yml"

function listIPs() {
	local __localip __interface __ip
	local __ifs=() __ips=();

	while read -r __interface __ip; do
		__ifs+=( "$__interface" )  __ips+=( "$__ip" );
	done < <(ip -o -4 addr | awk '!/^[0-9]*: ?lo|link\/ether/ {gsub("/", " "); print $2" "$4}')

	echo "You have the following addresses in your machine:"
	for i in "${!__ifs[@]}"; do
		printf ' * %5s %s\n' "${__ifs[i]}" "${__ips[i]}";
	done

	read -ep 'Which IP address do you want to use? ' -i "${__ips[0]}" __localip
	eval "$1=$__localip"
}

[ -z "$IP" ] && {
	listIPs IP
}

$dir/../backend/setup.sh

# Local consul instance

if [ -z "$MASTER_IP" ]
then
docker run -d --name cilium-consul \
    -v "$CONSUL_DIR":/data \
    -p $IP:8300:8300 \
    -p $IP:8301:8301 \
    -p $IP:8301:8301/udp \
    -p $IP:8302:8302 \
    -p $IP:8302:8302/udp \
    -p $IP:8400:8400 \
    -p $IP:8500:8500 \
    progrium/consul -server -advertise $IP -bootstrap-expect 1
MASTER_IP=$IP
else
docker run -d --name cilium-consul \
    -v "$CONSUL_DIR":/data \
    -p $IP:8300:8300 \
    -p $IP:8301:8301 \
    -p $IP:8301:8301/udp \
    -p $IP:8302:8302 \
    -p $IP:8302:8302/udp \
    -p $IP:8400:8400 \
    -p $IP:8500:8500 \
    progrium/consul -server -advertise $IP -join $MASTER_IP
fi

# Local elasticsearch instance
docker run \
	-d -p 9200:9200 -p 9300:9300 \
	--name "cilium-elastic" \
	--net "host" \
	-l "com.intent.service=gov_db" \
	-l "com.intent.logical-name=cilium-elastic" \
	elasticsearch:1.7.1 \
	elasticsearch \
		-Des.cluster.name="cilium-elastic" \
		-Des.multicast.enabled=true \
		-Des.transport.publish_host=$IP \
		-Des.http.publish_host=$IP \
		-Des.discovery.zen.ping.multicast.address=$IP

docker cp $dir/../external-deps/marvel-latest.zip cilium-elastic:/tmp

docker exec \
	cilium-elastic \
	/usr/share/elasticsearch/bin/plugin -i elasticsearch/marvel/latest \
	--url file:///tmp/marvel-latest.zip

docker restart cilium-elastic

docker run \
	-d --name cilium \
	--net "host" --pid "host" --privileged \
	-v /var/run/openvswitch/db.sock:/var/run/openvswitch/db.sock \
	-v /var/run/openvswitch/lxc-br0.mgmt:/var/run/openvswitch/lxc-br0.mgmt \
	-e HOST_IP=$IP \
	-e ELASTIC_IP=$IP \
	-e DOCKER_HOST="tcp://$IP:$SWARM_MASTER_PORT/" \
	cilium/cilium -l=debug -e=false -P $CILIUM_PORT

cat > $ADAPTER_FILE1 << EOF
version: 1
endpoints:
  "POST /containers/create":
    pre: [cilium]
  "POST /*/containers/create":
    pre: [cilium]
adapters:
  cilium: http://$IP:$CILIUM_PORT/docker/swarm/cilium-adapter
EOF

cat > $ADAPTER_FILE2 << EOF
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
  cilium: http://$IP:$CILIUM_PORT/docker/daemon/cilium-adapter
EOF

cat > $ADAPTER_FILE3 << EOF
version: 1
endpoints:
  "POST /api/v1/namespaces/*/pods":
    pre: [cilium]
adapters:
  cilium: http://$IP:$CILIUM_PORT/kubernetes/master/cilium-adapter
EOF

# powerstrip before docker swarm
if [ -z "$KUBERNETES" ]
then
docker run \
	-d --name cilium-powerstrip-pre-swarm \
	-e DOCKER_HOST=$IP:$SWARM_MASTER_PORT \
	-v $ADAPTER_FILE1:/etc/powerstrip/adapters.yml \
	-v /var/run/docker.sock:/var/run/docker.sock \
	-p $PWR_BEF_SWARM_PORT:2375 \
	cilium/powerstrip:latest
else
if [ -n "$IS_MASTER" ]
then
docker run \
	-d --name cilium-powerstrip-pre-k8s-master \
	-e KUBE_SERVER="tcp://$IP:$K8S_PORT" \
	-v $ADAPTER_FILE3:/etc/powerstrip/adapters.yml \
	-v /var/run/docker.sock:/var/run/docker.sock \
	-p $PWR_BEF_KUB_MASTER:8080 \
	cilium/powerstrip:kubernetes
fi
docker run \
	-d --name cilium-powerstrip-pre-pwr-daemon \
	-e DOCKER_HOST=$IP:$PWR_BEF_DAEMON_PORT \
	-v $ADAPTER_FILE1:/etc/powerstrip/adapters.yml \
	-v /var/run/docker.sock:/var/run/docker.sock \
	-p $PWR_BEF_SWARM_PORT:2375 \
	cilium/powerstrip:latest
fi

# powerstrip before docker daemon
docker run \
	-d --name cilium-powerstrip-pre-daemon \
	-v $ADAPTER_FILE2:/etc/powerstrip/adapters.yml \
	-v /var/run/docker.sock:/var/run/docker.sock \
	-p $PWR_BEF_DAEMON_PORT:2375 \
	cilium/powerstrip:latest

# Give powerstrip time to start up
sleep 3s

# Run a swarm agent on each node
docker run \
	-d --name cilium-swarm-agent \
	-l "com.intent.service=gov_swarm_events" \
	swarm:$SWARM_VERSION \
	join --advertise=$IP:$PWR_BEF_DAEMON_PORT consul://$IP:$CONSUL_PORT/ciliumnodes

# Given swarm agent time to start up
sleep 3s

# Run a swarm master on each node
docker run \
	-d -p $SWARM_MASTER_PORT:2375 \
	--name cilium-swarm-master \
	swarm:$SWARM_VERSION \
	manage --replication --advertise $IP:$SWARM_MASTER_PORT consul://$IP:$CONSUL_PORT/ciliumnodes

# Given swarm master time to start up or we might end up without an event manager
sleep 3s

docker run \
	-d --name cilium-swarm-event-handler \
	--net "host" --pid "host" --privileged \
	-e DOCKER_HOST=tcp://$IP:$SWARM_MASTER_PORT \
	-e HOST_IP=$IP \
	-e ELASTIC_IP=$IP \
	-v /var/run/openvswitch/db.sock:/var/run/openvswitch/db.sock \
	-v /var/run/openvswitch/lxc-br0.mgmt:/var/run/openvswitch/lxc-br0.mgmt \
	cilium/cilium -o -l=debug

# Run docker-collector
statisConfigDir=$(cd $dir/../external-deps/docker-collector-configs/ && pwd)

docker run \
        -d \
        --name cilium-docker-collector \
        -h "$(hostname)" \
        --pid host \
        --privileged \
        -e ELASTIC_IP=$IP \
        -v /var/run/docker.sock:/var/run/docker.sock \
        -v $statisConfigDir:/docker-collector/configs \
        cilium/docker-collector:latest \
	-f '(k8s_.*)|(cilium.*)' \
        -i 'cilium-docker-collector' \
        -l debug

echo "==========================================================================="
echo ""
echo " Successfully infected node $IP"
echo ""
echo " Master IP: $MASTER_IP"
echo ""
echo " Further nodes can be infected with: MASTER_IP=$MASTER_IP infect"
echo ""
echo "==========================================================================="
