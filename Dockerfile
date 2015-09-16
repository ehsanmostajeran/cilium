FROM fedora:22
MAINTAINER "André Martins <aanm90@gmail.com>"

ADD . /tmp/cilium-build/src/github.com/cilium-team/cilium

WORKDIR /tmp/cilium-build/src/github.com/cilium-team/cilium

RUN dnf -y update && \
dnf -y install golang openvswitch tcpdump sudo && \
systemctl disable openvswitch.service && \
mkdir -p /opt/cilium && \
GOPATH=/tmp/cilium-build:\
/tmp/cilium-build/src/github.com/cilium-team/cilium/Godep/_workspace:$GOPATH \
go build -o /opt/cilium/cilium cilium/cilium.go && \
mv backend /opt/cilium && \
rm -fr /tmp/cilium-build && \
dnf -y remove golang && \
dnf clean all

ENV PIPEWORK /opt/cilium/backend/pipework
ENV ADD_ENDPOINT /opt/cilium/backend/add-endpoint.sh
ENV REMOVE_ENDPOINT /opt/cilium/backend/remove-endpoint.sh

ENTRYPOINT ["/opt/cilium/cilium"]
