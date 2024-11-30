package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"

	"golang.org/x/mod/modfile"
)

type Golang struct {
	language
}

const (
	ParseModFile properties.Property = "parse_mod_file"
)

func (g *Golang) Template() string {
	return languageTemplate
}

func (g *Golang) Enabled() bool {
	g.extensions = []string{"*.go", "go.mod"}
	g.commands = []*cmd{
		{
			regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+)(.(?P<patch>[0-9]+))?))`,
			getVersion: g.getVersion,
		},
		{
			executable: "go",
			args:       []string{"version"},
			regex:      `(?:go(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+)(.(?P<patch>[0-9]+))?)))`,
		},
	}
	g.versionURLTemplate = "https://golang.org/doc/go{{ .Major }}.{{ .Minor }}"

	return g.language.Enabled()
}

func (g *Golang) getVersion() (string, error) {
	if !g.props.GetBool(ParseModFile, false) {
		return "", nil
	}

	gomod, err := g.language.env.HasParentFilePath("go.mod", false)
	if err != nil {
		return "", nil
	}

	contents := g.language.env.FileContent(gomod.Path)
	file, err := modfile.Parse(gomod.Path, []byte(contents), nil)
	if err != nil {
		return "", err
	}

	return file.Go.Version, nil
}
