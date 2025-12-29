package segments

import (
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

const (
	homeDir        = "/home/someone"
	homeDirWindows = "C:\\Users\\someone"
)

func renderTemplateNoTrimSpace(env *mock.Environment, segmentTemplate string, context any) string {
	env.On("Shell").Return("foo")

	if template.Cache == nil {
		template.Cache = &cache.Template{}
	}
	template.Init(env, nil, nil)

	text, err := template.Render(segmentTemplate, context)
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

		props := options.Map{
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

		if tc.Shell == "" {
			tc.Shell = shell.PWSH
		}
		env.On("Shell").Return(tc.Shell)

		displayCygpath := tc.Cygwin
		if displayCygpath {
			env.On("RunCommand", "cygpath", []string{"-u", tc.Pwd}).Return(tc.Cygpath, tc.CygpathError)
			env.On("RunCommand", "cygpath", testify_.Anything).Return("brrrr", nil)
		}

		props := options.Map{
			FolderSeparatorIcon: tc.FolderSeparatorIcon,
			options.Style:       tc.Style,
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
		if tc.PathSeparator == "" {
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
		if tc.Template == "" {
			tc.Template = "{{ if gt .StackCount 0 }}{{ .StackCount }} {{ end }}{{ .Path }}"
		}
		props := options.Map{
			options.Style: tc.Style,
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

		if tc.GOOS == "" {
			tc.GOOS = runtime.DARWIN
		}

		env.On("GOOS").Return(tc.GOOS)

		if tc.PathSeparator == "" {
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
		template.Init(env, nil, nil)

		props := options.Map{
			options.Style:          Full,
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

		props := options.Map{
			options.Style:        Agnoster,
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

		props := options.Map{
			options.Style:       AgnosterLeft,
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
			SimpleTemplate: cache.SimpleTemplate{
				Shell: "bash",
			},
		}
		template.Init(env, nil, nil)

		props := options.Map{}

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
	Cygwin        bool
}

func TestNormalizePath(t *testing.T) {
	for _, tc := range testNormalizePathCases {
		env := new(mock.Environment)
		env.On("Home").Return(tc.HomeDir)
		env.On("GOOS").Return(tc.GOOS)

		if tc.PathSeparator == "" {
			tc.PathSeparator = "/"
		}

		env.On("PathSeparator").Return(tc.PathSeparator)

		pt := &Path{cygPath: tc.Cygwin}
		pt.Init(options.Map{}, env)

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

		props := options.Map{
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
		template.Init(env, nil, nil)

		props := options.Map{
			MaxWidth: tc.MaxWidth,
		}

		path := &Path{}
		path.Init(props, env)

		got := path.getMaxWidth()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestAgnosterMaxWidth(t *testing.T) {
	cases := []struct {
		name        string
		pwd         string
		folderIcon  string
		separator   string
		expected    string
		goos        string
		maxWidth    int
		displayRoot bool
	}{
		{
			name:        "path shorter than maxWidth",
			pwd:         "/foob/user/docs",
			maxWidth:    20,
			displayRoot: false,
			separator:   "/",
			folderIcon:  `..`,
			expected:    "foob/user/docs",
			goos:        runtime.LINUX,
		},
		{
			name:        "path shorter than maxWidth, Windows",
			pwd:         `C:\Users\john\Documents`,
			maxWidth:    20,
			displayRoot: true,
			folderIcon:  `..`,
			separator:   `\`,
			expected:    `..\..\john\Documents`,
			goos:        runtime.WINDOWS,
		},
		{
			name:        "path shorter than maxWidth, wth root",
			pwd:         "/foob/user/docs",
			maxWidth:    20,
			displayRoot: true,
			folderIcon:  `..`,
			separator:   "/",
			expected:    "/foob/user/docs",
			goos:        runtime.LINUX,
		},
		{
			name:        "path exactly maxWidth",
			pwd:         "/foob/user/docs",
			maxWidth:    15,
			displayRoot: true,
			folderIcon:  `..`,
			separator:   "/",
			expected:    "/foob/user/docs",
			goos:        runtime.LINUX,
		},
		{
			name:        "path longer than maxWidth with folder icons",
			pwd:         "/foob/user/documents/projects",
			maxWidth:    15,
			displayRoot: false,
			folderIcon:  "..",
			separator:   "/",
			expected:    "../../projects",
			goos:        runtime.LINUX,
		},
		{
			name:        "very long path requiring multiple folder replacements",
			pwd:         "/foob/user/documents/projects/myproject/src/main",
			maxWidth:    21,
			displayRoot: false,
			folderIcon:  "..",
			separator:   "/",
			expected:    "../../../../../main",
			goos:        runtime.LINUX,
		},
		{
			name:        "path requiring final folder truncation",
			pwd:         "/foob/verylongfoldername",
			maxWidth:    15,
			displayRoot: false,
			separator:   "/",
			expected:    "verylongfolder…",
			goos:        runtime.LINUX,
		},
		{
			name:        "Windows path with custom separator",
			pwd:         `C:\Users\john\Documents`,
			maxWidth:    15,
			displayRoot: false,
			folderIcon:  "…",
			separator:   `\`,
			expected:    `…\…\…\Documents`,
			goos:        runtime.WINDOWS,
		},
		{
			name:        "single folder path",
			pwd:         "/foob",
			maxWidth:    10,
			displayRoot: false,
			separator:   "/",
			expected:    "foob",
			goos:        runtime.LINUX,
		},
		{
			name:        "empty relative path",
			pwd:         "/",
			maxWidth:    10,
			displayRoot: true,
			separator:   "/",
			expected:    "/",
			goos:        runtime.LINUX,
		},
		{
			name:        "custom folder icon",
			pwd:         "/foob/user/documents/projects",
			maxWidth:    15,
			displayRoot: false,
			folderIcon:  "⋯",
			separator:   "/",
			expected:    "⋯/⋯/⋯/projects",
			goos:        runtime.LINUX,
		},
		{
			name:        "maxwidth is smaller than folder name",
			pwd:         "/foob/user/documents/projects",
			maxWidth:    2,
			displayRoot: false,
			folderIcon:  "⋯",
			separator:   "/",
			expected:    "p…",
			goos:        runtime.LINUX,
		},
		{
			name:        "maxwidth is 0",
			pwd:         "/foob/user/documents/projects",
			maxWidth:    0,
			displayRoot: false,
			folderIcon:  "⋯",
			separator:   "/",
			expected:    "…",
			goos:        runtime.LINUX,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := &mock.Environment{}
			env.On("Pwd").Return(tc.pwd)
			env.On("Home").Return("/home")
			env.On("GOOS").Return(tc.goos)
			env.On("Shell").Return(shell.BASH)

			path := &Path{
				Base: Base{
					env: env,
					options: options.Map{
						DisplayRoot:         tc.displayRoot,
						FolderIcon:          tc.folderIcon,
						FolderSeparatorIcon: tc.separator,
					},
				},
				pathSeparator: tc.separator,
			}

			// Set up the path state
			path.setPaths()

			got := path.getAgnosterMaxWidth(tc.maxWidth)
			assert.Equal(t, tc.expected, got, tc.name)
		})
	}
}

func TestFishPath(t *testing.T) {
	cases := []struct {
		name           string
		pwd            string
		separator      string
		goos           string
		expected       string
		dirLength      int
		fullLengthDirs int
	}{
		{
			name:           "default settings",
			pwd:            "/home/user/documents/projects",
			dirLength:      1,
			fullLengthDirs: 1,
			expected:       "h/u/d/projects",
			separator:      "/",
		},
		{
			name:           "dir length 2",
			pwd:            "/home/user/documents/projects",
			dirLength:      2,
			fullLengthDirs: 1,
			expected:       "ho/us/do/projects",
			separator:      "/",
		},
		{
			name:           "full length dirs 2",
			pwd:            "/home/user/documents/projects/myproject",
			dirLength:      1,
			fullLengthDirs: 2,
			expected:       "h/u/d/projects/myproject",
			separator:      "/",
		},
		{
			name:           "dir length 3, full length dirs 2",
			pwd:            "/home/user/documents/projects/myproject",
			dirLength:      3,
			fullLengthDirs: 2,
			expected:       "hom/use/doc/projects/myproject",
			separator:      "/",
		},
		{
			name:           "full length dirs 2 - Windows",
			pwd:            `C:\Users\Jan\Documents\Projects\Myproject`,
			dirLength:      1,
			fullLengthDirs: 2,
			expected:       `C\U\J\D\Projects\Myproject`,
			separator:      `\`,
		},
		{
			name:           "dir length 3, full length dirs 2 - Windows",
			pwd:            `C:\Users\Jan\Documents\Projects\Myproject`,
			dirLength:      3,
			fullLengthDirs: 2,
			expected:       `C:\Use\Jan\Doc\Projects\Myproject`,
			separator:      `\`,
		},
		{
			name:           "single folder",
			pwd:            "/home",
			dirLength:      1,
			fullLengthDirs: 1,
			expected:       "home",
			separator:      "/",
		},
		{
			name:           "two folders with full length dirs 1",
			pwd:            "/home/user",
			dirLength:      1,
			fullLengthDirs: 1,
			expected:       "h/user",
			separator:      "/",
		},
		{
			name:           "root only",
			pwd:            "/",
			dirLength:      1,
			fullLengthDirs: 1,
			expected:       "/",
			separator:      "/",
		},
		{
			name:           "dir length 0 should disable shortening",
			pwd:            "/home/user/documents",
			dirLength:      0,
			fullLengthDirs: 1,
			expected:       "home/user/documents",
			separator:      "/",
		},
		{
			name:           "dir length negative should disable shortening",
			pwd:            "/home/user/documents",
			dirLength:      -1,
			fullLengthDirs: 1,
			expected:       "home/user/documents",
			separator:      "/",
		},
		{
			name:           "full length dirs 0 should fallback to 1",
			pwd:            "/home/user/documents",
			dirLength:      1,
			fullLengthDirs: 0,
			expected:       "h/u/documents",
			separator:      "/",
		},
		{
			name:           "full length dirs negative should fallback to 1",
			pwd:            "/home/user/documents",
			dirLength:      1,
			fullLengthDirs: -1,
			expected:       "h/u/documents",
			separator:      "/",
		},
		{
			name:           "full length dirs greater than total folders",
			pwd:            "/home/user",
			dirLength:      1,
			fullLengthDirs: 5,
			expected:       "home/user",
			separator:      "/",
		},
		{
			name:           "dir length greater than folder name",
			pwd:            "/a/b/c",
			dirLength:      10,
			fullLengthDirs: 1,
			expected:       "a/b/c",
			separator:      "/",
		},
		{
			name:           "multi-byte unicode home icon",
			pwd:            "/󰋜/Downloads/test",
			dirLength:      1,
			fullLengthDirs: 1,
			expected:       "󰋜/D/test",
			separator:      "/",
		},
		{
			name:           "multi-byte unicode home icon with dir length 2",
			pwd:            "/󰋜/Documents/Projects",
			dirLength:      2,
			fullLengthDirs: 1,
			expected:       "󰋜/Do/Projects",
			separator:      "/",
		},
		{
			name:           "path with emoji folders",
			pwd:            "/🏠/📁/💻",
			dirLength:      1,
			fullLengthDirs: 1,
			expected:       "🏠/📁/💻",
			separator:      "/",
		},
		{
			name:           "mixed multi-byte and ascii",
			pwd:            "/󰋜test/normal/󰨳end",
			dirLength:      2,
			fullLengthDirs: 1,
			expected:       "󰋜t/no/󰨳end",
			separator:      "/",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := &mock.Environment{}
			env.On("Pwd").Return(tc.pwd)
			env.On("Home").Return("/foob")
			env.On("GOOS").Return(tc.goos)
			env.On("Shell").Return(shell.BASH)

			path := &Path{
				Base: Base{
					env: env,
					options: options.Map{
						DirLength:      tc.dirLength,
						FullLengthDirs: tc.fullLengthDirs,
					},
				},
				pathSeparator: tc.separator,
			}

			path.setPaths()
			result := path.getFishPath()

			assert.Equal(t, result, tc.expected, tc.name)
		})
	}
}
