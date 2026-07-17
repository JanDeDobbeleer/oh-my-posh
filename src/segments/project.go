package segments

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gookit/goutil/jsonutil"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	toml "github.com/pelletier/go-toml/v2"
	yaml "go.yaml.in/yaml/v3"
)

const (
	ResolveTargetFromSolution options.Option = "resolve_target_from_solution"
	SolutionSearchDepth       options.Option = "solution_search_depth"
	Priority                  options.Option = "priority"
)

type ProjectItem struct {
	Fetcher func(item ProjectItem) *ProjectData
	Name    string
	Files   []string
}

type ProjectData struct {
	Type    string
	Version string
	Name    string
	Target  string
}

// Lake file package
type LakeFileTOML struct {
	Name string
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
	Base

	ProjectData
	Error    string
	projects []*ProjectItem
}

func (n *Project) Enabled() bool {
	n.projects = []*ProjectItem{
		{
			Name:    nodeToolName,
			Files:   []string{fileName},
			Fetcher: n.getNodePackage,
		},
		{
			Name:    denoToolName,
			Files:   []string{"deno.json", "deno.jsonc"},
			Fetcher: n.getDenoPackage,
		},
		{
			Name:    "jsr",
			Files:   []string{"jsr.json", "jsr.jsonc"},
			Fetcher: n.getJsrPackage,
		},
		{
			Name:    "cargo",
			Files:   []string{"Cargo.toml"},
			Fetcher: n.getCargoPackage,
		},
		{
			Name:    pythonToolName,
			Files:   []string{"pyproject.toml"},
			Fetcher: n.getPythonPackage,
		},
		{
			Name:    mojoToolName,
			Files:   []string{"mojoproject.toml"},
			Fetcher: n.getPythonPackage,
		},
		{
			Name:    phpToolName,
			Files:   []string{"composer.json"},
			Fetcher: n.getNodePackage,
		},
		{
			Name:    dartToolName,
			Files:   []string{pubspecFileName},
			Fetcher: n.getDartPackage,
		},
		{
			Name:    "nuspec",
			Files:   []string{"*.nuspec"},
			Fetcher: n.getNuSpecPackage,
		},
		{
			Name:    dotnetToolName,
			Files:   []string{"*.sln", "*.slnf", "*.slnx", "*.vbproj", "*.fsproj", "*.csproj"},
			Fetcher: n.getDotnetProject,
		},
		{
			Name:    juliaToolName,
			Files:   []string{"JuliaProject.toml", "Project.toml"},
			Fetcher: n.getProjectData,
		},
		{
			Name:    "lake",
			Files:   []string{"lakefile.lean", "lakefile.toml"},
			Fetcher: n.getLakePackage,
		},
		{
			Name:    "powershell",
			Files:   []string{"*.psd1"},
			Fetcher: n.getPowerShellModuleData,
		},
	}

	if priority := n.options.StringArray(Priority, nil); len(priority) != 0 {
		n.projects = reorderByPriority(n.projects, priority)
	}

	for _, item := range n.projects {
		// allow files override
		property := options.Option(fmt.Sprintf("%s_files", item.Name))
		item.Files = n.options.StringArray(property, item.Files)

		if !n.hasProjectFile(item) {
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

	return n.options.Bool(options.AlwaysEnabled, false)
}

func (n *Project) Template() string {
	return " {{ if .Error }}{{ .Error }}{{ else }}{{ if .Version }}\uf487 {{.Version}} {{ end }}{{ if .Name }}{{ .Name }} {{ end }}{{ if .Target }}\uf4de {{.Target}} {{ end }}{{ end }}" //nolint:lll
}

func (n *Project) hasProjectFile(p *ProjectItem) bool {
	return slices.ContainsFunc(p.Files, n.env.HasFiles)
}

// reorderByPriority moves the items named in priority to the front of items, in the
// given order. Items not named in priority keep their original relative order and are
// appended afterward; names in priority that don't match any item are ignored.
func reorderByPriority(items []*ProjectItem, priority []string) []*ProjectItem {
	byName := make(map[string]*ProjectItem, len(items))
	for _, item := range items {
		byName[item.Name] = item
	}

	promoted := make(map[string]bool, len(priority))
	ordered := make([]*ProjectItem, 0, len(items))

	for _, name := range priority {
		item, ok := byName[name]
		if !ok || promoted[name] {
			continue
		}

		ordered = append(ordered, item)
		promoted[name] = true
	}

	for _, item := range items {
		if promoted[item.Name] {
			continue
		}

		ordered = append(ordered, item)
	}

	return ordered
}

func (n *Project) getNodePackage(item ProjectItem) *ProjectData {
	return n.getJSONPackage(item, false)
}

func (n *Project) getDenoPackage(item ProjectItem) *ProjectData {
	data := n.getJSONPackage(item, true)
	if data == nil {
		return nil
	}

	// Deno projects prefer to publish via JSR; merge JSR metadata when available.
	jsrFile := n.firstExistingFile([]string{"jsr.json", "jsr.jsonc"})
	if len(jsrFile) == 0 {
		return data
	}

	jsrData, err := n.parseJSONPackage(jsrFile, true)
	if err != nil {
		log.Error(err)
		return data
	}

	if len(jsrData.Version) != 0 {
		data.Version = jsrData.Version
	}

	if len(jsrData.Name) != 0 {
		data.Name = jsrData.Name
	}

	return data
}

func (n *Project) getJsrPackage(item ProjectItem) *ProjectData {
	return n.getJSONPackage(item, true)
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

	if target == "" && (extension == ".sln" || extension == ".slnx") && n.options.Bool(ResolveTargetFromSolution, true) {
		maxDepth := n.options.Int(SolutionSearchDepth, 2)
		if projContent := n.findProjectFile(files, maxDepth); projContent != "" {
			values = regex.FindNamedRegexMatch(tag, projContent)
			if len(values) != 0 {
				target = values["TFM"]
			}
		}
	}

	// mirror MSBuild's implicit import of Directory.Build.props when the
	// project/solution itself does not define a TargetFramework
	if target == "" {
		if props, err := n.env.HasParentFilePath("Directory.Build.props", false); err == nil {
			propsContent := n.env.FileContent(props.Path)
			values = regex.FindNamedRegexMatch(tag, propsContent)
			if len(values) != 0 {
				target = values["TFM"]
			}
		}
	}

	if target == "" {
		log.Error(fmt.Errorf("cannot extract TFM from %s project file", name))
	}

	return &ProjectData{
		Target: target,
		Name:   name,
	}
}

// findProjectFile scans rootEntries and their subdirectories breadth-first,
// up to maxDepth levels deep, and returns the content of the first
// .csproj/.fsproj/.vbproj file found. Paths are kept relative to pwd so
// FileContent resolves them the same way the caller does.
func (n *Project) findProjectFile(rootEntries []fs.DirEntry, maxDepth int) string {
	projectExts := []string{".csproj", ".fsproj", ".vbproj"}
	pwd := n.env.Pwd()

	var dirs []string
	for _, entry := range rootEntries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
			continue
		}

		// a project file can live next to the solution
		if slices.Contains(projectExts, filepath.Ext(entry.Name())) {
			return n.env.FileContent(entry.Name())
		}
	}

	for depth := 1; depth <= maxDepth && len(dirs) > 0; depth++ {
		var next []string

		for _, dir := range dirs {
			for _, entry := range n.env.LsDir(filepath.Join(pwd, dir)) {
				if entry.IsDir() {
					next = append(next, filepath.Join(dir, entry.Name()))
					continue
				}

				if slices.Contains(projectExts, filepath.Ext(entry.Name())) {
					return n.env.FileContent(filepath.Join(dir, entry.Name()))
				}
			}
		}

		dirs = next
	}

	return ""
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
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
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

func (n *Project) getLakePackage(item ProjectItem) *ProjectData {
	file := n.firstExistingFile(item.Files)
	if len(file) == 0 {
		return nil
	}

	if strings.HasSuffix(file, ".lean") {
		return n.getLakeLeanPackage(file)
	}

	return n.getLakeTomlPackage(file)
}

func (n *Project) getLakeLeanPackage(file string) *ProjectData {
	content := n.env.FileContent(file)

	match := regex.FindNamedRegexMatch(`package\s+(?P<NAME>.+?)\s+where`, content)
	name, ok := match["NAME"]
	if !ok || len(name) == 0 {
		return nil
	}

	// Strip guillemets (« U+00AB and » U+00BB) if present
	name = strings.Trim(name, "\u00AB\u00BB")

	return &ProjectData{
		Name: strings.TrimSpace(name),
	}
}

func (n *Project) getLakeTomlPackage(file string) *ProjectData {
	content := n.env.FileContent(file)

	var data LakeFileTOML
	err := toml.Unmarshal([]byte(content), &data)
	if err != nil {
		n.Error = err.Error()
		return nil
	}

	return &ProjectData{
		Name: data.Name,
	}
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

func (n *Project) getJSONPackage(item ProjectItem, allowJSONC bool) *ProjectData {
	file := n.firstExistingFile(item.Files)
	if len(file) == 0 {
		return nil
	}

	data, err := n.parseJSONPackage(file, allowJSONC)
	if err != nil {
		n.Error = err.Error()
		return nil
	}

	return data
}

func (n *Project) firstExistingFile(files []string) string {
	for _, file := range files {
		if !n.env.HasFiles(file) {
			continue
		}
		return file
	}

	return ""
}

func (n *Project) parseJSONPackage(file string, allowJSONC bool) (*ProjectData, error) {
	content := n.env.FileContent(file)
	if allowJSONC && filepath.Ext(file) == ".jsonc" {
		content = jsonutil.StripComments(content)
	}

	var data ProjectData
	err := json.Unmarshal([]byte(content), &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
