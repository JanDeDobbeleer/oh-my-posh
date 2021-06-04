package main

import (
	"encoding/json"
)

type nbgv struct {
	props *properties
	env   environmentInfo
	nbgv  *versionInfo
}

type versionInfo struct {
	VersionFileFound             bool   `json:"VersionFileFound"`
	Version                      string `json:"Version"`
	AssemblyVersion              string `json:"AssemblyVersion"`
	AssemblyInformationalVersion string `json:"AssemblyInformationalVersion"`
	NuGetPackageVersion          string `json:"NuGetPackageVersion"`
	ChocolateyPackageVersion     string `json:"ChocolateyPackageVersion"`
	NpmPackageVersion            string `json:"NpmPackageVersion"`
	SimpleVersion                string `json:"SimpleVersion"`
}

func (n *nbgv) enabled() bool {
	nbgv := "nbgv"
	if !n.env.hasCommand(nbgv) {
		return false
	}
	response, err := n.env.runCommand(nbgv, "get-version", "--format=json")
	if err != nil {
		return false
	}
	n.nbgv = new(versionInfo)
	err = json.Unmarshal([]byte(response), n.nbgv)
	if err != nil {
		return false
	}
	return n.nbgv.VersionFileFound
}

func (n *nbgv) string() string {
	segmentTemplate := n.props.getString(SegmentTemplate, "{{ .Version }}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  n.nbgv,
		Env:      n.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (n *nbgv) init(props *properties, env environmentInfo) {
	n.props = props
	n.env = env
}
