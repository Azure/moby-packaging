#!/bin/sh

set -ex

if [ ! -c /dev/kvm ]; then
    mknod /dev/kvm c 10 232
    chmod a+rw /dev/kvm
fi
[ ! -f /tmp/rootfs.qcow2 ] || rm /tmp/rootfs.qcow2
qemu-img create -f qcow2 -b /tmp/rootfs-base.qcow2 -F qcow2 /tmp/rootfs.qcow2

debug=""
if [ "${DEBUG}" = "true" ]; then
    debug="--debug"
fi

exec /usr/local/bin/docker-entrypoint --vm-port-forward=22 --uid=65534 --gid=65534 ${debug}
