#!/bin/bash
set -e
dir=`dirname $0`

cd $dir/..
vagrant snapshot go node1 all-installed
vagrant snapshot go node2 all-installed

vagrant ssh -c "cd cilium; make import-images" node1
vagrant ssh -c "cd cilium; make import-images" node2

vagrant snapshot delete node1 all-installed
vagrant snapshot delete node2 all-installed

vagrant snapshot take node1 all-installed
vagrant snapshot take node2 all-installed

