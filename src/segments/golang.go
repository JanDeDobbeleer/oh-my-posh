package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"golang.org/x/mod/modfile"
)

type Golang struct {
	language
}

const (
	ParseModFile  properties.Property = "parse_mod_file"
	ParseWorkFile properties.Property = "parse_work_file"
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
	if g.props.GetBool(ParseModFile, false) {
		return g.parseModFile()
	}

	if g.props.GetBool(ParseWorkFile, false) {
		return g.parseWorkFile()
	}

	return "", nil
}

func (g *Golang) parseModFile() (string, error) {
	gomod, err := g.env.HasParentFilePath("go.mod", false)
	if err != nil {
		return "", err
	}

	contents := g.env.FileContent(gomod.Path)
	file, err := modfile.Parse(gomod.Path, []byte(contents), nil)
	if err != nil {
		return "", err
	}

	if file.Go.Version != "" {
		return file.Go.Version, nil
	}

	// ignore when no version is found in go.mod file
	return "", nil
}

func (g *Golang) parseWorkFile() (string, error) {
	goWork, err := g.env.HasParentFilePath("go.work", false)
	if err != nil {
		return "", err
	}

	contents := g.env.FileContent(goWork.Path)
	version, _ := regex.FindStringMatch(`go (\d(\.\d{1,2})?(\.\d{1,2})?)`, contents, 1)
	if len(version) > 0 {
		return version, nil
	}

	// ignore when no version is found in go.work file
	return "", nil
}
