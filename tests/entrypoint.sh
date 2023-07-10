#!/usr/bin/env bash

set -e -o xtrace

systemctl mask getty
systemctl mask systemd-networkd.service || true
systemctl mask systemd-networkd.socket || true
systemctl mask systemd-networkd-wait-online.service || true
systemctl mask systemd-resolved.service || true
systemctl mask systemd-resolved.socket || true

cat /etc/hostname >/tmp/hostname
umount /etc/hostname
mv /tmp/hostname /etc/hostname

cat /etc/hosts >/tmp/hosts
umount /etc/hosts
mv /tmp/hosts /etc/hosts

cat /etc/resolv.conf >/tmp/resolv.conf
umount /etc/resolv.conf
mv /tmp/resolv.conf /etc/resolv.conf

exec /sbin/init
