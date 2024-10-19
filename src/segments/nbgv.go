package segments

import (
	"encoding/json"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Nbgv struct {
	props properties.Properties
	env   runtime.Environment

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

func (n *Nbgv) Init(props properties.Properties, env runtime.Environment) {
	n.props = props
	n.env = env
}
