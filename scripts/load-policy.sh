#!/bin/bash

dir="$(dirname "$0")"

$dir/../cilium-Linux-x86_64 -l debug -F
for policy in $dir/../policy/*; do
	$dir/../cilium-Linux-x86_64 -f "$policy" -l debug
done

$dir/../cilium-Linux-x86_64 -f  $dir/../examples/compose/app-policy.yml -l debug