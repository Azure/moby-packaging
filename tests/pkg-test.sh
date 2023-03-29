#!/usr/bin/env bash

set -u -o pipefail
shopt -s expand_aliases

: ${BUILD:="docker build"}
: ${TESTDIR:=.test}
_TESTDIR="${TESTDIR}/${DISTRO}"

DISTRO="$1"

if [ -z "${DISTRO}" ]; then
    echo must provide target distro as first argument
    exit 1
fi

install() {
    if [ -n "${2}" ]; then
        docker cp "${2}/" "${1}:/var/pkg/" || return $?
    fi
    docker exec \
        -e ENGINE_VERSION \
        -e CLI_VERSION \
        -e RUNC_VERSION \
        -e CONTAINERD_VERSION \
        -e BUILDX_VERSION \
        -e COMPOSE_VERSION \
        "${1}" /opt/moby/install.sh /var/pkg
}

test() {
    docker exec \
        -e CLI_COMMIT \
        -e CLI_VERSION \
        -e CONTRAINERD_COMMIT \
        -e CONTAINERD_VERSION \
        -e ENGINE_COMMIT \
        -e ENGINE_VERSION  \
        -e RUNC_VERSION \
        -e RUNC_COMMIT \
        -e BUILDX_VERSION \
        -e BUILDX_COMMIT \
        -e COMPOSE_VERSION \
        $1 bats -T /opt/moby/test.sh
}

log_and_exit() {
    let ec=$1
    shift
    >&2 echo $@
    exit $ec
}


create() {
    mkdir -p ${TESTDIR}/${DISTRO}

    extra_build_args=""
    : ${CUSTOM_IMG:=}
    if [ -n "${CUSTOM_IMG}" ]; then
        extra_build_args="--build-arg ${DISTRO^^}_IMG=${CUSTOM_IMG}"
    fi

    ${BUILD} --build-arg "${DISTRO^^}_IMG" ${extra_build_args} --pull --target $@ --target ${DISTRO}-test  --iidfile ${_TESTDIR}/imageid . || exit 1

    docker run \
        --cidfile ${_TESTDIR}/cid \
        -d -t \
        --security-opt seccomp:unconfined \
        --security-opt apparmor:unconfined \
        --security-opt label:disabled \
        --cap-add SYS_ADMIN \
        --cap-add NET_ADMIN \
        -e container=docker \
        --tmpfs /tmp \
        --tmpfs /run \
        --tmpfs /run/lock \
        -v /sys/fs/cgroup:/sys/fs/cgroup:ro \
        -v /var/lib/docker \
        -v /var/lib/containerd \
    $(cat ${_TESTDIR}/imageid)
}

cleanup() {
    docker rm -fv $1 > /dev/null || true
    rm -rf ${_TESTDIR}
}

trap "cleanup $id" EXIT

# validate that systemd is running correctly
let n=0
while true; do
    docker exec $id systemctl show-environment > /dev/null && break
    let n=$n+1
    [ $n -eq 5 ] && log_and_exit 1 "Could not validate that systemd is running properly"
    docker logs --tail 10 $id
    sleep 1
done

: ${LOCAL_PKG_DIR:=}
install "$id" "${LOCAL_PKG_DIR}" || log_and_exit $? "Installation failed"

test $id || log_and_exit $? "Failed tests"

log_and_exit 0 "Success!"
