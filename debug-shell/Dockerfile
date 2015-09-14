FROM fedora:22

RUN dnf -y update
RUN dnf -y install iproute tcpdump iperf bmon iputils bind-utils

ENTRYPOINT ["bash"]
