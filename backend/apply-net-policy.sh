#!/bin/bash

set -o errtrace
set -o nounset

dir="$(dirname "$0")"
source "$dir/config.sh"
source "$dir/utils.sh"

SRC=$1
DST=$2
POLICY=$3

# Remove default allow-all policy upon first manual rule
ofctl del-flows $BRIDGE "table=$TBL_POLICY, cookie=0xfffffffffffffff/-1" | true

FILTER1=""
FILTER2=""
ACTION=""
PRIO=100

if [ "$SRC" = "any" ]; then
	PRIO=$((PRIO - 10))
else
	FILTER1="$FILTER1, $REG_SGRP=$SRC"
	FILTER2="$FILTER2, $REG_DGRP=$SRC"
fi

if [ "$DST" = "any" ]; then
	PRIO=$((PRIO - 10))
else
	FILTER1="$FILTER1, $REG_DGRP=$DST"
	FILTER2="$FILTER2, $REG_SGRP=$DST"
fi

if [ "$POLICY" = "allow" ]; then
	ACTION="actions=output:$REG_PORT_OF"
elif [ "$POLICY" = "drop" ]; then
	ACTION="actions=drop"
else
	echo "Unknown action $POLICY"
	exit 1
fi

FLOW="table=$TBL_POLICY, priority=$PRIO"

echo "Inserting flow $FLOW, $FILTER1, $ACTION"
ofctl add-flow $BRIDGE "$FLOW, $FILTER1, $ACTION"

[ "$FILTER1" != "$FILTER2" ] && {
	echo "Inserting flow $FLOW, $FILTER2, $ACTION"
	ofctl add-flow $BRIDGE "$FLOW, $FILTER2, $ACTION"
}

exit 0
