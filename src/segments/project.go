package segments

import (
	"encoding/json"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type ProjectItem struct {
	Name    string
	File    string
	Fetcher func(item ProjectItem) string
}

type NodePackageJSON struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type Project struct {
	props properties.Properties
	env   environment.Environment

	projects []*ProjectItem
	Version  string
	Error    string
}

func (n *Project) Enabled() bool {
	var enabled = false
	for _, item := range n.projects {
		if !enabled {
			enabled = n.hasProjectFile(item)
		}
	}

	return enabled
}

func (n *Project) Template() string {
	return " {{ if .Error }}{{ .Error }}{{ else }}{{ if .Version }}\uf487 {{.Version}}{{ end }}{{ end }} "
}

func (n *Project) Init(props properties.Properties, env environment.Environment) {
	n.props = props
	n.env = env

	n.projects = []*ProjectItem{
		{
			Name:    "node",
			File:    "package.json",
			Fetcher: n.getNodePackageVersion,
		},
	}

	n.Version = ""
	for _, item := range n.projects {
		if n.hasProjectFile(item) {
			n.Version = item.Fetcher(*item)
			break
		}
	}
}

func (n *Project) hasProjectFile(p *ProjectItem) bool {
	return n.env.HasFiles(p.File)
}

func (n *Project) getNodePackageVersion(item ProjectItem) string {
	content := n.env.FileContent(item.File)

	var data NodePackageJSON
	err := json.Unmarshal([]byte(content), &data)
	if err != nil {
		n.Error = err.Error()
		return ""
	}

	return data.Version
}
