#!/bin/bash

function ofctl()
{
	sudo ovs-ofctl -O $OFVERSION "$@"
}

# mac2hex MAC - return MAC address in HEX format
#   MAC: Mac address in form XX:XX:XX:XX:XX:XX
function mac2hex()
{
        echo 0x${1//:/}
}

# ip2hex IP - return IP address in HEX format
#   IP: Mac address in form 1.1.1.1
function ip2hex()
{
        echo -n 0x
        printf '%02X' ${1//./ }
}

function hex2cookie()
{
	echo 0x$(echo $1 | cut -b '-14')
}

# get_ofport BRIDGE PORT - return openflow port number
#   BRIDGE:     Bridge name
#   PORT:       Name of port as configured with ovs-vsctl add-port [...]
function get_ofport()
{
        local BRNAME=$1
        local N=$(ofctl show $BRNAME | grep $2 | cut -d'(' -f1)

        if [ -z "$N" ]; then
                echo "Unable to determine OpenFlow port number for $1"
                exit 1
        fi

        echo $N
}

# add_flow_broadcast BRIDGE BD - Create a new BD context
#   BRIDGE:     Name of bridge to configure
#   BD:         BD context
#
# This function creates a new BD context on a BRIDGE. It must be called before
# the first member is added.
#
# The function:
#   * creates a new group table taking the BD ID as group ID
#   * inserts a flow in the $TBL_MAIN to redirect packets with a destination
#     MAC of ff:ff:ff:ff:ff:ff to the new group
#
function add_flow_broadcast()
{
        local BRNAME=$1
        local BD=$2

        ofctl add-group $BRNAME "group_id=$BD, type=all"

        ofctl add-flow $BRNAME \
                "priority=10, table=$TBL_MAIN, $REG_BD=$BD, \
                 dl_dst=ff:ff:ff:ff:ff:ff, actions=group:$BD"
}

# arp_respond BRIDGE IP MAC BD - Install flow to repond to ARP requests for IP with MAC in BD
#   BRIDGE:     Name of bridge to configure
#   IP:         IP to respond to
#   MAC:        MAC to respond with
#   [BD:]       BD context to respond in (optional)
#
# Installs a flow in the $TBL_MAIN to respond with ARP replies to ARP requests
# for the given IP address. If a BD is specified, only requests in the BD are
# responded to, otherwise all requests are answered.
#
function arp_respond()
{
        local BRNAME=$1
        local IP=$2
        local MAC=$3

        local FILTER="arp, arp_op=1, arp_tpa=$IP"

        # Limited to a BD?
        [[ ${4:-} ]] && {
                FILTER="$FILTER, $REG_BD=$4"
        }

        local ACTIONS="move:NXM_OF_ETH_SRC[]->NXM_OF_ETH_DST[], \
                       mod_dl_src:$MAC, \
                       load:2->NXM_OF_ARP_OP[], \
                       move:NXM_NX_ARP_SHA[]->NXM_NX_ARP_THA[], \
                       load:$(mac2hex $MAC)->NXM_NX_ARP_SHA[], \
                       move:NXM_OF_ARP_SPA[]->NXM_OF_ARP_TPA[], \
                       load:$(ip2hex $IP)->NXM_OF_ARP_SPA[], \
                       in_port"

        # Reply to ARP requests for router gateway address
        ofctl add-flow $BRNAME "priority=20, table=$TBL_MAIN, $FILTER, actions=$ACTIONS"
}

# arp_optimize BRIDGE IP MAC BD OFPORT GRP TUN_DST - Install flow to repond to ARP requests for IP with MAC in BD
#   BRIDGE:     Name of bridge to configure
#   IP:         IP request to optimize
#   MAC:        MAC of IP
#   BD:         BD context of MAC
#   OFPORT:     OF port of EP (local or tunnel)
#   GRP:        Group of endpoint
#   [TUN_DST:]  Tunnel destination of EP (optional)
#
# Installs a flow in the $TBL_MAIN to translate broadcassted ARP requests to unicast
# ARP requests.
#
function arp_optimize()
{
        local BRNAME=$1
        local IP=$2
        local MAC=$3
        local BD=$4
        local OFPORT=$5
        local GRP=$6
	local COOKIE=$7
        local TUNNEL=

        local FILTER="$REG_BD=$4, arp, arp_op=1, arp_tpa=$IP, dl_dst=ff:ff:ff:ff:ff:ff"

        local ACTIONS="mod_dl_dst:$MAC, \
                       load:${GRP}->$REG_DGRP_OF, \
                       load:${OFPORT}->$REG_PORT_OF"

        # Tunnel endpoint?
        [[ ${8:-} ]] && {
                ACTIONS="$ACTIONS,move:$REG_SGRP_OF->NXM_NX_TUN_ID[0..31]"
                ACTIONS="$ACTIONS,load:$(ip2hex $8)->NXM_NX_TUN_IPV4_DST[]"
                TUNNEL=$8
        }

        ACTIONS="$ACTIONS,goto_table:$TBL_POLICY"

        # The priority of the flow must be higher than the flood broadcast catcher
        ofctl add-flow $BRNAME "priority=15, table=$TBL_MAIN, $FILTER, actions=$ACTIONS"
}

# add_router BRIDGE ADDR BD - Agent is told to setup a new default router
#   BRIDGE:     Name of bridge
#   ADDR:       Address of default router
#   BD:		BD context
#
function add_router()
{
        local BRNAME=$1
        local ADDR=$2
	local BD=$3
	local ROUTER_MAC=$4

        # The agent can choose any MAC address as long as it is unique
        arp_respond $BRNAME $ADDR $ROUTER_MAC $BD
}
