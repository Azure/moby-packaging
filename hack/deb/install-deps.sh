#!/usr/bin/env bash

set -e -u -o xtrace

grep 'Build\-Depends' ./debian/control || exit 0

if [ ! -f ./debian/changelog ]; then
    echo 'dummy (0.0.1) dummy; urgency=low' > debian/changelog
    echo '  * Version: 1.0' >> debian/changelog
    echo ' -- Microsoft <support@microsoft.com>  Mon, 12 Mar 2018 00:00:00 +0000' >> debian/changelog
fi

if [ "${CROSS}" == "true" ] && [ ! "${BUILDPLATFORM}" = "${TARGETPLATFORM}" ]; then
    if [ ! -v DEB_BUILD_PROFILES ]; then
        export DEB_BUILD_PROFILES="cross"
    else
        export DEB_BUILD_PROFILES="${DEB_BUILD_PROFILES} cross"
    fi
fi

dpkg --add-architecture ${DEB_TARGETARCH}
apt-get update && mk-build-deps -i -r --host-arch ${DEB_TARGETARCH} -t \
    "apt-get -o Debug::pkgProblemResolver=yes -y"
