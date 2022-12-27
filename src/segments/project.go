package segments

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"oh-my-posh/platform"
	"oh-my-posh/properties"
	"oh-my-posh/regex"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type ProjectItem struct {
	Name    string
	Files   []string
	Fetcher func(item ProjectItem) *ProjectData
}

type ProjectData struct {
	Type    string
	Version string
	Name    string
	Target  string
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
	env   platform.Environment

	projects []*ProjectItem
	Error    string

	ProjectData
}

func (n *Project) Enabled() bool {
	for _, item := range n.projects {
		if n.hasProjectFile(item) {
			data := item.Fetcher(*item)
			if data == nil {
				continue
			}
			n.ProjectData = *data
			n.ProjectData.Type = item.Name
			return true
		}
	}
	return n.props.GetBool(properties.AlwaysEnabled, false)
}

func (n *Project) Template() string {
	return " {{ if .Error }}{{ .Error }}{{ else }}{{ if .Version }}\uf487 {{.Version}} {{ end }}{{ if .Name }}{{ .Name }} {{ end }}{{ if .Target }}\uf9fd {{.Target}} {{ end }}{{ end }}" //nolint:lll
}

func (n *Project) Init(props properties.Properties, env platform.Environment) {
	n.props = props
	n.env = env

	n.projects = []*ProjectItem{
		{
			Name:    "node",
			Files:   []string{"package.json"},
			Fetcher: n.getNodePackage,
		},
		{
			Name:    "cargo",
			Files:   []string{"Cargo.toml"},
			Fetcher: n.getCargoPackage,
		},
		{
			Name:    "poetry",
			Files:   []string{"pyproject.toml"},
			Fetcher: n.getPoetryPackage,
		},
		{
			Name:    "php",
			Files:   []string{"composer.json"},
			Fetcher: n.getNodePackage,
		},
		{
			Name:    "nuspec",
			Files:   []string{"*.nuspec"},
			Fetcher: n.getNuSpecPackage,
		},
		{
			Name:    "dotnet",
			Files:   []string{"*.vbproj", "*.fsproj", "*.csproj"},
			Fetcher: n.getDotnetProject,
		},
		{
			Name:    "julia",
			Files:   []string{"JuliaProject.toml", "Project.toml"},
			Fetcher: n.getProjectData,
		},
	}
}

func (n *Project) hasProjectFile(p *ProjectItem) bool {
	for _, file := range p.Files {
		if n.env.HasFiles(file) {
			return true
		}
	}
	return false
}

func (n *Project) getNodePackage(item ProjectItem) *ProjectData {
	content := n.env.FileContent(item.Files[0])

	var data ProjectData
	err := json.Unmarshal([]byte(content), &data)
	if err != nil {
		n.Error = err.Error()
		return nil
	}

	return &data
}

func (n *Project) getCargoPackage(item ProjectItem) *ProjectData {
	content := n.env.FileContent(item.Files[0])

	var data CargoTOML
	_, err := toml.Decode(content, &data)
	if err != nil {
		n.Error = err.Error()
		return nil
	}

	return &ProjectData{
		Version: data.Package.Version,
		Name:    data.Package.Name,
	}
}

func (n *Project) getPoetryPackage(item ProjectItem) *ProjectData {
	content := n.env.FileContent(item.Files[0])

	var data PyProjectTOML
	_, err := toml.Decode(content, &data)
	if err != nil {
		n.Error = err.Error()
		return nil
	}

	return &ProjectData{
		Version: data.Tool.Poetry.Version,
		Name:    data.Tool.Poetry.Name,
	}
}

func (n *Project) getNuSpecPackage(item ProjectItem) *ProjectData {
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
		return nil
	}

	return &ProjectData{
		Version: data.MetaData.Version,
		Name:    data.MetaData.Title,
	}
}

func (n *Project) getDotnetProject(item ProjectItem) *ProjectData {
	files := n.env.LsDir(n.env.Pwd())
	var name string
	var content string
	// get the first match only
	for _, file := range files {
		extension := filepath.Ext(file.Name())
		if extension == ".csproj" || extension == ".fsproj" || extension == ".vbproj" {
			name = strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			content = n.env.FileContent(file.Name())
			break
		}
	}
	// the name of the parameter may differ depending on the version,
	// so instead of xml.Unmarshal() we use regex:
	tag := "(?P<TAG><.*TargetFramework.*>(?P<TFM>.*)</.*TargetFramework.*>)"
	values := regex.FindNamedRegexMatch(tag, content)
	if len(values) == 0 {
		n.Error = errors.New("cannot extract TFM from " + name + " project file").Error()
		return nil
	}
	target := values["TFM"]

	return &ProjectData{
		Target: target,
		Name:   name,
	}
}

func (n *Project) getProjectData(item ProjectItem) *ProjectData {
	content := n.env.FileContent(item.Files[0])

	var data ProjectData
	_, err := toml.Decode(content, &data)
	if err != nil {
		n.Error = err.Error()
		return nil
	}

	return &data
}
