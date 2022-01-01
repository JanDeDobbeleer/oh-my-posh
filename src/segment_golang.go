package main

import (
	"golang.org/x/mod/modfile"
)

type golang struct {
	language
}

const (
	ParseModFile Property = "parse_mod_file"
)

func (g *golang) string() string {
	return g.language.string()
}

func (g *golang) init(props Properties, env environmentInfo) {
	g.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.go", "go.mod"},
		commands: []*cmd{
			{
				regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+)(.(?P<patch>[0-9]+))?))`,
				getVersion: g.getVersion,
			},
			{
				executable: "go",
				args:       []string{"version"},
				regex:      `(?:go(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+)(.(?P<patch>[0-9]+))?)))`,
			},
		},
		versionURLTemplate: "[%s](https://golang.org/doc/go%s.%s)",
	}
}

func (g *golang) getVersion() (string, error) {
	if !g.props.getBool(ParseModFile, false) {
		return "", nil
	}
	gomod, err := g.language.env.hasParentFilePath("go.mod")
	if err != nil {
		return "", nil
	}
	contents := g.language.env.getFileContent(gomod.path)
	file, err := modfile.Parse(gomod.path, []byte(contents), nil)
	if err != nil {
		return "", err
	}
	return file.Go.Version, nil
}

func (g *golang) enabled() bool {
	return g.language.enabled()
}
