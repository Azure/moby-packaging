#!/usr/bin/env bash

set -e

: ${TEST_ENGINE_PACKAGE_VERSION:=''}
: ${TEST_CLI_PACKAGE_VERSION:=''}
: ${TEST_CONTAINERD_PACKAGE_VERSION:=''}
: ${TEST_RUNC_PACKAGE_VERSION:=''}
: ${TEST_BUILDX_PACKAGE_VERSION:=''}
: ${TEST_COMPOSE_PACKAGE_VERSION:=''}

: ${DEBIAN_FRONTEND=noninteractive}
export DEBIAN_FRONTEND

DEFAULT_REPO_DIR=/var/pkg

prepare_local_apt() {
    dir="${DEFAULT_REPO_DIR}"
    if [ -n "${1}" ]; then
        dir="${1}"
    fi
    if [ -z "$(ls ${dir}/*.deb)" ]; then
        return
    fi

    aptly repo create unstable
    aptly repo add unstable "${dir}"
    aptly publish repo -distribution=moby-local-testing -skip-signing unstable
    aptly serve -listen=127.0.0.1:8080 &

    echo "waiting for apt server to be ready"
    while true; do
        curl 127.0.0.1:8080 >/dev/null 2>&1 && break
        sleep 1
    done

    echo "deb [trusted=yes arch=amd64,armhf,arm64] http://localhost:8080/ moby-local-testing main" >/etc/apt/sources.list.d/local.list
}

install() {
    apt-get update
    apt-get install -y \
        moby-engine="${TEST_ENGINE_PACKAGE_VERSION}*" \
        moby-cli="${TEST_CLI_PACKAGE_VERSION}*" \
        moby-containerd="${TEST_CONTAINERD_PACKAGE_VERSION}*" \
        moby-runc="${TEST_RUNC_PACKAGE_VERSION}*" \
        moby-buildx="${TEST_BUILDX_PACKAGE_VERSION}*" \
        moby-compose="${TEST_COMPOSE_PACKAGE_VERSION}*"
}

init() {
    systemctl start docker
    if [ ! $? -eq 0 ]; then
        journalctl -u docker
        journalctl -xe
        return 1
    fi

    systemctl start containerd
    if [ ! $? -eq 0 ]; then
        journalctl -u containerd
        return 1
    fi
}

case "${1}" in
repo)
    prepare_local_apt "${2}"
    ;;
install)
    install
    init
    ;;
"")
    prepare_local_apt
    install
    pkill -9 aptly || true
    init
    ;;
*)
    if [ -d "${1}" ]; then
        prepare_local_apt
    fi
    install
    init
    ;;
esac
