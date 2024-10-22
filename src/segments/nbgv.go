package segments

import (
	"encoding/json"
)

type Nbgv struct {
	base

	VersionInfo
}

type VersionInfo struct {
	Version                      string `json:"Version"`
	AssemblyVersion              string `json:"AssemblyVersion"`
	AssemblyInformationalVersion string `json:"AssemblyInformationalVersion"`
	NuGetPackageVersion          string `json:"NuGetPackageVersion"`
	ChocolateyPackageVersion     string `json:"ChocolateyPackageVersion"`
	NpmPackageVersion            string `json:"NpmPackageVersion"`
	SimpleVersion                string `json:"SimpleVersion"`
	VersionFileFound             bool   `json:"VersionFileFound"`
}

func (n *Nbgv) Template() string {
	return " {{ .Version }} "
}

func (n *Nbgv) Enabled() bool {
	nbgv := "nbgv"
	if !n.env.HasCommand(nbgv) {
		return false
	}
	response, err := n.env.RunCommand(nbgv, "get-version", "--format=json")
	if err != nil {
		return false
	}
	n.VersionInfo = VersionInfo{}
	err = json.Unmarshal([]byte(response), &n.VersionInfo)
	if err != nil {
		return false
	}
	return n.VersionInfo.VersionFileFound
}
