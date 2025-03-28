FROM ubuntu:22.04

ENV container=docker

# Don't start any optional services except for the few we need.
RUN find /etc/systemd/system \
    /lib/systemd/system \
    -path '*.wants/*' \
    -not -name '*journald*' \
    -not -name '*systemd-tmpfiles*' \
    -not -name '*systemd-user-sessions*' \
    -exec rm -f {} +

RUN apt-get update && \
    apt-get install -y \
        dbus systemd openssh-server net-tools iproute2 iputils-ping curl wget vim-tiny sudo \
        conntrack iptables nftables && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN >/etc/machine-id
RUN >/var/lib/dbus/machine-id

# Make sure user "ubuntu" has sudoer permission
RUN useradd -ms /bin/bash ubuntu && \
    echo "ubuntu ALL=(ALL:ALL) NOPASSWD: ALL" > /etc/sudoers.d/ubuntu

EXPOSE 22

RUN systemctl set-default multi-user.target
RUN systemctl mask \
      dev-hugepages.mount \
      sys-fs-fuse-connections.mount \
      systemd-update-utmp.service \
      systemd-tmpfiles-setup.service \
      console-getty.service
RUN systemctl disable \
      networkd-dispatcher.service

# This container image doesn't have locales installed. Disable forwarding the
# user locale env variables or we get warnings such as:
#  bash: warning: setlocale: LC_ALL: cannot change locale
RUN sed -i -e 's/^AcceptEnv LANG LC_\*$/#AcceptEnv LANG LC_*/' /etc/ssh/sshd_config

# https://www.freedesktop.org/wiki/Software/systemd/ContainerInterface/
STOPSIGNAL SIGRTMIN+3

# Create the /kind folder to facilitate kind's hacking/patching process
RUN mkdir /kind

# The entrypoint is copied from kind project
# Here: https://github.com/kubernetes-sigs/kind/blob/main/images/base/files/usr/local/bin/entrypoint
COPY --chmod=0755 entrypoint /usr/local/bin/entrypoint

# ENTRYPOINT [ "/usr/local/bin/entrypoint", "/bin/bash" ]
