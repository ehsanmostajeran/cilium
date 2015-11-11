#!/bin/bash
set -e
dir=`dirname $0`

if [ -z "$ELASTIC_IP" ]; then
    ELASTIC_IP=192.168.50.1
fi

ENDPOINT=$(curl -s -XGET "http://$ELASTIC_IP:9200/cilium-state/endpoint/_search?q=$@")A
GROUP=$(echo $ENDPOINT | grep -e '"group":[0-9]*')

[ -z "$GROUP" ] &&  exit 1
echo $GROUP | sed 's/^.*"group":\([0-9]*\).*$/\1/'
exit 0
