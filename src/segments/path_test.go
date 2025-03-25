package segments

import (
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

const (
	homeDir        = "/home/someone"
	homeDirWindows = "C:\\Users\\someone"
	fooBarMan      = "\\foo\\bar\\man"
	abc            = "/abc"
	abcd           = "/a/b/c/d"
	cdefg          = "/c/d/e/f/g"
)

func renderTemplateNoTrimSpace(env *mock.Environment, segmentTemplate string, context any) string {
	env.On("Shell").Return("foo")

	if template.Cache == nil {
		template.Cache = &cache.Template{}
	}
	template.Init(env, nil)

	tmpl := &template.Text{
		Template: segmentTemplate,
		Context:  context,
	}

	text, err := tmpl.Render()
	if err != nil {
		return err.Error()
	}

	return text
}

func renderTemplate(env *mock.Environment, segmentTemplate string, context any) string {
	return strings.TrimSpace(renderTemplateNoTrimSpace(env, segmentTemplate, context))
}

type testParentCase struct {
	Case                string
	Expected            string
	HomePath            string
	Pwd                 string
	GOOS                string
	PathSeparator       string
	FolderSeparatorIcon string
}

func TestParent(t *testing.T) {
	for _, tc := range testParentCases {
		env := new(mock.Environment)
		env.On("Home").Return(tc.HomePath)
		env.On("Pwd").Return(tc.Pwd)
		env.On("Flags").Return(&runtime.Flags{})
		env.On("Shell").Return(shell.GENERIC)
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("GOOS").Return(tc.GOOS)

		props := properties.Map{
			FolderSeparatorIcon: tc.FolderSeparatorIcon,
		}

		path := &Path{}
		path.Init(props, env)

		path.setPaths()

		got := path.Parent()
		assert.EqualValues(t, tc.Expected, got, tc.Case)
	}
}

type testAgnosterPathStyleCase struct {
	CygpathError        error
	GOOS                string
	Shell               string
	Pswd                string
	Pwd                 string
	PathSeparator       string
	HomeIcon            string
	HomePath            string
	Style               string
	FolderSeparatorIcon string
	Cygpath             string
	Expected            string
	MaxDepth            int
	MaxWidth            int
	HideRootLocation    bool
	Cygwin              bool
	DisplayRoot         bool
}

func TestAgnosterPathStyles(t *testing.T) {
	for _, tc := range testAgnosterPathStyleCases {
		env := new(mock.Environment)
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("Home").Return(tc.HomePath)
		env.On("Pwd").Return(tc.Pwd)
		env.On("GOOS").Return(tc.GOOS)
		env.On("IsCygwin").Return(tc.Cygwin)
		env.On("StackCount").Return(0)
		env.On("IsWsl").Return(false)
		args := &runtime.Flags{
			PSWD: tc.Pswd,
		}
		env.On("Flags").Return(args)

		if len(tc.Shell) == 0 {
			tc.Shell = shell.PWSH
		}
		env.On("Shell").Return(tc.Shell)

		displayCygpath := tc.Cygwin
		if displayCygpath {
			env.On("RunCommand", "cygpath", []string{"-u", tc.Pwd}).Return(tc.Cygpath, tc.CygpathError)
			env.On("RunCommand", "cygpath", testify_.Anything).Return("brrrr", nil)
		}

		props := properties.Map{
			FolderSeparatorIcon: tc.FolderSeparatorIcon,
			properties.Style:    tc.Style,
			MaxDepth:            tc.MaxDepth,
			MaxWidth:            tc.MaxWidth,
			HideRootLocation:    tc.HideRootLocation,
			DisplayCygpath:      displayCygpath,
			DisplayRoot:         tc.DisplayRoot,
		}

		path := &Path{}
		path.Init(props, env)

		path.setPaths()
		path.setStyle()
		got := renderTemplateNoTrimSpace(env, "{{ .Path }}", path)
		assert.Equal(t, tc.Expected, got)
	}
}

type testFullAndFolderPathCase struct {
	Style                  string
	HomePath               string
	FolderSeparatorIcon    string
	Pwd                    string
	Pswd                   string
	Expected               string
	GOOS                   string
	PathSeparator          string
	Template               string
	StackCount             int
	DisableMappedLocations bool
}

func TestFullAndFolderPath(t *testing.T) {
	for _, tc := range testFullAndFolderPathCases {
		env := new(mock.Environment)
		if len(tc.PathSeparator) == 0 {
			tc.PathSeparator = "/"
		}
		env.On("PathSeparator").Return(tc.PathSeparator)
		if tc.GOOS == runtime.WINDOWS {
			env.On("Home").Return(homeDirWindows)
		} else {
			env.On("Home").Return(homeDir)
		}
		env.On("Pwd").Return(tc.Pwd)
		env.On("GOOS").Return(tc.GOOS)
		env.On("StackCount").Return(tc.StackCount)
		env.On("IsWsl").Return(false)
		args := &runtime.Flags{
			PSWD: tc.Pswd,
		}
		env.On("Flags").Return(args)
		env.On("Shell").Return(shell.GENERIC)
		if len(tc.Template) == 0 {
			tc.Template = "{{ if gt .StackCount 0 }}{{ .StackCount }} {{ end }}{{ .Path }}"
		}
		props := properties.Map{
			properties.Style: tc.Style,
		}
		if tc.FolderSeparatorIcon != "" {
			props[FolderSeparatorIcon] = tc.FolderSeparatorIcon
		}
		if tc.DisableMappedLocations {
			props[MappedLocationsEnabled] = false
		}

		path := &Path{
			StackCount: env.StackCount(),
		}
		path.Init(props, env)

		path.setPaths()
		path.setStyle()
		got := renderTemplateNoTrimSpace(env, tc.Template, path)
		assert.Equal(t, tc.Expected, got)
	}
}

type testFullPathCustomMappedLocationsCase struct {
	Pwd             string
	MappedLocations map[string]string
	GOOS            string
	PathSeparator   string
	Expected        string
}

func TestFullPathCustomMappedLocations(t *testing.T) {
	for _, tc := range testFullPathCustomMappedLocationsCases {
		env := new(mock.Environment)
		env.On("Home").Return(homeDir)
		env.On("Pwd").Return(tc.Pwd)

		if len(tc.GOOS) == 0 {
			tc.GOOS = runtime.DARWIN
		}

		env.On("GOOS").Return(tc.GOOS)

		if len(tc.PathSeparator) == 0 {
			tc.PathSeparator = "/"
		}

		env.On("PathSeparator").Return(tc.PathSeparator)
		args := &runtime.Flags{
			PSWD: tc.Pwd,
		}

		env.On("Flags").Return(args)
		env.On("Shell").Return(shell.GENERIC)
		env.On("Getenv", "HOME").Return(homeDir)

		template.Cache = new(cache.Template)
		template.Init(env, nil)

		props := properties.Map{
			properties.Style:       Full,
			MappedLocationsEnabled: false,
			MappedLocations:        tc.MappedLocations,
		}

		path := &Path{}
		path.Init(props, env)

		path.setPaths()
		path.setStyle()

		got := renderTemplateNoTrimSpace(env, "{{ .Path }}", path)
		assert.Equal(t, tc.Expected, got)
	}
}

type testAgnosterPathCase struct {
	Case           string
	Expected       string
	Home           string
	PWD            string
	GOOS           string
	PathSeparator  string
	Cycle          []string
	ColorSeparator bool
}

func TestAgnosterPath(t *testing.T) {
	for _, tc := range testAgnosterPathCases {
		env := new(mock.Environment)
		env.On("Home").Return(tc.Home)
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("Pwd").Return(tc.PWD)
		env.On("GOOS").Return(tc.GOOS)
		args := &runtime.Flags{
			PSWD: tc.PWD,
		}
		env.On("Flags").Return(args)
		env.On("Shell").Return(shell.PWSH)

		props := properties.Map{
			properties.Style:     Agnoster,
			FolderSeparatorIcon:  " > ",
			FolderIcon:           "f",
			HomeIcon:             "~",
			Cycle:                tc.Cycle,
			CycleFolderSeparator: tc.ColorSeparator,
		}

		path := &Path{}
		path.Init(props, env)

		path.setPaths()
		path.setStyle()
		got := renderTemplateNoTrimSpace(env, "{{ .Path }}", path)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

type testAgnosterLeftPathCase struct {
	Case          string
	Expected      string
	Home          string
	PWD           string
	GOOS          string
	PathSeparator string
}

func TestAgnosterLeftPath(t *testing.T) {
	for _, tc := range testAgnosterLeftPathCases {
		env := new(mock.Environment)
		env.On("Home").Return(tc.Home)
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("Pwd").Return(tc.PWD)
		env.On("GOOS").Return(tc.GOOS)
		args := &runtime.Flags{
			PSWD: tc.PWD,
		}
		env.On("Flags").Return(args)
		env.On("Shell").Return(shell.PWSH)

		props := properties.Map{
			properties.Style:    AgnosterLeft,
			FolderSeparatorIcon: " > ",
			FolderIcon:          "f",
			HomeIcon:            "~",
		}

		path := &Path{}
		path.Init(props, env)

		path.setPaths()
		path.setStyle()
		got := renderTemplateNoTrimSpace(env, "{{ .Path }}", path)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestGetFolderSeparator(t *testing.T) {
	cases := []struct {
		Case                    string
		FolderSeparatorIcon     string
		FolderSeparatorTemplate string
		Expected                string
	}{
		{Case: "default", Expected: "/"},
		{Case: "icon - no template", FolderSeparatorIcon: "\ue5fe", Expected: "\ue5fe"},
		{Case: "template", FolderSeparatorTemplate: "{{ if eq .Shell \"bash\" }}\\{{ end }}", Expected: "\\"},
		{Case: "template empty", FolderSeparatorTemplate: "{{ if eq .Shell \"pwsh\" }}\\{{ end }}", Expected: "/"},
		{Case: "invalid template", FolderSeparatorTemplate: "{{ if eq .Shell \"pwsh\" }}", Expected: "/"},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Shell").Return(shell.GENERIC)

		template.Cache = &cache.Template{
			Shell: "bash",
		}
		template.Init(env, nil)

		props := properties.Map{}

		if len(tc.FolderSeparatorTemplate) > 0 {
			props[FolderSeparatorTemplate] = tc.FolderSeparatorTemplate
		}

		if len(tc.FolderSeparatorIcon) > 0 {
			props[FolderSeparatorIcon] = tc.FolderSeparatorIcon
		}

		path := &Path{
			pathSeparator: "/",
		}
		path.Init(props, env)

		got := path.getFolderSeparator()
		assert.Equal(t, tc.Expected, got)
	}
}

type testNormalizePathCase struct {
	Case          string
	Input         string
	HomeDir       string
	GOOS          string
	PathSeparator string
	Expected      string
}

func TestNormalizePath(t *testing.T) {
	for _, tc := range testNormalizePathCases {
		env := new(mock.Environment)
		env.On("Home").Return(tc.HomeDir)
		env.On("GOOS").Return(tc.GOOS)

		if len(tc.PathSeparator) == 0 {
			tc.PathSeparator = "/"
		}

		env.On("PathSeparator").Return(tc.PathSeparator)

		pt := &Path{}
		pt.Init(properties.Map{}, env)

		got := pt.normalize(tc.Input)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

type testSplitPathCase struct {
	Case         string
	GOOS         string
	Relative     string
	Root         string
	GitDir       *runtime.FileInfo
	GitDirFormat string
	Expected     Folders
}

func TestSplitPath(t *testing.T) {
	for _, tc := range testSplitPathCases {
		env := new(mock.Environment)
		env.On("PathSeparator").Return("/")
		env.On("Home").Return("/a/b")
		env.On("HasParentFilePath", ".git", false).Return(tc.GitDir, nil)
		env.On("GOOS").Return(tc.GOOS)

		props := properties.Map{
			GitDirFormat: tc.GitDirFormat,
		}

		path := &Path{
			root:          tc.Root,
			relative:      tc.Relative,
			pathSeparator: "/",
			windowsPath:   tc.GOOS == runtime.WINDOWS,
		}
		path.Init(props, env)

		got := path.splitPath()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestGetMaxWidth(t *testing.T) {
	cases := []struct {
		MaxWidth any
		Case     string
		Expected int
	}{
		{
			Case:     "Nil",
			Expected: 0,
		},
		{
			Case:     "Empty string",
			MaxWidth: "",
			Expected: 0,
		},
		{
			Case:     "Invalid template",
			MaxWidth: "{{ .Unknown }}",
			Expected: 0,
		},
		{
			Case:     "Environment variable",
			MaxWidth: "{{ .Env.MAX_WIDTH }}",
			Expected: 120,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Getenv", "MAX_WIDTH").Return("120")
		env.On("Shell").Return(shell.BASH)

		template.Cache = new(cache.Template)
		template.Init(env, nil)

		props := properties.Map{
			MaxWidth: tc.MaxWidth,
		}

		path := &Path{}
		path.Init(props, env)

		got := path.getMaxWidth()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
