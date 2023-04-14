package build

type Spec struct {
	Pkg      string `yaml:"package"`
	Distro   string `yaml:"distro"`
	Arch     string `yaml:"arch"`
	OS       string `yaml:"os"`
	Repo     string `yaml:"repo"`
	Commit   string `yaml:"commit"`
	Tag      string `yaml:"tag"`
	Revision string `yaml:"revision"`
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
