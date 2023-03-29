#!/usr/bin/env bash
set -ex

: ${SRCROOT:=}
: ${DSTROOT:=}

[ -z "$SRCROOT" ] && exit 1
[ -z "$DSTROOT" ] && exit 1

. "$SRCROOT/mapping"

for src in "${!MAPPING[@]}"; do
    dst="$DSTROOT/${MAPPING["$src"]}"
    mkdir -vp "$(dirname "$dst")"
    cp -rv "$src" "$dst"
done
