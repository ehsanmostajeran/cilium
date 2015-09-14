#!/bin/bash

set -e
dir=`dirname $0`

echo "Installing vagrant-vbox-snapshot"
vagrant plugin install vagrant-vbox-snapshot
