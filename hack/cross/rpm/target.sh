#!/bin/sh

# This script is used to convert buildkit target arch values into
#  values that debbuild likes.
# To use it, "source" it into your script and use "DBIANTARGETARCH"

set -e -u

: ${RPMTARGETARCH=}

case "${TARGETARCH}" in
    "amd64")
        RPMTARGETARCH=x86_64
        ;;
	"arm64")
		RPMTARGETARCH=aarch64
		;;
	"arm")
		case "${TARGETVARIANT}" in
			"v6"|"v5")
				RPMTARGETARCH=armel;
				;;
			"v7"|"")
				RPMTARGETARCH=armv7hl
				;;
			*)
				RPMTARGETARCH="${TARGETARCH}"
				;;
			esac
		;;
	*)
		RPMTARGETARCH="${TARGETARCH}"
		;;
esac

echo ${RPMTARGETARCH}