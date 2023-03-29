#!/bin/sh

# This script is used to convert buildkit target arch values into
#  values that debbuild likes.
# To use it, "source" it into your script and use "DBIANTARGETARCH"

set -e -u

: ${DEBIANTARGETARCH=}
: ${CROSS:=false}

if [ "${CROSS}" = "false" ]; then
    dpkg --print-architecture
	return
fi

case "${TARGETARCH}" in
	"arm")
		case "${TARGETVARIANT}" in
			"v5"|"v6")
				DEBIANTARGETARCH=armel;
				;;
			"v7"|"")
				DEBIANTARGETARCH=armhf
				;;
			*)
				DEBIANTARGETARCH="${TARGETARCH}"
				;;
			esac
		;;
	*)
		DEBIANTARGETARCH="${TARGETARCH}"
		;;
esac

echo ${DEBIANTARGETARCH}