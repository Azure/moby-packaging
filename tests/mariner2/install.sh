#!/usr/bin/env bash

set -e

: ${TEST_ENGINE_PACKAGE_VERSION:=''}
: ${TEST_CLI_PACKAGE_VERSION:=''}
: ${TEST_CONTAINERD_PACKAGE_VERSION:=''}
: ${TEST_RUNC_PACKAGE_VERSION:=''}
: ${TEST_BUILDX_PACKAGE_VERSION:=''}
: ${TEST_COMPOSE_PACKAGE_VERSION:=''}

DEFAULT_REPO_DIR=/var/pkg

prepare_local_yum() {
    dir="${DEFAULT_REPO_DIR}"
    if [ -n "${1}" ]; then
        dir="${1}"
    fi
    createrepo "${dir}"
    dnf config-manager --add-repo "file://${dir}"
}

install() {
    local packages=(
        "moby-engine$(with_glob ${TEST_ENGINE_PACKAGE_VERSION})"
        "moby-cli$(with_glob ${TEST_CLI_PACKAGE_VERSION})"
        "moby-containerd$(with_glob ${TEST_CONTAINERD_PACKAGE_VERSION})"
        "moby-runc$(with_glob ${TEST_RUNC_PACKAGE_VERSION})"
        "moby-buildx$(with_glob ${TEST_BUILDX_PACKAGE_VERSION})"
        "moby-compose$(with_glob ${TEST_COMPOSE_PACKAGE_VERSION})"
        "moby-tini$(with_glob ${TEST_TINI_PACKAGE_VERSION})"
    )

    dnf install -y --nogpgcheck "${packages[@]}"
}

with_glob() {
    # If $1 is nonempty, expand it with a glob. Otherwise, print nothing.
    printf "%s" ${1:+"-$1*"}
}

init() {
    systemctl enable --now docker
    if [ ! $? -eq 0 ]; then
        journalctl -u docker
        journalctl -xe
        exit 1
    fi
    systemctl enable --now docker.socket
    if [ ! $? -eq 0 ]; then
        journalctl -u docker
        journalctl -xe
        exit 1
    fi

    systemctl enable --now containerd
    if [ ! $? -eq 0 ]; then
        journalctl -u containerd
        journalctl -xe
        exit 1
    fi

    systemctl start containerd
    if [ ! $? -eq 0 ]; then
        journalctl -u containerd
        journalctl -xe
        exit 1
    fi
    systemctl start docker
    if [ ! $? -eq 0 ]; then
        journalctl -u docker
        journalctl -xe
        exit 1
    fi
}

case "${1}" in
repo)
    prepare_local_yum "${2}"
    ;;
install)
    install
    init
    ;;
"")
    prepare_local_yum
    install
    init
    ;;
*)
    prepare_local_yum
    install
    init
    ;;
esac
