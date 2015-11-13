#!/usr/bin/env bash
set -e
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

cd "${dir}/.."

for image in ./images/*.ditar; do
    echo "Importing ${image}..."
    docker load -i "${image}"
done

echo "All images successfully imported"

exit 0
