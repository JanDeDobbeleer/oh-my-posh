package segments

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"golang.org/x/exp/slices"

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

// Python package
type PyProjectTOML struct {
	Project ProjectData
	Tool    PyProjectToolTOML
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
	return " {{ if .Error }}{{ .Error }}{{ else }}{{ if .Version }}\uf487 {{.Version}} {{ end }}{{ if .Name }}{{ .Name }} {{ end }}{{ if .Target }}\uf4de {{.Target}} {{ end }}{{ end }}" //nolint:lll
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
			Name:    "python",
			Files:   []string{"pyproject.toml"},
			Fetcher: n.getPythonPackage,
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
			Files:   []string{"*.sln", "*.slnf", "*.vbproj", "*.fsproj", "*.csproj"},
			Fetcher: n.getDotnetProject,
		},
		{
			Name:    "julia",
			Files:   []string{"JuliaProject.toml", "Project.toml"},
			Fetcher: n.getProjectData,
		},
		{
			Name:    "powershell",
			Files:   []string{"*.psd1"},
			Fetcher: n.getPowerShellModuleData,
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

func (n *Project) getPythonPackage(item ProjectItem) *ProjectData {
	content := n.env.FileContent(item.Files[0])

	var data PyProjectTOML
	_, err := toml.Decode(content, &data)
	if err != nil {
		n.Error = err.Error()
		return nil
	}

	if len(data.Tool.Poetry.Version) != 0 || len(data.Tool.Poetry.Name) != 0 {
		return &ProjectData{
			Version: data.Tool.Poetry.Version,
			Name:    data.Tool.Poetry.Name,
		}
	}
	return &ProjectData{
		Version: data.Project.Version,
		Name:    data.Project.Name,
	}
}

func (n *Project) getNuSpecPackage(_ ProjectItem) *ProjectData {
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

func (n *Project) getDotnetProject(_ ProjectItem) *ProjectData {
	var name string
	var content string
	var extension string

	extensions := []string{".sln", ".slnf", ".csproj", ".fsproj", ".vbproj"}
	files := n.env.LsDir(n.env.Pwd())

	// get the first match only
	for _, file := range files {
		extension = filepath.Ext(file.Name())
		if slices.Contains(extensions, extension) {
			name = strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			content = n.env.FileContent(file.Name())
			break
		}
	}

	// the name of the parameter may differ depending on the version,
	// so instead of xml.Unmarshal() we use regex:
	var target string
	tag := "(?P<TAG><.*TargetFramework.*>(?P<TFM>.*)</.*TargetFramework.*>)"

	values := regex.FindNamedRegexMatch(tag, content)
	if len(values) != 0 {
		target = values["TFM"]
	}

	if len(target) == 0 {
		n.env.Error(fmt.Errorf("cannot extract TFM from %s project file", name))
	}

	return &ProjectData{
		Target: target,
		Name:   name,
	}
}

func (n *Project) getPowerShellModuleData(_ ProjectItem) *ProjectData {
	files := n.env.LsDir(n.env.Pwd())
	var content string
	// get the first match only
	// excluding PSScriptAnalyzerSettings.psd1
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".psd1" && file.Name() != "PSScriptAnalyzerSettings.psd1" {
			content = n.env.FileContent(file.Name())
			break
		}
	}

	if len(content) == 0 {
		return nil
	}

	data := &ProjectData{}
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		splitted := strings.SplitN(line, "=", 2)
		if len(splitted) < 2 {
			continue
		}
		key := strings.TrimSpace(splitted[0])
		value := strings.TrimSpace(splitted[1])
		value = strings.Trim(value, "'\"")

		switch key {
		case "ModuleVersion":
			data.Version = value
		case "RootModule":
			data.Name = strings.TrimRight(value, ".psm1")
		}
	}

	return data
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
