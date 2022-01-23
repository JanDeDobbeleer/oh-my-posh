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
	segmentTemplate := g.language.props.getString(SegmentTemplate, "{{ if .Error }}{{ .Error }}{{ else }}{{ .Full }}{{ end }}")
	return g.language.string(segmentTemplate, g)
}

func (g *golang) init(props Properties, env Environment) {
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
		versionURLTemplate: "https://golang.org/doc/go{{ .Major }}.{{ .Minor }})",
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
