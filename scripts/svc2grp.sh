#!/bin/bash
set -e
dir=`dirname $0`

ENDPOINT=$(curl -s -XGET "http://localhost:9200/cilium-state/endpoint/_search?q=$@")A
GROUP=$(echo $ENDPOINT | grep -e '"group":[0-9]*')

[ -z "$GROUP" ] &&  exit 1
echo $GROUP | sed 's/^.*"group":\([0-9]*\).*$/\1/'
exit 0
