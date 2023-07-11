set -e

fetch_missing_packages() {
    local pck_dir="${1}"
    local ALT_ARCH="$(uname -m)"
    local ARCH=$ALT_ARCH
    case "${ALT_ARCH}" in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    esac
    local pkg_srcs=($(curl -fsSL --retry 50 -Y 665600 https://mobyartifacts.azureedge.net/index/mariner2/latest.json | jq -rs --arg arch "${ARCH}" '.[] | map(map(select(.arch == $arch) | .uri)) | flatten | .[]'))
    mkdir -p "$pck_dir"
    local build_pkgs="$(ls $pck_dir)"
    for uri in "${pkg_srcs[@]}"; do
        local pkg_name=$(echo $uri | cut -d '/' -f5)
        local result=$(echo $build_pkgs | grep $pkg_name)

        offset=0
        if [ -f "${pck_dir}/${pkg_name}" ]; then
            offset="$(stat --printf="%s" "${pck_dir}/${pkg_name}")"
        fi
        # -Y sets a speed limit, when the download speed is lower than that limit it causes the download to fail
        #   The value for -Y is in bytes per second. 665600 is 650KB/s
        # At that time it will be retried (do to --retry)
        # -C sets the offset to continue the download from (if the file already exists)
        #   The offset is calculated above.
        # This all makes the download more robust, particularly because we seem to hit issues in mariner.
        curl -Y 665600 --retry 5 --output-dir $pck_dir -O -fSL -C "${offset}" $uri
    done
}

fetch_missing_packages "${1}"
