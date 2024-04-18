ARG MIRROR=mcr.microsoft.com/mirror/docker/library/

ARG MARINER2_IMG=mcr.microsoft.com/cbl-mariner/base/core:2.0
ARG CENTOS7_IMG=${MIRROR}centos:7
ARG RHEL8_IMG=${MIRROR}almalinux:8
ARG RHEL9_IMG=${MIRROR}almalinux:9
ARG BUSTER_IMG=${MIRROR}buildpack-deps:buster
ARG BULLSEYE_IMG=${MIRROR}buildpack-deps:bullseye
ARG BOOKWORM_IMG=${MIRROR}buildpack-deps:bookworm
ARG BIONIC_IMG=${MIRROR}buildpack-deps:bionic
ARG FOCAL_IMG=${MIRROR}buildpack-deps:focal
ARG JAMMY_IMG=${MIRROR}buildpack-deps:jammy
ARG NOBLE_IMG=${MIRROR}buildpack-deps:noble

ARG INCLUDE_TESTING=0
ARG INCLUDE_STAGING=0


# For jammy which uses zstd compressed packages, but aptly from the distro packages does not.
FROM mcr.microsoft.com/oss/go/microsoft/golang:1.19.12 AS aptly
ENV GOCACHE=/root/.gobuild GOPROXY=direct GOGC=off CGO_ENABLED=0
RUN \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.gobuild \
    go install github.com/aptly-dev/aptly@0a1798869a3c37b269f8f34fef41095233cd8750

FROM ${MIRROR}buildpack-deps:bullseye AS git

FROM git AS bats-core
RUN mkdir -p /root/bats
WORKDIR /root/src
ARG BATS_CORE_REPO=https://github.com/bats-core/bats-core.git
ARG BATS_CORE_COMMIT=743b02b27c888eba6bb60931656cc16bd751e544
ENV REPO=${BATS_CORE_REPO}
ENV COMMIT=${BATS_CORE_COMMIT}
RUN git clone ${REPO} . && \
    git checkout ${COMMIT} && \
    git archive HEAD | tar -C /root/bats -xvf -

FROM git as bats-support
RUN mkdir -p /root/bats
WORKDIR /root/src
ARG BATS_SUPPORT_REPO=https://github.com/bats-core/bats-support.git
ARG BATS_SUPPORT_COMMIT=24a72e14349690bcbf7c151b9d2d1cdd32d36eb1
ENV REPO=${BATS_SUPPORT_REPO}
ENV COMMIT=${BATS_SUPPORT_COMMIT}
RUN git clone ${REPO} . && \
    git checkout ${COMMIT} && \
    git archive HEAD | tar -C  /root/bats -xvf -

FROM git as bats-assert
RUN mkdir -p /root/bats
WORKDIR /root/src
ARG BATS_ASSERT_REPO=https://github.com/bats-core/bats-assert.git
ARG BATS_ASSERT_COMMIT=0a8dd57e2cc6d4cc064b1ed6b4e79b9f7fee096f
ENV REPO=${BATS_ASSERT_REPO}
ENV COMMIT=${BATS_ASSERT_COMMIT}
RUN git clone ${REPO} . && \
    git checkout ${COMMIT} && \
    git archive HEAD | tar -C  /root/bats -xvf -

FROM scratch AS bats
COPY --from=bats-core /root/bats/ /

FROM ${BUSTER_IMG} AS buster
RUN apt-get update && apt-get install -y systemd curl ca-certificates aptly && c_rehash
RUN \
    curl -SLf https://packages.microsoft.com/config/debian/10/packages-microsoft-prod.deb > /tmp/ms.deb && \
    dpkg -i /tmp/ms.deb && \
    rm /tmp/ms.deb
COPY entrypoint.sh /usr/local/bin/docker-entrypoint.sh
STOPSIGNAL SIGRTMIN+3
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
COPY --from=bats / /opt/bats
ARG INCLUDE_TESTING
ARG INCLUDE_STAGING
RUN if [ "${INCLUDE_TESTING}" = "1" ]; then \
    echo "deb [arch=amd64,arm64,armhf,arm64,armhf] https://packages.microsoft.com/debian/10/prod testing main" >> /etc/apt/sources.list.d/microsoft-testing.list; \
    fi; \
    if [ "${INCLUDE_STAGING}" = "1" ]; then \
    echo "deb [arch=amd64,arm64,armhf,arm64,armhf] https://packages.microsoft.com/repos/iotedge-buster buster main" >> /etc/apt/sources.list.d/microsoft-testing.list; \
    fi; \
    cd /opt/bats && ./install.sh /usr/local

FROM ${BULLSEYE_IMG} AS bullseye
RUN apt-get update && apt-get install -y systemd curl ca-certificates aptly && c_rehash
RUN \
    curl -SLf https://packages.microsoft.com/config/debian/11/packages-microsoft-prod.deb > /tmp/ms.deb && \
    dpkg -i /tmp/ms.deb && \
    rm /tmp/ms.deb
COPY entrypoint.sh /usr/local/bin/docker-entrypoint.sh
STOPSIGNAL SIGRTMIN+3
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
COPY --from=bats / /opt/bats
ARG INCLUDE_TESTING
ARG INCLUDE_STAGING
RUN if [ "${INCLUDE_TESTING}" = "1" ]; then \
    echo "deb [arch=amd64,arm64,armhf,arm64,armhf] https://packages.microsoft.com/debian/11/prod testing main" >> /etc/apt/sources.list.d/microsoft-testing.list; \
    fi; \
    if [ "${INCLUDE_STAGING}" = "1" ]; then \
    echo "STAGING NOT SUPPORTED ON BULLSEYE YET"; \
    exit 1; \
    fi; \
    cd /opt/bats && ./install.sh /usr/local
RUN ln -s /lib/systemd/systemd /sbin/init

FROM ${BOOKWORM_IMG} AS bookworm
RUN apt-get update && apt-get install -y systemd curl ca-certificates aptly && c_rehash
RUN \
    curl -SLf https://packages.microsoft.com/config/debian/12/packages-microsoft-prod.deb > /tmp/ms.deb && \
    dpkg -i /tmp/ms.deb && \
    rm /tmp/ms.deb
COPY entrypoint.sh /usr/local/bin/docker-entrypoint.sh
STOPSIGNAL SIGRTMIN+3
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
COPY --from=bats / /opt/bats
ARG INCLUDE_TESTING
ARG INCLUDE_STAGING
RUN if [ "${INCLUDE_TESTING}" = "1" ]; then \
    echo "deb [arch=amd64,arm64,armhf,arm64,armhf] https://packages.microsoft.com/debian/12/prod testing main" >> /etc/apt/sources.list.d/microsoft-testing.list; \
    fi; \
    if [ "${INCLUDE_STAGING}" = "1" ]; then \
    echo "STAGING NOT SUPPORTED ON BULLSEYE YET"; \
    exit 1; \
    fi; \
    cd /opt/bats && ./install.sh /usr/local
RUN ln -s /lib/systemd/systemd /sbin/init


FROM ${BIONIC_IMG} AS bionic
RUN apt-get update && apt-get install -y systemd curl ca-certificates aptly
RUN \
    curl -SLf https://packages.microsoft.com/config/ubuntu/18.04/multiarch/packages-microsoft-prod.deb > /tmp/ms.deb && \
    dpkg -i /tmp/ms.deb && \
    rm /tmp/ms.deb
COPY entrypoint.sh /usr/local/bin/docker-entrypoint.sh
STOPSIGNAL SIGRTMIN+3
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
COPY --from=bats / /opt/bats
ARG INCLUDE_TESTING
ARG INCLUDE_STAGING
RUN if [ "${INCLUDE_TESTING}" = "1" ]; then \
    echo "deb [arch=amd64,arm64,armhf] https://packages.microsoft.com/ubuntu/18.04/prod testing main" >> /etc/apt/sources.list.d/microsoft-testing.list; \
    fi; \
    if [ "${INCLUDE_STAGING}" = "1" ]; then \
    echo "deb [arch=amd64,arm64,armhf,arm64,armhf] https://packages.microsoft.com/repos/iotedge-bionic bionic main" >> /etc/apt/sources.list.d/microsoft-testing.list; \
    fi; \
    cd /opt/bats && ./install.sh /usr/local

FROM ${FOCAL_IMG} AS focal
RUN apt-get update && apt-get install -y systemd curl ca-certificates aptly
RUN \
    curl -SLf https://packages.microsoft.com/config/ubuntu/20.04/packages-microsoft-prod.deb > /tmp/ms.deb && \
    dpkg -i /tmp/ms.deb && \
    rm /tmp/ms.deb
ARG INCLUDE_TESTING
COPY entrypoint.sh /usr/local/bin/docker-entrypoint.sh
STOPSIGNAL SIGRTMIN+3
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
COPY --from=bats / /opt/bats
RUN if [ "${INCLUDE_TESTING}" = "1" ]; then \
    echo "deb [arch=amd64,arm64,armhf] https://packages.microsoft.com/ubuntu/20.04/prod testing main" >> /etc/apt/sources.list.d/microsoft-testing.list; \
    fi; \
    cd /opt/bats && ./install.sh /usr/local
RUN ln -s /lib/systemd/systemd /sbin/init

FROM ${JAMMY_IMG} AS jammy
RUN apt-get update && apt-get install -y systemd curl ca-certificates
COPY --link --from=aptly /go/bin/aptly /usr/local/bin/
RUN \
    curl -SLf https://packages.microsoft.com/config/ubuntu/22.04/packages-microsoft-prod.deb > /tmp/ms.deb && \
    dpkg -i /tmp/ms.deb && \
    rm /tmp/ms.deb
ARG INCLUDE_TESTING
COPY entrypoint.sh /usr/local/bin/docker-entrypoint.sh
STOPSIGNAL SIGRTMIN+3
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
COPY --from=bats / /opt/bats
RUN if [ "${INCLUDE_TESTING}" = "1" ]; then \
    echo "deb [arch=amd64,arm64,armhf] https://packages.microsoft.com/ubuntu/22.04/prod testing main" >> /etc/apt/sources.list.d/microsoft-testing.list; \
    fi; \
    cd /opt/bats && ./install.sh /usr/local
RUN ln -s /lib/systemd/systemd /sbin/init

FROM ${NOBLE_IMG} AS noble
RUN apt-get update && apt-get install -y systemd curl ca-certificates aptly
RUN \
    curl -SLf https://packages.microsoft.com/config/ubuntu/24.04/packages-microsoft-prod.deb > /tmp/ms.deb && \
    dpkg -i /tmp/ms.deb && \
    rm /tmp/ms.deb
ARG INCLUDE_TESTING
COPY entrypoint.sh /usr/local/bin/docker-entrypoint.sh
STOPSIGNAL SIGRTMIN+3
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
COPY --from=bats / /opt/bats
RUN if [ "${INCLUDE_TESTING}" = "1" ]; then \
    echo "deb [arch=amd64,arm64,armhf] https://packages.microsoft.com/ubuntu/24.04/prod testing main" >> /etc/apt/sources.list.d/microsoft-testing.list; \
    fi; \
    cd /opt/bats && ./install.sh /usr/local
RUN ln -s /lib/systemd/systemd /sbin/init

FROM ${CENTOS7_IMG} AS centos7-jq
WORKDIR /jq
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETARCH
ARG TARGETVARIANT
RUN yum install -y epel-release
RUN yum install --downloadonly --downloaddir=. jq

FROM ${CENTOS7_IMG} AS centos7
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETARCH
ARG TARGETVARIANT
RUN yum install -y systemd ca-certificates
RUN yum install -y https://packages.microsoft.com/config/centos/7/packages-microsoft-prod.rpm
ARG INCLUDE_TESTING
COPY entrypoint.sh /usr/local/bin/docker-entrypoint.sh
STOPSIGNAL SIGRTMIN+3
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
COPY --from=bats / /opt/bats
COPY --from=centos7-jq /jq /opt/jq
RUN cd /opt/jq && yum install -y ./*.rpm
RUN if [ "${INCLUDE_TESTING}" = "1" ]; then \
    yum-config-manager --add-repo https://packages.microsoft.com/yumrepos/microsoft-centos7-testing-prod || exit 1; \
    fi; \
    cd /opt/bats && ./install.sh /usr/local

FROM centos7 AS centos7-test
RUN yum install -y jq createrepo
COPY centos7/install.sh /opt/moby/
COPY test.sh /opt/moby/
COPY --from=bats-support /root/bats /opt/moby/test_helper/bats-support
COPY --from=bats-assert /root/bats /opt/moby/test_helper/bats-assert

FROM ${MARINER2_IMG} AS mariner2
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETARCH
ARG TARGETVARIANT
RUN tdnf install -y systemd ca-certificates util-linux dnf-utils libcgroup tar libseccomp iptables awk
ARG INCLUDE_TESTING
COPY entrypoint.sh /usr/local/bin/docker-entrypoint.sh
STOPSIGNAL SIGRTMIN+3
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
COPY --from=bats / /opt/bats
RUN cd /opt/bats && ./install.sh /usr/local

FROM mariner2 AS mariner2-test
RUN tdnf install -y jq createrepo wget
COPY mariner2/ /opt/moby/
COPY test.sh /opt/moby/
COPY --from=bats-support /root/bats /opt/moby/test_helper/bats-support
COPY --from=bats-assert /root/bats /opt/moby/test_helper/bats-assert

FROM ${RHEL8_IMG} AS rhel8
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETARCH
ARG TARGETVARIANT
ARG INCLUDE_TESTING
RUN yum install -y systemd ca-certificates 'dnf-command(config-manager)'
RUN yum install -y https://packages.microsoft.com/config/rhel/8/packages-microsoft-prod.rpm
COPY entrypoint.sh /usr/local/bin/docker-entrypoint.sh
STOPSIGNAL SIGRTMIN+3
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
COPY --from=bats / /opt/bats
RUN if [ "${INCLUDE_TESTING}" = "1" ]; then \
    dnf config-manager --add-repo https://packages.microsoft.com/yumrepos/microsoft-rhel8.0-testing-prod || exit 1; \
    fi; \
    cd /opt/bats && ./install.sh /usr/local

FROM rhel8 AS rhel8-test
RUN yum install -y jq createrepo
COPY centos8/install.sh /opt/moby/
COPY test.sh /opt/moby/
COPY --from=bats-support /root/bats /opt/moby/test_helper/bats-support
COPY --from=bats-assert /root/bats /opt/moby/test_helper/bats-assert

FROM ${RHEL9_IMG} AS rhel9
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETARCH
ARG TARGETVARIANT
ARG INCLUDE_TESTING
RUN yum install -y systemd ca-certificates 'dnf-command(config-manager)'
RUN yum install -y https://packages.microsoft.com/config/rhel/9.0/packages-microsoft-prod.rpm
COPY entrypoint.sh /usr/local/bin/docker-entrypoint.sh
STOPSIGNAL SIGRTMIN+3
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
COPY --from=bats / /opt/bats
RUN if [ "${INCLUDE_TESTING}" = "1" ]; then \
    dnf config-manager --add-repo https://packages.microsoft.com/yumrepos/microsoft-rhel9.0-testing-prod || exit 1; \
    fi; \
    cd /opt/bats && ./install.sh /usr/local

FROM rhel9 AS rhel9-test
RUN yum install -y jq createrepo
COPY centos8/install.sh /opt/moby/
COPY test.sh /opt/moby/
COPY --from=bats-support /root/bats /opt/moby/test_helper/bats-support
COPY --from=bats-assert /root/bats /opt/moby/test_helper/bats-assert

FROM buster AS buster-test
RUN apt-get update && apt-get install -y jq
COPY deb/install.sh /opt/moby/
COPY test.sh /opt/moby/
COPY --from=bats-support /root/bats /opt/moby/test_helper/bats-support
COPY --from=bats-assert /root/bats /opt/moby/test_helper/bats-assert

FROM bullseye AS bullseye-test
RUN apt-get update && apt-get install -y jq
COPY deb/install.sh /opt/moby/
COPY test.sh /opt/moby/
COPY --from=bats-support /root/bats /opt/moby/test_helper/bats-support
COPY --from=bats-assert /root/bats /opt/moby/test_helper/bats-assert

FROM bookworm AS bookworm-test
RUN apt-get update && apt-get install -y jq
COPY deb/install.sh /opt/moby/
COPY test.sh /opt/moby/
COPY --from=bats-support /root/bats /opt/moby/test_helper/bats-support
COPY --from=bats-assert /root/bats /opt/moby/test_helper/bats-assert

FROM bionic AS bionic-test
RUN apt-get update && apt-get install -y jq
COPY deb/install.sh /opt/moby/
COPY test.sh /opt/moby/
COPY --from=bats-support /root/bats /opt/moby/test_helper/bats-support
COPY --from=bats-assert /root/bats /opt/moby/test_helper/bats-assert

FROM focal AS focal-test
RUN apt-get update && apt-get install -y jq
COPY deb/install.sh /opt/moby/
COPY test.sh /opt/moby/
COPY --from=bats-support /root/bats /opt/moby/test_helper/bats-support
COPY --from=bats-assert /root/bats /opt/moby/test_helper/bats-assert

FROM jammy AS jammy-test
RUN apt-get update && apt-get install -y jq
COPY deb/install.sh /opt/moby/
COPY test.sh /opt/moby/
COPY --from=bats-support /root/bats /opt/moby/test_helper/bats-support
COPY --from=bats-assert /root/bats /opt/moby/test_helper/bats-assert

FROM noble AS noble-test
RUN apt-get update && apt-get install -y jq
COPY deb/install.sh /opt/moby/
COPY test.sh /opt/moby/
COPY --from=bats-support /root/bats /opt/moby/test_helper/bats-support
COPY --from=bats-assert /root/bats /opt/moby/test_helper/bats-assert
