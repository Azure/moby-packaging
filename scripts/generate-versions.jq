#!/usr/bin/env -S jq -e -f

# https://github.com/tianon/debian-bin/blob/master/jq/dpkg-version.jq
def version_sort_split: (
    [
		if index(":") then . else "0:" + . end # force epoch to be specified
		| if index("-") then . else . + "-0" end # force revision to be specified
		| scan("[0-9]+|[:~-]|[^0-9:~-]+")
		| try tonumber // (
			split("")
			| map(
				# https://metacpan.org/release/GUILLEM/Dpkg-1.20.9/source/lib/Dpkg/Version.pm#L338-350
				if . == "~" then
					-2
				elif . == "-" or . == ":" then # account for me being a little *too* clever (as discovered by using the Dpkg_Version.t test suite)
					-1
				else
					explode[0]
					+ if test("[a-zA-Z]") then 0 else 256 end
				end
			)
		)
	] + [[0]] # gotta add an extra [0] at the end to make sure "1.0" ([1,[302],0]) is higher than "1.0~" ([1,[302],0,[-1]])
);


# Example values:
#   moby-engine_20.10.9+azure-1
#   moby-engine_20.10.17+azure-ubuntu22.04u3
def deb_version(pkg): capture("^\(pkg)_(?<version>\\d+\\.\\d+\\.\\d+)(\\+azure)?\\-(\\w+\\d+(\\.\\d+)?u)?(?<revision>\\d+)");

# Example values:
#   moby-engine-20.10.17+azure-1
def rpm_version(pkg): capture("^\(pkg)\\-(?<version>\\d+\\.\\d+\\.\\d+)(\\+azure)?\\-(?<revision>\\d+).*");

# Example values:
#   moby-engine-20.10.17+azure-u3
#   moby-engine-20.10.2+azure-1
def zip_version(pkg): capture("^\(pkg)\\-(?<version>\\d+\\.\\d+\\.\\d+)(\\+azure)?-\\u?(?<revision>\\d+).*");

def get_version(pkg): if test("\\.rpm$") then rpm_version(pkg) elif test("\\.zip$") then zip_version(pkg) else deb_version(pkg) end;

def parse_version: (
	capture("^(?<major>\\d+)\\.(?<minor>\\d+)\\.(?<patch>\\d+)\\-(?<revision>\\d+)$")
	| {major: .major | tonumber, minor: .minor | tonumber, patch: .patch | tonumber, revision: .revision | tonumber, prefix: "\(.major).\(.minor)"}
);

def join_arch: if .variant == null then .arch else "\(.arch)/\(.variant)" end;

def get_arch: split("_") | {os: .[0], arch: .[1], variant: .[2]} | join_arch;

def get_package: (
    split("/") as $split
    | { name: $split[0], distro: $split[2], version: $split[-1] | get_version($split[0]) | "\(.version)-\(.revision)", uri: "\($URL_PREFIX)/\(.)", arch: $split[3] | get_arch }
);

def version_gte(other): (
    if . == other then true
    elif .major < other.major then false
    elif .major > other.major then true
    elif .minor < other.minor then false
    elif .minor > other.minor then true
    elif .patch < other.patch then false
    elif .patch > other.patch then true
    elif .revision < other.revision then false
    elif .revision > other.revision then true
    else false
    end
);

def get_min: (
	$minVersions[.name] as $min
	| if $min == null then "0.0.0-0" | parse_version else
		$min | parse_version
	end
);

def check_min_version: (
	. as $self
	| $self.version | parse_version | version_gte($self | get_min)
);

def reduce_pkg: reduce(.[]) as $item (
    []; . += [$item.name | get_package + {sha256: $item.sha256}]
) | map(select(check_min_version)) | sort_by(.name, .version | version_sort_split);

def sort_by_version: sort_by( .version | version_sort_split);

def only_latest: (
	sort_by(.version | version_sort_split) | group_by(.distro, .name, .arch) | reduce(.[]) as $group ([]; . += [$group[-1]])
);

def build_index(filter): (
	reduce(.[]) as $item (
		{}; .[$item.distro][$item.name] += [$item] 
	)
);

def distros: map(.distro) | unique;

def only_distro(distro): select(.distro == distro);

def distro_packages(distro): map(only_distro(distro) | .name) | unique;

def build_index: build_index(.);

def to_rss:  reduce(.[]) as $item (
	""; . += "
	<item>
		<guid>\($item.uri)</guid>
		<title>\($item.name) \($item.version) \($item.arch)</title>
		<link>\($item.uri)</link>
		<description>
			Package: \($item.name), Version: \($item.version), Architecture: \($item.arch)
			SHA256 Digest: \($item.sha256)
			\($item.uri)
		</description>
	</item>
"
);
