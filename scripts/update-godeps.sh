#!/usr/bin/env bash
set -e
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

cd "${dir}/.."

deps=$(grep '[[:space:]]"ImportPath": ".*' ./Godeps/Godeps.json | \
	      sed 's/[[:space:]]*"ImportPath": "//g' | \
	      sed 's/",$//g' | \
	      sed -e '1d')

for dep in ${deps}
do
    echo "Updating: ${dep}"
    go get -u "${dep}"
    godep update "${dep}"
done

godep save -r ./...

exit 0
