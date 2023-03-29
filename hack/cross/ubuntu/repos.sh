#!/usr/bin/env bash

set -e -u -o xtrace

: ${CROSS:=""}

. /etc/lsb-release

if [ "${DISTRIB_ID}" == "Ubuntu" ] && [ "${CROSS}" = "true" ] && [ ! "${BUILDPLATFORM}" = "${TARGETPLATFORM}" ]; then \
    echo deb [arch=$(dpkg --print-architecture)] http://archive.ubuntu.com/ubuntu/ ${DISTRO} main multiverse restricted universe > /etc/apt/sources.list
    echo deb [arch=${DEB_TARGETARCH}] http://ports.ubuntu.com/ubuntu-ports/ ${DISTRO} main multiverse restricted universe >> /etc/apt/sources.list
    echo deb [arch=${DEB_TARGETARCH}] http://ports.ubuntu.com/ubuntu-ports/ ${DISTRO}-updates main multiverse restricted universe >> /etc/apt/sources.list
    echo deb [arch=$(dpkg --print-architecture)] http://archive.ubuntu.com/ubuntu/ ${DISTRO}-updates main multiverse restricted universe >> /etc/apt/sources.list
    echo deb [arch=$(dpkg --print-architecture)] http://security.ubuntu.com/ubuntu/ ${DISTRO}-security main multiverse restricted universe >> /etc/apt/sources.list
    cat /etc/apt/sources.list
fi