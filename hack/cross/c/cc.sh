#!/bin/sh

case "${BUILDPLATFORM}" in
"linux/amd64")
	export C_BUILDARCH="x86_64-pc-linux-gnu"
	;;
"linux/arm64")
	export C_BUILDARCH="aarch64-linux-gnu"
	;;
"linux/arm/v5")
	export C_BUILDARCH="arm-linux-gnueabi"
	;;
"linux/arm/v7")
	export C_BUILDARCH="arm-linux-gnueabihf"
	;;
esac

if [ "${BUILDPLATFORM}" = "${TARGETPLATFORM}" ]; then
	return
fi


ccache="$(command -v ccache)"
if [ -n "$ccache" ]; then
	ccache="${ccache} "
fi

case "${TARGETOS}/${TARGETARCH}" in
"linux/amd64")
	export CC="${ccache}x86_64-linux-gnu-gcc"
	;;
"linux/ppc64le")
	export CC="${ccache}powerpc64le-linux-gnu-gcc"
	;;
"linux/s390x")
	export CC="${ccache}s390x-linux-gnu-gcc"
	;;
"linux/arm64")
	export C_HOSTARCH="aarch64-linux-gnu"
	export CC="${ccache}${C_HOSTARCH}-gcc"
	export CXX="${ccache}${C_HOSTARCH}-g++"
	export AR="${ccache}${C_HOSTARCH}-ar"
	export RANLIB="${ccache}${C_HOSTARCH}-ranlib"
	export LD="${ccache}${C_HOSTARCH}-ld"
	;;
"linux/arm")
	case "${TARGETVARIANT}" in
	"v5")
		export C_HOSTARCH="arm-linux-gnueabi"
		export CC="${ccache}${C_HOSTARCH}-gcc"
		export CXX="${ccache}${C_HOSTARCH}-g++"
		export AR="${ccache}${C_HOSTARCH}-ar"
		export RANLIB="${ccache}${C_HOSTARCH}-ranlib"
		export LD="${ccache}${C_HOSTARCH}-ld"
		;;
	*)
		export C_HOSTARCH="arm-linux-gnueabihf"
		export CC="${ccache}${C_HOSTARCH}-gcc"
		export CXX="${ccache}${C_HOSTARCH}-g++"
		export AR="${ccache}${C_HOSTARCH}-ar"
		export RANLIB="${ccache}${C_HOSTARCH}-ranlib"
		export LD="${ccache}${C_HOSTARCH}-ld"
		;;
	esac
	;;
"windows/amd64")
	export CC="x86_64-w64-mingw32-gcc"
	;;
esac

echo ${CC}