#!/bin/sh

if [ ! -c /dev/kvm ]; then
    mknod /dev/kvm c 10 232
    chmod a+rw /dev/kvm
fi
rm /tmp/rootfs.qcow2 2>/dev/null
qemu-img create -f qcow2 -b /tmp/rootfs-base.qcow2 -F qcow2 /tmp/rootfs.qcow2

debug=""
if [ "${DEBUG}" = "true" ]; then
    debug="--debug"
fi

exec /usr/local/bin/docker-entrypoint --vm-port-forward=22 --uid=65534 --gid=65534 ${debug}
