#!/bin/bash

set -e
dir=`dirname $0`

for IMAGE in $dir/../images/*.ditar; do
	echo "Importing ${IMAGE}..."
	docker load -i "$IMAGE"
done
