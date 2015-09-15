.PHONY: cilium \
tests \
cilium-image \
cilium-binary \
clean-containers \
reset-cluster \
redeploy-cilium-image \
redeploy-images \
clean-swarm-vms \
clean-images \
build-images \
import-images \
setup-vagrant-machines \
infect \
refresh-policy \
run-kibana \
shell \
update-godeps \
start-services \
help

KERNEL:= $(shell uname -s)
MACHINE := $(shell uname -m)

cilium: tests ./cilium/cilium.go $(wildcard backend/*)
	@godep go clean -i
	@godep go build -o cilium-${KERNEL}-${MACHINE} ./cilium/cilium.go

tests:
	@godep go fmt ./cilium/...
	@godep go test ./cilium/...

cilium-image:
	@./scripts/build-cilium-image.sh

cilium-binary:
	@./scripts/build-cilium-binary.sh

clean-containers:
	@./scripts/clean-containers.sh

reset-cluster:
	@./scripts/reset-cluster.sh
	make refresh-policy
	@./scripts/setup-services.sh

redeploy-cilium-image:
	@./scripts/redeploy-cilium-image.sh

redeploy-images:
	@./scripts/redeploy-images.sh

clean-swarm-vms:
	@vagrant destroy -f
	@rm -fr .vagrant/

clean-images:
	rm -f images/*.ditar

build-images: clean-images
	@./scripts/build-images.sh

import-images:
	@./scripts/import-images.sh

setup-vagrant-machines: clean-swarm-vms
	@./scripts/setup-vagrant.sh
	@vagrant up --provider=virtualbox node1
	@vagrant up --provider=virtualbox node2
	@vagrant snapshot take node1 all-installed
	@vagrant snapshot take node2 all-installed

infect:
	@./scripts/infect-node.sh

refresh-policy:
	@./scripts/load-policy.sh

run-kibana:
	@./scripts/run-kibana.sh

shell:
	@DOCKER_HOST=localhost:2375 docker run -ti --privileged --rm -l com.intent.service=debug debug_shell

update-godeps:
	@./scripts/update-godeps.sh

start-services:
	@./scripts/setup-services.sh

help:
	@./scripts/print-help.sh
