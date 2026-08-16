package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

type Golang struct {
	Language
}

const (
	ParseModFile    options.Option = "parse_mod_file"
	ParseGoWorkFile options.Option = "parse_go_work_file"
)

func (g *Golang) Template() string {
	return languageTemplate
}

func (g *Golang) Enabled() bool {
	g.loadSpec()

	return g.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (g *Golang) Activation() Activation {
	g.loadSpec()

	return g.activation()
}

func (g *Golang) loadSpec() {
	g.extensions = []string{"*.go", "go.mod", "go.sum", "go.work", "go.work.sum"}
	g.tooling = map[string]*cmd{
		"mod": {
			regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+)(.(?P<patch>[0-9]+))?))`,
			getVersion: g.getVersion,
		},
		"go": {
			executable: "go",
			args:       []string{versionArg},
			regex:      `(?:go(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+)(.(?P<patch>[0-9]+))?)))`,
		},
	}
	g.defaultTooling = []string{"mod", "go"}
	g.versionURLTemplate = "https://golang.org/doc/go{{ .Major }}.{{ .Minor }}"
}

func (g *Golang) getVersion() (string, error) {
	if g.options.Bool(ParseModFile, false) {
		return g.parseModFile()
	}

	if g.options.Bool(ParseGoWorkFile, false) {
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

	// the go directive is a top-level "go <version>" line; module paths in
	// require blocks always contain a slash or dot so they never match
	for line := range strings.Lines(contents) {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "go" && fields[1][0] >= '0' && fields[1][0] <= '9' {
			return fields[1], nil
		}
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
