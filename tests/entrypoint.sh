#!/usr/bin/env bash

set -e -o xtrace

systemctl mask getty
systemctl mask systemd-networkd || true
systemctl disable systemd-networkd || true
systemctl disable systemd-sysctl.service || true
systemctl mask systemd-sysctl.service || true
systemctl disable systemd-udevd || true
systemctl mask systemd-udevd || true
systemctl disable systemd-resolved.service || true
systemctl mask systemd-resolved.service || true
systemctl disable systemd-logind || true
systemctl mask systemd-logind || true
systemctl disable systemd-oomd || true
systemctl mask systemd-oomd || true
systemctl disable systemd-network-generator.service || true
systemctl mask systemd-network-generator.service || true
systemctl disable systemd-networkd-wait-online.service || true
systemctl mask systemd-networkd-wait-online.service || true
systemctl enable testingapi.service

cat /etc/hostname >/tmp/hostname
umount /etc/hostname || true
mv /tmp/hostname /etc/hostname

cat /etc/hosts >/tmp/hosts
umount /etc/hosts || true
mv /tmp/hosts /etc/hosts

cat /etc/resolv.conf >/tmp/resolv.conf
umount /etc/resolv.conf || true
mv /tmp/resolv.conf /etc/resolv.conf

exec /sbin/init
