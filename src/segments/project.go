package segments

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"golang.org/x/mod/modfile"

	toml "github.com/pelletier/go-toml/v2"
	yaml "gopkg.in/yaml.v3"
)

type ProjectItem struct {
	Fetcher    func(item ProjectItem) *ProjectData
	Name       string
	OtherFiles bool
	Files      []string
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
	base

	ProjectData
	Error    string
	projects []*ProjectItem
}

func (n *Project) Enabled() bool {
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
			Name:    "mojo",
			Files:   []string{"mojoproject.toml"},
			Fetcher: n.getPythonPackage,
		},
		{
			Name:    "php",
			Files:   []string{"composer.json"},
			Fetcher: n.getNodePackage,
		},
		{
			Name:    "dart",
			Files:   []string{"pubspec.yaml"},
			Fetcher: n.getDartPackage,
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
		{
			Name:       "golang",
			Files:      []string{"go.mod", "go.work"},
			Fetcher:    n.getGoProjectData,
			OtherFiles: true,
		},
	}

	for _, item := range n.projects {
		// allow files override
		property := properties.Property(fmt.Sprintf("%s_files", item.Name))
		item.Files = n.props.GetStringArray(property, item.Files)

		if !item.OtherFiles && !n.hasProjectFile(item) {
			continue
		}

		data := item.Fetcher(*item)
		if data == nil {
			continue
		}

		n.ProjectData = *data
		n.Type = item.Name
		return true
	}

	return n.props.GetBool(properties.AlwaysEnabled, false)
}

func (n *Project) Template() string {
	return " {{ if .Error }}{{ .Error }}{{ else }}{{ if .Version }}\uf487 {{.Version}} {{ end }}{{ if .Name }}{{ .Name }} {{ end }}{{ if .Target }}\uf4de {{.Target}} {{ end }}{{ end }}" //nolint:lll
}

func (n *Project) hasProjectFile(p *ProjectItem) bool {
	return slices.ContainsFunc(p.Files, n.env.HasFiles)
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
	err := toml.Unmarshal([]byte(content), &data)
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
	err := toml.Unmarshal([]byte(content), &data)
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

func (n *Project) getDartPackage(item ProjectItem) *ProjectData {
	content := n.env.FileContent(item.Files[0])
	var data ProjectData
	err := yaml.Unmarshal([]byte(content), &data)
	if err != nil {
		n.Error = err.Error()
		return nil
	}

	return &data
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

func (n *Project) getDotnetProject(item ProjectItem) *ProjectData {
	var name string
	var content string
	var extension string

	files := n.env.LsDir(n.env.Pwd())

	extensions := make([]string, len(item.Files))
	for i, file := range item.Files {
		// Remove leading * and keep only the extension
		extensions[i] = strings.TrimPrefix(file, "*")
	}

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

	if target == "" {
		log.Error(fmt.Errorf("cannot extract TFM from %s project file", name))
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

	if content == "" {
		return nil
	}

	data := &ProjectData{}
	lines := strings.SplitSeq(content, "\n")

	for line := range lines {
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

func (n *Project) getGoProjectData(item ProjectItem) *ProjectData {
	var content string
	var match *runtime.FileInfo
	var err error

	if item.Name != "golang" {
		return nil
	}

	for _, file := range item.Files {
		match, err = n.env.HasParentFilePath(file, true)
		if err != nil {
			if errors.Is(err, runtime.ErrNoMatchAtRootLevel) {
				// couldn't find the file in any parent directory, move on to the next file
				continue
			}

			n.Error = err.Error()
			return nil
		}

		if match != nil {
			break
		}
	}

	if match == nil {
		return nil
	}

	content = n.env.FileContent(match.Path)

	if content == "" {
		return nil
	}

	var data ProjectData
	if strings.HasSuffix(match.Path, "go.work") {
		f, err := modfile.ParseWork(match.Path, []byte(content), nil)
		if err != nil {
			n.Error = err.Error()
			return nil
		}

		data.Target = f.Go.Version

		return &data
	}

	f, err := modfile.Parse(match.Path, []byte(content), nil)
	if err != nil {
		n.Error = err.Error()
		return nil
	}

	data.Target = f.Go.Version
	return &data
}

func (n *Project) getProjectData(item ProjectItem) *ProjectData {
	content := n.env.FileContent(item.Files[0])

	var data ProjectData
	err := toml.Unmarshal([]byte(content), &data)
	if err != nil {
		n.Error = err.Error()
		return nil
	}

	return &data
}
