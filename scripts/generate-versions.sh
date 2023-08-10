#!/usr/bin/env bash

export AZURE_STORAGE_AUTH_MODE=login

set -u

: "${AZURE_STORAGE_ACCOUNT}"
: "${STORAGE_CONTAINER}"
: "${FILTER_PREFIX:=}"
: "${OUTPUT:=_versions}"
: "${URL_PREFIX:="https://mobyartifacts.azureedge.net/${STORAGE_CONTAINER}"}"
: "${INDEX_URL:="https://mobyartifacts.azureedge.net/index"}"

jsonFile="$(mktemp --suffix=.json tmp.parse-package-list.XXXXXXXXXX)"
trap "rm -rf $jsonFile*" EXIT

log() {
    echo "$@" >&2
}

az storage blob list \
    --auth-mode=login \
    -o json \
    --container-name="${STORAGE_CONTAINER}" \
    --prefix="${FILTER_PREFIX}" \
    --include m \
    --num-results=* \
    --query="[].{name: name, sha256: metadata.sha256}" >"$jsonFile"

minVersionsFile="${BASH_SOURCE[0]%/*}/min-versions.json"

jqFlags="-L ${BASH_SOURCE[0]%/*} --arg URL_PREFIX "${URL_PREFIX}" --argjson minVersions $(jq -c . "${minVersionsFile}")"

# set -x
mkdir -p "${OUTPUT}"

read -r -d '' Q <<-'EOF'
include "generate-versions";
reduce_pkg
EOF

parsed_file="${jsonFile}.parsed"
jq ${jqFlags} "${Q}" "${jsonFile}" >"${parsed_file}"

read -r -d '' Q <<-'EOF'
include "generate-versions";
only_latest | build_index
EOF

latest_file="${OUTPUT}/latest.json"
jq ${jqFlags} "${Q}" "${parsed_file}" >"${latest_file}"

rss_feed_from_entries() {
    entries="$1"
    out="$2"

    cat <<EOF >"${out}"
<?xml version="1.0" encoding="utf-8" ?>
<rss xmlns:atom="http://www.w3.org/2005/Atom" version="2.0">
<channel>
    <atom:link href="${INDEX_URL}/latest.rss" rel="self" type="application/rss+xml" />
    <link>${INDEX_URL}/latest.json</link>
    <title>Latest Packages</title>
    <description>Latest Packages</description>
    ${entries}
</channel>
</rss>
EOF
}

read -r -d '' Q <<-'EOF'
include "generate-versions";

only_latest | to_rss
EOF
latest_rss="$(jq ${jqFlags} -r "${Q}" "${parsed_file}")"
rss_feed_from_entries "${latest_rss}" "${OUTPUT}/latest.rss"

jq ${jqFlags} 'include "generate-versions"; distros' "${parsed_file}" >${OUTPUT}/distros.json

for distro in $(jq -r '.[]' ${OUTPUT}/distros.json); do
    mkdir -p "${OUTPUT}/${distro}"
    jq --arg distro $distro '.[$distro]' ${OUTPUT}/latest.json >"${OUTPUT}/${distro}/latest.json"
    jq ${jqFlags} --arg distro $distro 'include "generate-versions"; distro_packages($distro)' "${parsed_file}" >"${OUTPUT}/${distro}/packages.json"

    for pkg in $(jq -r '.[]' "${OUTPUT}/${distro}/packages.json"); do
        mkdir -p "${OUTPUT}/${distro}/${pkg}"
        jq --arg distro $distro --arg pkg $pkg '.[$distro][$pkg]' ${OUTPUT}/latest.json >"${OUTPUT}/${distro}/${pkg}/latest.json"
        jq ${jqFlags} --arg distro $distro --arg pkg $pkg 'include "generate-versions"; map(select((.distro == $distro) and (.name == $pkg)) | .version | parse_version | .prefix) | unique' "${parsed_file}" >"${OUTPUT}/${distro}/${pkg}/versions.json"

        for arch in $(jq ${jqFlags} -r 'map(.arch) | .[]' "${OUTPUT}/${distro}/${pkg}/latest.json"); do
            # mkdir handles the case where arch is, for instance, arm/v7.
            # It makes sure the `arm` directory exists.
            # Otherwise for other arches, such as `amd64` it will just mkdir on the `latest/` directory.
            mkdir -p $(dirname "${OUTPUT}/${distro}/${pkg}/latest/${arch}")

            jq ${jqFlags} -r --arg distro $distro --arg pkg $pkg --arg arch $arch 'include "generate-versions"; only_latest | map(select((.distro == $distro) and (.name == $pkg) and (.arch == $arch))) | .[0].uri' "${parsed_file}" >"${OUTPUT}/${distro}/${pkg}/latest/${arch}"

            # Also handle other notations for the same arch, such as `armv7` instead of `arm/v7`.
            if [[ "${arch}" =~ "/" ]]; then
                cp "${OUTPUT}/${distro}/${pkg}/latest/${arch}" "${OUTPUT}/${distro}/${pkg}/latest/${arch//\//}"
            fi

        done

        for v in $(jq -r '.[]' "${OUTPUT}/${distro}/${pkg}/versions.json"); do
            mkdir -p "${OUTPUT}/${distro}/${pkg}/${v}"
            jq ${jqFlags} --arg distro $distro --arg pkg $pkg --arg v $v 'include "generate-versions"; map(select((.distro == $distro) and (.name == $pkg) and (.version | parse_version | .prefix == $v)))' "${parsed_file}" >"${OUTPUT}/${distro}/${pkg}/${v}/index.json"
            jq ${jqFlags} 'include "generate-versions"; only_latest' "${OUTPUT}/${distro}/${pkg}/${v}/index.json" >"${OUTPUT}/${distro}/${pkg}/${v}/latest.json"
        done
    done
done
