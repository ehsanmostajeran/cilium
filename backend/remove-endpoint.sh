#!/bin/bash

set -o errtrace

dir="$(dirname "$0")"
source "$dir/config.sh"
source "$dir/utils.sh"

COOKIE=$(hex2cookie $1)
PORT=$2

[ -z "PORT" ] && {
	PORT=$(ofctl dump-flows lxc-br0 cookie=$COOKIE/-1,table=0)
	PORT=$(echo $PORT | sed -n 's/.*in_port=\([0-9]*\).*/\1/p')
}

ofctl del-flows $BRIDGE cookie=$COOKIE/-1 | true

[ "$PORT" ] && {
	ovs-vsctl del-port $BRIDGE $PORT
}

exit 0
