package build

type Spec struct {
	Pkg      string `json:"package"`
	Distro   string `json:"distro"`
	Arch     string `json:"arch"`
	OS       string `json:"os"`
	Repo     string `json:"repo"`
	Commit   string `json:"commit"`
	Tag      string `json:"tag"`
	Revision string `json:"revision"`
}

// values from moby trigger
// "arch": "arm/v7",
// "commit": "6c0083137f9dc5817d5bb11d35b9a883f7f37211",
// "distro": "buster",
// "github_repo": "containerd/containerd",
// "os": "linux",
// "package_name": "moby-engine",
// "pmc_repo": "TODO",
// "release_main": "true",
// "release_testing": "false",
// "tag": "1.5.17"
