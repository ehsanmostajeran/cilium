#!/bin/bash

set -o nounset
set -e

dir="$(dirname "$0")"
cd $dir/..

if [ $# -lt 3 ]; then
	echo "Usage: net-policy.sh SRC-SVC DST-SVC { allow | drop }"
	exit 0
fi

SRC=$1
DST=$2
POLICY=$3

if [ "$SRC" != "any" ]; then
	SRC=$($dir/../scripts/svc2grp.sh $SRC) || {
		echo "Unknown source service $1"
		exit 1
	}
fi

if [ "$DST" != "any" ]; then
	DST=$($dir/../scripts/svc2grp.sh $DST) || {
		echo "Unknown destination service $2"
		exit 1
	}
fi

if [ "$POLICY" != "allow" -a  "$POLICY" != "drop" ]; then
	echo "Unknown policy action $POLICY"
	exit 1
fi


echo "Applying policy: $1 ($SRC) <=> $2 ($DST): $POLICY"

sudo backend/apply-net-policy.sh $SRC $DST $POLICY
vagrant ssh -c "sudo cilium/backend/apply-net-policy.sh $SRC $DST $POLICY" node1
vagrant ssh -c "sudo cilium/backend/apply-net-policy.sh $SRC $DST $POLICY" node2
