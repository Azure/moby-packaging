set -e

fetch_missing_packages() {
    local pck_dir="${1}"
    local ALT_ARCH="$(uname -m)"
    local ARCH=$ALT_ARCH
    case "${ALT_ARCH}" in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    esac
    local pkg_srcs=($(curl -fsSL https://mobyartifacts.azureedge.net/index/mariner2/latest.json | jq -rs --arg arch "${ARCH}" '.[] | map(map(select(.arch == $arch) | .uri)) | flatten | .[]'))
    local build_pkgs="$(ls $pck_dir)"
    mkdir -p "$pck_dir"
    for uri in "${pkg_srcs[@]}"; do
        local pkg_name=$(echo $uri | cut -d '/' -f5)
        local result=$(echo $build_pkgs | grep $pkg_name)
        if [[ -z $result ]]; then
            (cd $pck_dir && curl -O $uri)
        fi
    done
}

fetch_missing_packages "${1}"
