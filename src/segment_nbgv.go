package main

import (
	"encoding/json"
	"oh-my-posh/environment"
)

type nbgv struct {
	props Properties
	env   environment.Environment

	VersionInfo
}

type VersionInfo struct {
	VersionFileFound             bool   `json:"VersionFileFound"`
	Version                      string `json:"Version"`
	AssemblyVersion              string `json:"AssemblyVersion"`
	AssemblyInformationalVersion string `json:"AssemblyInformationalVersion"`
	NuGetPackageVersion          string `json:"NuGetPackageVersion"`
	ChocolateyPackageVersion     string `json:"ChocolateyPackageVersion"`
	NpmPackageVersion            string `json:"NpmPackageVersion"`
	SimpleVersion                string `json:"SimpleVersion"`
}

func (n *nbgv) template() string {
	return "{{ .Version }}"
}

func (n *nbgv) enabled() bool {
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

func (n *nbgv) init(props Properties, env environment.Environment) {
	n.props = props
	n.env = env
}
