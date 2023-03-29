#!/usr/bin/env sh

if [ "${SKIP_GOLANG_WRAPPER}" = "1" ]; then
    exec /usr/local/go/bin/go "$@"
fi

: ${TARGETPLATFORM=}
: ${TARGETOS=}
: ${TARGETARCH=}
: ${TARGETVARIANT=}
: ${CGO_ENABLED=}
: ${GOARCH=}
: ${GOOS=}
: ${GOARM=}

if [ ! -z "$TARGETPLATFORM" ]; then
  os="$(echo $TARGETPLATFORM | cut -d"/" -f1)"
  arch="$(echo $TARGETPLATFORM | cut -d"/" -f2)"
  if [ ! -z "$os" ] && [ ! -z "$arch" ]; then
    export GOOS="$os"
    export GOARCH="$arch"
    if [ "$arch" = "arm" ]; then
      case "$(echo $TARGETPLATFORM | cut -d"/" -f3)" in
      "v5")
        export GOARM="5"
        ;;
      "v6")
        export GOARM="6"
        ;;
      *)
        export GOARM="7"
        ;;
      esac
    fi
  fi
fi

if [ ! -z "$TARGETOS" ]; then
  export GOOS="$TARGETOS"
fi

if [ ! -z "$TARGETARCH" ]; then
  export GOARCH="$TARGETARCH"
fi

if [ "$TARGETARCH" = "arm" ]; then
  if [ ! -z "$TARGETVARIANT" ]; then
    case "$TARGETVARIANT" in
    "v5")
      export GOARM="5"
      ;;
    "v6")
      export GOARM="6"
      ;;
    *)
      export GOARM="7"
      ;;
    esac
  else
    export GOARM="7"
  fi
fi

if [ "$CGO_ENABLED" = "1" ] && [ "${CROSS}" = "true" ] ; then
    if [ -z "${TARGETPLATFORM}" ] || [ ! "${TARGETPLATFORM}" = "${BUILDPLATFORM}" ]; then
      case "$GOOS" in
        "linux")
          case "$GOARCH" in
            "amd64")
              export CC="x86_64-linux-gnu-gcc"
              ;;
            "ppc64le")
              export CC="powerpc64le-linux-gnu-gcc"
              ;;
            "s390x")
              export CC="s390x-linux-gnu-gcc"
              ;;
            "arm64")
              export CC="aarch64-linux-gnu-gcc"
              ;;
            "arm")
              case "$GOARM" in
              "5")
                export CC="arm-linux-gnueabi-gcc"
                ;;
              *)
                export CC="arm-linux-gnueabihf-gcc"
                ;;
              esac
              ;;
          esac
          ;;
        "windows")
          case "$GOARCH" in
            "amd64")
              CC="x86_64-w64-mingw32-gcc"
              ;;
            *)
              CGO_ENABLED=0
              ;;
          esac
          ;;
        *)
          CGO_ENABLED=0
          ;;
      esac
  fi
fi

if [ "$1" = "build" ] && [ "$GOOS" = "linux" ] && [ ! "${BUILDPLATFORM}" = "${TARGETPLATFORM}" ]; then
  # This is only neccessary if we are cross compiling.
  # It makes sure libs for other architectures are visible to pkg-config.
  if command -v pkg-config; then
      PKG_CONFIG_PATH="$(pkg-config --variable pc_path pkg-config)";
      for i in $(find /usr/lib -name 'pkgconfig'); do
        PKG_CONFIG_PATH="${PKG_CONFIG_PATH}:$i"
      done
      export PKG_CONFIG_PATH
  fi
fi

if [ ! "$1" = "env" ]; then
  >&2 echo "${BUILDPLATFORM} -> ${TARGETPLATFORM}"
  >&2 echo GOOS=$GOOS GOARCH=$GOARCH GOARM=$GOARM /usr/local/go/bin/go "$@"
fi

exec /usr/local/go/bin/go "$@"