package archive

type Spec struct {
	Pkg      string `json:"package"`
	Distro   string `json:"distro"`
	Arch     string `json:"arch"`
	Repo     string `json:"repo"`
	Commit   string `json:"commit"`
	Tag      string `json:"tag"`
	Revision string `json:"revision"`
}
