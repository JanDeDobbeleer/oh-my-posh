package segments

import (
	"regexp"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"golang.org/x/mod/modfile"
)

type Golang struct {
	language
}

const (
	ParseModFile    properties.Property = "parse_mod_file"
	ParseGoWorkFile properties.Property = "parse_go_work_file"
)

func (g *Golang) Template() string {
	return languageTemplate
}

func (g *Golang) Enabled() bool {
	g.extensions = []string{"*.go", "go.mod", "go.sum", "go.work", "go.work.sum"}
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

// getVersion returns the version of the Go language
// It first checks if the go.mod file is present and if it is, it parses the file to get the version
// If the go.mod file is not present, it checks if the go.work file is present and if it is, it parses the file to get the version
// If neither file is present, it returns an empty string
func (g *Golang) getVersion() (string, error) {
	if !g.props.GetBool(ParseModFile, false) && !g.props.GetBool(ParseGoWorkFile, false) {
		return "", nil
	}

	if g.props.GetBool(ParseModFile, false) {
		gomod, err := g.language.env.HasParentFilePath("go.mod", false)
		if err != nil {
			return "", nil
		}

		contents := g.language.env.FileContent(gomod.Path)
		file, err := modfile.Parse(gomod.Path, []byte(contents), nil)
		if err != nil {
			return "", err
		}

		if file.Go.Version != "" {
			return file.Go.Version, nil
		}
	}

	if g.props.GetBool(ParseGoWorkFile, false) {
		goWork, err := g.language.env.HasParentFilePath("go.work", false)
		if err != nil {
			return "", err
		}

		contents := g.language.env.FileContent(goWork.Path)
		goVersionRegex := regexp.MustCompile(`go (\d(\.\d{1,2})?(\.\d{1,2})?)`)
		if version := goVersionRegex.FindStringSubmatch(contents); version != nil {
			return version[1], nil
		}
	}

	return "", nil
}
