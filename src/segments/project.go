package segments

import (
	"encoding/json"
	"encoding/xml"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type ProjectItem struct {
	Name    string
	File    string
	Fetcher func(item ProjectItem) (string, string)
}

type ProjectData struct {
	Version string
	Name    string
}

// Rust Cargo package
type CargoTOML struct {
	Package ProjectData
}

// Python Poetry package
type PyProjectTOML struct {
	Tool PyProjectToolTOML
}

type PyProjectToolTOML struct {
	Poetry ProjectData
}

type NuSpec struct {
	XMLName  xml.Name `xml:"package"`
	MetaData struct {
		Title   string `xml:"title"`
		Version string `xml:"version"`
	} `xml:"metadata"`
}

type Project struct {
	props properties.Properties
	env   environment.Environment

	projects []*ProjectItem
	Error    string

	ProjectData
}

func (n *Project) Enabled() bool {
	for _, item := range n.projects {
		if n.hasProjectFile(item) {
			n.Version, n.Name = item.Fetcher(*item)
			return len(n.Version) > 0 || len(n.Name) > 0
		}
	}
	return false
}

func (n *Project) Template() string {
	return " {{ if .Error }}{{ .Error }}{{ else }}{{ if .Version }}\uf487 {{.Version}}{{ end }} {{ if .Name }}{{ .Name }}{{ end }}{{ end }} "
}

func (n *Project) Init(props properties.Properties, env environment.Environment) {
	n.props = props
	n.env = env

	n.projects = []*ProjectItem{
		{
			Name:    "node",
			File:    "package.json",
			Fetcher: n.getNodePackage,
		},
		{
			Name:    "cargo",
			File:    "Cargo.toml",
			Fetcher: n.getCargoPackage,
		},
		{
			Name:    "poetry",
			File:    "pyproject.toml",
			Fetcher: n.getPoetryPackage,
		},
		{
			Name:    "php",
			File:    "composer.json",
			Fetcher: n.getNodePackage,
		},
		{
			Name:    "nuspec",
			File:    "*.nuspec",
			Fetcher: n.getNuSpecPackage,
		},
	}
}

func (n *Project) hasProjectFile(p *ProjectItem) bool {
	return n.env.HasFiles(p.File)
}

func (n *Project) getNodePackage(item ProjectItem) (string, string) {
	content := n.env.FileContent(item.File)

	var data ProjectData
	err := json.Unmarshal([]byte(content), &data)
	if err != nil {
		n.Error = err.Error()
		return "", ""
	}

	return data.Version, data.Name
}

func (n *Project) getCargoPackage(item ProjectItem) (string, string) {
	content := n.env.FileContent(item.File)

	var data CargoTOML
	_, err := toml.Decode(content, &data)
	if err != nil {
		n.Error = err.Error()
		return "", ""
	}

	return data.Package.Version, data.Package.Name
}

func (n *Project) getPoetryPackage(item ProjectItem) (string, string) {
	content := n.env.FileContent(item.File)

	var data PyProjectTOML
	_, err := toml.Decode(content, &data)
	if err != nil {
		n.Error = err.Error()
		return "", ""
	}

	return data.Tool.Poetry.Version, data.Tool.Poetry.Name
}

func (n *Project) getNuSpecPackage(item ProjectItem) (string, string) {
	files := n.env.LsDir(n.env.Pwd())
	var content string
	// get the first match only
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".nuspec" {
			content = n.env.FileContent(file.Name())
			break
		}
	}

	var data NuSpec
	err := xml.Unmarshal([]byte(content), &data)
	if err != nil {
		n.Error = err.Error()
		return "", ""
	}

	return data.MetaData.Version, data.MetaData.Title
}
