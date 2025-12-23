#!/usr/bin/env bats
load 'test_helper/bats-support/load'
load 'test_helper/bats-assert/load'

# Store all the version output so we don't have to call this for every version check test
export _VERSION_DETAILS=$(timeout --kill-after 60s 30s docker version --format '{{ json . }}')
if [ -n "${TARGETARCH}" ]; then
    TEST_PLATFORM="--platform linux/${TARGETARCH}"
fi

@test "test docker run hello world" {
    run timeout --kill-after=60s 30s docker run ${TEST_PLATFORM} --rm mcr.microsoft.com/azurelinux/base/core:3.0 echo hello from azl
    assert_output --partial "hello from azl"

    # Make sure --init works which has an extra binary
    run timeout --kill-after=60s 30s docker run --init --rm ${TEST_PLATFORM} mcr.microsoft.com/azurelinux/base/core:3.0 echo hello from azl
    assert_output --partial "hello from azl"
}

@test "test containerd run hello world" {
    timeout --kill-after=60s 40s ctr image pull ${TEST_PLATFORM} mcr.microsoft.com/azurelinux/base/core:3.0
    run timeout --kill-after=60s 40s ctr run ${TEST_PLATFORM} --rm mcr.microsoft.com/azurelinux/base/core:3.0 test echo hello from azl
    assert_output --partial "hello from azl"
}

@test "test buildx build" {
    timeout --kill-after=60s 30s docker buildx build ${TEST_PLATFORM} - <<-EOF
        FROM mcr.microsoft.com/azurelinux/base/core:3.0
        RUN echo hello world
EOF
}

@test "validate engine version" {
    if [ -z "${TEST_ENGINE_VERSION}" ]; then
        skip "no engine version specified to compare against"
    fi
    v="$(echo "${_VERSION_DETAILS}" | jq -r '.Server.Components[] | select(.Name=="Engine") | .Version')"
    assert_equal "${v}" "${TEST_ENGINE_VERSION}"
}

@test "validate engine commit" {
    if [ -z "${TEST_ENGINE_COMMIT}" ]; then
        skip "no engine commit specified to compare against"
    fi
    v="$(echo "${_VERSION_DETAILS}" | jq -r '.Server.Components[] | select(.Name=="Engine") | .Details.GitCommit')"
    assert_equal "${v}" "${TEST_ENGINE_COMMIT}"
}

@test "validate cli version" {
    if [ -z "${TEST_CLI_VERSION}" ]; then
        skip "no cli version specified to compare against"
    fi
    v="$(echo "${_VERSION_DETAILS}" | jq -r '.Client.Version')"
    assert_equal "${v}" "${TEST_CLI_VERSION}"
}

@test "validate cli commit" {
    if [ -z "${TEST_CLI_COMMIT}" ]; then
        skip "no cli commit specified to compare against"
    fi
    v="$(echo "${_VERSION_DETAILS}" | jq -r '.Client.GitCommit')"
    assert_equal "${v}" "${TEST_CLI_COMMIT}"
}

@test "validate containerd version" {
    if [ -z "${TEST_CONTAINERD_VERSION}" ]; then
        skip "no containerd version specified to compare against"
    fi
    v="$(echo "${_VERSION_DETAILS}" | jq -r '.Server.Components[] | select(.Name=="containerd") | .Version')"
    assert_equal "${v}" "${TEST_CONTAINERD_VERSION}"
}

@test "validate containerd commit" {
    if [ -z "${TEST_CONTAINERD_COMMIT}" ]; then
        skip "no containerd commit specified to compare against"
    fi
    v="$(echo "${_VERSION_DETAILS}" | jq -r '.Server.Components[] | select(.Name=="containerd") | .Details.GitCommit')"
    assert_equal "${v}" "${TEST_CONTAINERD_COMMIT}"
}

@test "validate runc commit" {
    if [ -z "${TEST_RUNC_COMMIT}" ]; then
        skip "no runc commit specified to compare against"
    fi
    v="$(echo "${_VERSION_DETAILS}" | jq -r '.Server.Components[] | select(.Name=="runc") | .Details.GitCommit')"
    assert_equal "${v}" "${TEST_RUNC_COMMIT}"
}

@test "validate runc version" {
    if [ -z "${TEST_RUNC_VERSION}" ]; then
        skip "no runc version specified to compare against"
    fi
    v="$(echo "${_VERSION_DETAILS}" | jq -r '.Server.Components[] | select(.Name=="runc") | .Version')"
    assert_equal "${v}" "${TEST_RUNC_VERSION}"
}

check_pkg() {
    if [ -n "$(command -v dpkg)" ]; then
        check_dpkg $1 $2
        return $?
    fi
    if [ -n "$(command -v rpm)" ]; then
        check_rpm $1 $2
        return $?
    fi
    return 1
}

check_dpkg() {
    ver="$(dpkg -l | grep $1 | awk '{ print $3 }')"
    if [ $? -gt 0 ]; then
        return 1
    fi
    [ "${ver}" = "${2}" ] && return 0
    echo "expected: ${2} -- actual: ${ver}"
    return 1
}

check_rpm() {
    v="$(rpm -qi $1 | grep  'Version' | awk '{ print $3 }')"
    if [ $? -gt 0 ]; then
        return 1
    fi
    rel="$(rpm -qi $1 | grep  'Release' | awk '{ print $3 }')"
    if [ $? -gt 0 ]; then
        return 1
    fi
    vv="${v}-${rel}"
    [ "${vv}" = "$2" ] && return 0
    echo "expected: ${2} -- actual: ${vv}"
    return 1
}

@test "validate runc package version" {
    if [ -z "${TEST_RUNC_PACKAGE_VERSION}" ]; then
        skip "no package version specified to compare against"
    fi

    check_pkg moby-runc "${TEST_RUNC_PACKAGE_VERSION}"
}

@test "validate engine package version" {
    if [ -z "${TEST_ENGINE_PACKAGE_VERSION}" ]; then
        skip "no package version specified to compare against"
    fi

    check_pkg moby-engine "${TEST_ENGINE_PACKAGE_VERSION}"
}

@test "validate cli package version" {
    if [ -z "${TEST_CLI_PACKAGE_VERSION}" ]; then
        skip "no package version specified to compare against"
    fi

    check_pkg moby-cli "${TEST_CLI_PACKAGE_VERSION}"
}

@test "validate containerd package version" {
    if [ -z "${TEST_CONTAINERD_PACKAGE_VERSION}" ]; then
        skip "no package version specified to compare against"
    fi

    check_pkg moby-containerd "${TEST_CONTAINERD_PACKAGE_VERSION}"
}

@test "validate buildx package version" {
    if [ -z "${TEST_BUILDX_PACKAGE_VERSION}" ]; then
        skip "no package version specified to compare against"
    fi

    check_pkg moby-buildx "${TEST_BUILDX_PACKAGE_VERSION}"
}

@test "validate compose package version" {
    if [ -z "${TEST_COMPOSE_PACKAGE_VERSION}" ]; then
        skip "no package version specified to compare against"
    fi

    check_pkg moby-compose "${TEST_COMPOSE_PACKAGE_VERSION}"
}

@test "compose plugin is installed" {
    run docker compose version
    assert_success
}

@test "buildx plugin is installed" {
    run docker buildx version
    assert_success
}
