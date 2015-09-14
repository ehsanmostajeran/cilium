#!/bin/bash

set -o errtrace
set -o nounset

dir="$(dirname "$0")"
source "$dir/config.sh"
source "$dir/utils.sh"

OFPORT=$(get_ofport $BRIDGE $1)
GRP=$2
BD=$3
NS=$4
IP=$5
MAC=$6
COOKIE=$(hex2cookie $7)

# Map OFPORT to sGRP, BD, and NS 
ofctl add-flow $BRIDGE "table=$TBL_PRE, cookie=$COOKIE, \
			in_port=$OFPORT, \
			actions=load:${GRP}->$REG_SGRP_OF,  \
				load:${BD}->$REG_BD_OF,  \
				load:${NS}->$REG_NS_OF,  \
				goto_table:$TBL_MAIN"

# Map dMAC to dGRP and OFPORT
ofctl add-flow $BRIDGE "table=$TBL_MAIN, cookie=$COOKIE, \
			$REG_BD=$BD, dl_dst=$MAC,\
			actions=load:${GRP}->$REG_DGRP_OF, \
				load:${OFPORT}->$REG_PORT_OF \
				goto_table:$TBL_POLICY"

# Translate local ARP broadcasts to unicasts
arp_optimize $BRIDGE $IP $MAC $BD $OFPORT $GRP $COOKIE

# Map dIP to dGRP and OFPORT and perform L3
ofctl add-flow $BRIDGE "priority=15, table=$TBL_MAIN, cookie=$COOKIE, \
			$REG_NS=$NS, dl_dst=$LOGICAL_ROUTER_MAC, ip, nw_dst=$IP, \
			actions=load:${GRP}->$REG_DGRP_OF, \
				load:${OFPORT}->$REG_PORT_OF, \
				mod_dl_dst:$MAC, \
				dec_ttl, \
				mod_dl_src:$LOGICAL_ROUTER_MAC, \
				goto_table:$TBL_POLICY"
