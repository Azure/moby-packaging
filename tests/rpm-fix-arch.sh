#!/usr/bin/env bash

set -e -u

[ "${TARGETPLATFORM}" = "${BUILDPLATFORM}" ] && exit 0

case "${TARGETARCH}/${TARGETVARIANT}" in
    "arm/v7")
        echo "armhfp" > /etc/yum/vars/basearch
        echo "armv7hl" > /etc/yum/vars/arch
        echo "armv7hl-redhat-linux-gpu" > /etc/rpm/platform
        ;;
    *)
        rm -rf /etc/yum.repos.d/CentOS-Vault.repo
        ;;
esac