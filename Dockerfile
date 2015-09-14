FROM fedora:22
MAINTAINER "André Martins <aanm90@gmail.com>"
RUN dnf -y update
RUN dnf -y install openvswitch tcpdump sudo
RUN dnf clean all
RUN systemctl disable openvswitch.service
ADD cilium-Linux-x86_64 /usr/bin/cilium
ADD backend /opt/cilium/backend
ADD entrypoint.sh /opt/cilium/
ADD scripts /opt/cilium/scripts
ENV PIPEWORK /opt/cilium/backend/pipework
ENV ADD_ENDPOINT /opt/cilium/backend/add-endpoint.sh
ENV REMOVE_ENDPOINT /opt/cilium/backend/remove-endpoint.sh
WORKDIR /opt/cilium/
ENTRYPOINT ["cilium"]
