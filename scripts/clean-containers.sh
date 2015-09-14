#!/bin/bash

CONT=$(docker ps -aq --filter name=cilium)
[ "$CONT" ] && {
	docker rm -f $CONT
}

exit 0
