package main

import "regexp"

type golang struct {
	props         *properties
	env           environmentInfo
	golangVersion string
}

func (g *golang) string() string {
	if g.props.getBool(DisplayVersion, true) {
		return g.golangVersion
	}
	return ""
}

func (g *golang) init(props *properties, env environmentInfo) {
	g.props = props
	g.env = env
}

func (g *golang) enabled() bool {
	if !g.env.hasFiles("*.go") {
		return false
	}
	if !g.env.hasCommand("go") {
		return false
	}
	versionInfo, _ := g.env.runCommand("go", "version")
	r := regexp.MustCompile(`go(?P<version>[0-9]+.[0-9]+.[0-9]+)`)
	values := groupDict(r, versionInfo)
	g.golangVersion = values["version"]
	return true
}
