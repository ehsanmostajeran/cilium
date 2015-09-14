#!/bin/bash

set -o errtrace
set -o nounset

dir="$(dirname "$0")"
source "$dir/config.sh"
source "$dir/utils.sh"

echo "Creating bridge $BRIDGE ..."
sudo ovs-vsctl del-br $BRIDGE 2> /dev/null
sudo ovs-vsctl add-br $BRIDGE

# Set OpenFlow versions
sudo ovs-vsctl set bridge $BRIDGE protocols=$OFVERSION

# Delete all flows
ofctl del-flows $BRIDGE

# Default table policies
ofctl mod-table $BRIDGE $TBL_PRE drop
ofctl mod-table $BRIDGE $TBL_MAIN drop
ofctl mod-table $BRIDGE $TBL_POLICY drop

echo "Creating tunnel port $TUNNEL ..."
sudo ovs-vsctl add-port $BRIDGE $TUNNEL -- \
	set Interface $TUNNEL type=vxlan \
		options:remote_ip=flow \
		options:key=flow
TUNNEL_OFPORT=$(get_ofport $BRIDGE $TUNNEL)

add_flow_broadcast $BRIDGE $BD_MAIN
add_router $BRIDGE $LOGICAL_ROUTER $BD_MAIN $LOGICAL_ROUTER_MAC

sudo ovs-vsctl add-port $BRIDGE $HOSTPORT -- \
	set Interface $HOSTPORT type=internal
sudo ip link set $HOSTPORT up

# Default policy: allow everything
ofctl add-flow $BRIDGE "table=$TBL_POLICY, cookie=0xfffffffffffffff, actions=output:$REG_PORT_OF"
