package segments

import (
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"oh-my-posh/template"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

func renderTemplate(env *mock.MockedEnvironment, segmentTemplate string, context interface{}) string {
	found := false
	for _, call := range env.Mock.ExpectedCalls {
		if call.Method == "TemplateCache" {
			found = true
			break
		}
	}
	if !found {
		env.On("TemplateCache").Return(&environment.TemplateCache{
			Env: make(map[string]string),
		})
	}
	env.On("Log", mock2.Anything, mock2.Anything, mock2.Anything)
	tmpl := &template.Text{
		Template: segmentTemplate,
		Context:  context,
		Env:      env,
	}
	text, err := tmpl.Render()
	if err != nil {
		return err.Error()
	}
	return strings.TrimSpace(text)
}

const (
	homeBill        = "/home/bill"
	homeJan         = "/usr/home/jan"
	homeBillWindows = "C:\\Users\\Bill"
	levelDir        = "/level"
)

func TestIsInHomeDirTrue(t *testing.T) {
	home := homeBill
	env := new(mock.MockedEnvironment)
	env.On("Home").Return(home)
	path := &Path{
		env: env,
	}
	got := path.inHomeDir(home)
	assert.True(t, got)
}

func TestIsInHomeDirLevelTrue(t *testing.T) {
	home := homeBill
	pwd := home
	for i := 0; i < 99; i++ {
		pwd += levelDir
	}
	env := new(mock.MockedEnvironment)
	env.On("Home").Return(home)
	path := &Path{
		env: env,
	}
	got := path.inHomeDir(pwd)
	assert.True(t, got)
}

func TestRootLocationHome(t *testing.T) {
	cases := []struct {
		Expected      string
		HomePath      string
		Pswd          string
		Pwd           string
		PathSeparator string
		HomeIcon      string
		RegistryIcon  string
	}{
		{Expected: "~", HomeIcon: "~", HomePath: "/home/bill/", Pwd: "/home/bill/", PathSeparator: "/"},
		{Expected: "usr", HomePath: "/home/bill/", Pwd: "/usr/error/what", PathSeparator: "/"},
		{Expected: "C:", HomePath: "C:\\Users\\Bill", Pwd: "C:\\Program Files\\Go", PathSeparator: "\\"},
		{Expected: "REG", RegistryIcon: "REG", HomePath: "C:\\Users\\Bill", Pwd: "HKCU:\\Program Files\\Go", PathSeparator: "\\"},
		{Expected: "~", HomeIcon: "~", HomePath: "C:\\Users\\Bill", Pwd: "Microsoft.PowerShell.Core\\FileSystem::C:\\Users\\Bill", PathSeparator: "\\"},
		{Expected: "C:", HomePath: "C:\\Users\\Jack", Pwd: "Microsoft.PowerShell.Core\\FileSystem::C:\\Users\\Bill", PathSeparator: "\\"},
		{Expected: "", HomePath: "C:\\Users\\Jack", Pwd: "", PathSeparator: "\\"},
		{Expected: "DRIVE:", HomePath: "/home/bill/", Pwd: "/usr/error/what", Pswd: "DRIVE:", PathSeparator: "/"},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Home").Return(tc.HomePath)
		env.On("Pwd").Return(tc.Pwd)
		args := &environment.Flags{
			PSWD: tc.Pswd,
		}
		env.On("Flags").Return(args)
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("GOOS").Return("")
		path := &Path{
			env: env,
			props: properties.Map{
				HomeIcon:            tc.HomeIcon,
				WindowsRegistryIcon: tc.RegistryIcon,
			},
		}
		got := path.rootLocation()
		assert.EqualValues(t, tc.Expected, got)
	}
}

func TestParent(t *testing.T) {
	// there's no Windows support/validation for this just yet
	// mainly due to root being a special case
	if runtime.GOOS == environment.WINDOWS {
		return
	}
	cases := []struct {
		Case          string
		Expected      string
		HomePath      string
		Pwd           string
		PathSeparator string
	}{
		{Case: "Home folder", Expected: "", HomePath: "/home/bill", Pwd: "/home/bill", PathSeparator: "/"},
		{Case: "Inside home folder", Expected: "~/", HomePath: "/home/bill", Pwd: "/home/bill/test", PathSeparator: "/"},
		{Case: "Root", Expected: "", HomePath: "/home/bill", Pwd: "/", PathSeparator: "/"},
		{Case: "Root + 1", Expected: "/", HomePath: "/home/bill", Pwd: "/usr", PathSeparator: "/"},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Home").Return(tc.HomePath)
		env.On("Pwd").Return(tc.Pwd)
		env.On("Flags").Return(&environment.Flags{})
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("GOOS").Return(environment.DARWIN)
		path := &Path{
			env:   env,
			props: properties.Map{},
		}
		path.pwd = tc.Pwd
		got := path.Parent()
		assert.EqualValues(t, tc.Expected, got, tc.Case)
	}
}

func TestIsInHomeDirFalse(t *testing.T) {
	home := homeBill
	env := new(mock.MockedEnvironment)
	env.On("Home").Return(home)
	path := &Path{
		env: env,
	}
	got := path.inHomeDir("/usr/error")
	assert.False(t, got)
}

func TestPathDepthMultipleLevelsDeep(t *testing.T) {
	pwd := "/usr"
	for i := 0; i < 99; i++ {
		pwd += levelDir
	}
	env := new(mock.MockedEnvironment)
	env.On("PathSeparator").Return("/")
	env.On("getRunteGOOS").Return("")
	path := &Path{
		env: env,
	}
	got := path.pathDepth(pwd)
	assert.Equal(t, 99, got)
}

func TestPathDepthZeroLevelsDeep(t *testing.T) {
	pwd := "/usr/"
	env := new(mock.MockedEnvironment)
	env.On("PathSeparator").Return("/")
	path := &Path{
		env: env,
	}
	got := path.pathDepth(pwd)
	assert.Equal(t, 0, got)
}

func TestPathDepthOneLevelDeep(t *testing.T) {
	pwd := "/usr/location"
	env := new(mock.MockedEnvironment)
	env.On("PathSeparator").Return("/")
	path := &Path{
		env: env,
	}
	got := path.pathDepth(pwd)
	assert.Equal(t, 1, got)
}

func TestAgnosterPathStyles(t *testing.T) {
	cases := []struct {
		Expected            string
		HomePath            string
		Pswd                string
		Pwd                 string
		PathSeparator       string
		HomeIcon            string
		FolderSeparatorIcon string
		Style               string
		GOOS                string
		MaxDepth            int
		HideRootLocation    bool
	}{
		{Style: AgnosterFull, Expected: "usr > location > whatever", HomePath: "/usr/home", Pwd: "/usr/location/whatever", PathSeparator: "/", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "usr > .. > man", HomePath: "/usr/home", Pwd: "/usr/location/whatever/man", PathSeparator: "/", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "~ > .. > man", HomePath: "/usr/home", Pwd: "/usr/home/whatever/man", PathSeparator: "/", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "~ > projects", HomePath: "/usr/home", Pwd: "/usr/home/projects", PathSeparator: "/", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "C:", HomePath: homeBillWindows, Pwd: "C:", PathSeparator: "\\", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "/", HomePath: homeBillWindows, Pwd: "/", PathSeparator: "/", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "foo", HomePath: homeBillWindows, Pwd: "/foo", PathSeparator: "/", FolderSeparatorIcon: " > "},

		{Style: AgnosterShort, Expected: "usr > .. > bar > man", HomePath: "/usr/home", Pwd: "/usr/foo/bar/man", PathSeparator: "/", FolderSeparatorIcon: " > ", MaxDepth: 2},
		{Style: AgnosterShort, Expected: "usr > foo > bar > man", HomePath: "/usr/home", Pwd: "/usr/foo/bar/man", PathSeparator: "/", FolderSeparatorIcon: " > ", MaxDepth: 3},
		{Style: AgnosterShort, Expected: "~ > .. > bar > man", HomePath: "/usr/home", Pwd: "/usr/home/foo/bar/man", PathSeparator: "/", FolderSeparatorIcon: " > ", MaxDepth: 2},
		{Style: AgnosterShort, Expected: "~ > foo > bar > man", HomePath: "/usr/home", Pwd: "/usr/home/foo/bar/man", PathSeparator: "/", FolderSeparatorIcon: " > ", MaxDepth: 3},

		{
			Style:               AgnosterShort,
			Expected:            "C: > .. > bar > man",
			HomePath:            homeBillWindows,
			Pwd:                 "C:\\usr\\foo\\bar\\man",
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            2,
		},
		{
			Style:               AgnosterShort,
			Expected:            "C: > .. > foo > bar > man",
			HomePath:            homeBillWindows,
			Pwd:                 "C:\\usr\\foo\\bar\\man",
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            3,
		},
		{
			Style:               AgnosterShort,
			Expected:            "~ > .. > bar > man",
			HomePath:            homeBillWindows,
			Pwd:                 "C:\\Users\\Bill\\foo\\bar\\man",
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            2,
		},
		{
			Style:               AgnosterShort,
			Expected:            "~ > foo > bar > man",
			HomePath:            homeBillWindows,
			Pwd:                 "C:\\Users\\Bill\\foo\\bar\\man",
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            3,
		},

		{Style: AgnosterFull, Expected: "PSDRIVE: | src", HomePath: homeBillWindows, Pwd: "/foo", Pswd: "PSDRIVE:/src", PathSeparator: "/", FolderSeparatorIcon: " | "},
		{Style: AgnosterShort, Expected: "PSDRIVE: | .. | init", HomePath: homeBillWindows, Pwd: "/foo", Pswd: "PSDRIVE:/src/init", PathSeparator: "/", FolderSeparatorIcon: " | "},

		{Style: AgnosterShort, Expected: "src | init", HomePath: homeBillWindows, Pwd: "/foo", Pswd: "PSDRIVE:/src/init", PathSeparator: "/", FolderSeparatorIcon: " | ", MaxDepth: 2,
			HideRootLocation: true},
		{Style: AgnosterShort, Expected: "PSDRIVE: | src", HomePath: homeBillWindows, Pwd: "/foo", Pswd: "PSDRIVE:/src", PathSeparator: "/", FolderSeparatorIcon: " | ", MaxDepth: 2,
			HideRootLocation: true},
		{Style: AgnosterShort, Expected: "~", HomePath: homeBillWindows, Pwd: homeBillWindows, PathSeparator: "\\", FolderSeparatorIcon: " > ", MaxDepth: 1, HideRootLocation: true},
		{Style: AgnosterShort, Expected: "foo", HomePath: homeBillWindows, Pwd: homeBillWindows + "\\foo", PathSeparator: "\\", FolderSeparatorIcon: "\\", MaxDepth: 1,
			HideRootLocation: true},
		{Style: AgnosterShort, Expected: "~\\foo", HomePath: homeBillWindows, Pwd: homeBillWindows + "\\foo", PathSeparator: "\\", FolderSeparatorIcon: "\\", MaxDepth: 2,
			HideRootLocation: true},
		{Style: AgnosterShort, Expected: "~", HomePath: "/usr/home", Pwd: "/usr/home", PathSeparator: "/", FolderSeparatorIcon: " > ", MaxDepth: 1, HideRootLocation: true},
		{Style: AgnosterShort, Expected: "foo", HomePath: "/usr/home", Pwd: "/usr/home/foo", PathSeparator: "/", FolderSeparatorIcon: "/", MaxDepth: 1, HideRootLocation: true},
		{Style: AgnosterShort, Expected: "bar > man", HomePath: "/usr/home", Pwd: "/usr/foo/bar/man", PathSeparator: "/", FolderSeparatorIcon: " > ", MaxDepth: 2,
			HideRootLocation: true},
		{Style: AgnosterShort, Expected: "foo > bar > man", HomePath: "/usr/home", Pwd: "/usr/foo/bar/man", PathSeparator: "/", FolderSeparatorIcon: " > ", MaxDepth: 3,
			HideRootLocation: true},
		{Style: AgnosterShort, Expected: "~ > foo", HomePath: "/usr/home", Pwd: "/usr/home/foo", PathSeparator: "/", FolderSeparatorIcon: " > ", MaxDepth: 2, HideRootLocation: true},
		{Style: AgnosterShort, Expected: "~ > foo > bar > man", HomePath: "/usr/home", Pwd: "/usr/home/foo/bar/man", PathSeparator: "/", FolderSeparatorIcon: " > ", MaxDepth: 4,
			HideRootLocation: true},
		{Style: AgnosterShort, Expected: "C:", HomePath: "/usr/home", Pwd: "/mnt/c", Pswd: "C:", PathSeparator: "/", FolderSeparatorIcon: " | ", MaxDepth: 2, HideRootLocation: true},
		{Style: AgnosterShort, Expected: "~ | space foo", HomePath: "/usr/home", Pwd: "/usr/home/space foo", PathSeparator: "/", FolderSeparatorIcon: " | ", MaxDepth: 2,
			HideRootLocation: true},
		{Style: AgnosterShort, Expected: "space foo", HomePath: "/usr/home", Pwd: "/usr/home/space foo", PathSeparator: "/", FolderSeparatorIcon: " | ", MaxDepth: 1,
			HideRootLocation: true},

		{Style: Mixed, Expected: "~ > .. > man", HomePath: "/usr/home", Pwd: "/usr/home/whatever/man", PathSeparator: "/", FolderSeparatorIcon: " > "},
		{Style: Mixed, Expected: "~ > ab > .. > man", HomePath: "/usr/home", Pwd: "/usr/home/ab/whatever/man", PathSeparator: "/", FolderSeparatorIcon: " > "},

		{Style: Letter, Expected: "~ > a > w > man", HomePath: "/usr/home", Pwd: "/usr/home/ab/whatever/man", PathSeparator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "u > b > a > w > man", HomePath: "/usr/home", Pwd: "/usr/burp/ab/whatever/man", PathSeparator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "u > .b > a > w > man", HomePath: "/usr/home", Pwd: "/usr/.burp/ab/whatever/man", PathSeparator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "u > .b > a > .w > man", HomePath: "/usr/home", Pwd: "/usr/.burp/ab/.whatever/man", PathSeparator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "u > .b > a > ._w > man", HomePath: "/usr/home", Pwd: "/usr/.burp/ab/._whatever/man", PathSeparator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "u > .ä > ū > .w > man", HomePath: "/usr/home", Pwd: "/usr/.äufbau/ūmgebung/.whatever/man", PathSeparator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "u > .b > 1 > .w > man", HomePath: "/usr/home", Pwd: "/usr/.burp/12345/.whatever/man", PathSeparator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "u > .b > 1 > .w > man", HomePath: "/usr/home", Pwd: "/usr/.burp/12345abc/.whatever/man", PathSeparator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "u > .b > __p > .w > man", HomePath: "/usr/home", Pwd: "/usr/.burp/__pycache__/.whatever/man", PathSeparator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "➼ > .w > man", HomePath: "/usr/home", Pwd: "➼/.whatever/man", PathSeparator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "➼ s > .w > man", HomePath: "/usr/home", Pwd: "➼ something/.whatever/man", PathSeparator: "/", FolderSeparatorIcon: " > "},

		{Style: Unique, Expected: "~ > a > ab > abcd", HomePath: "/usr/home", Pwd: "/usr/home/ab/abc/abcd", PathSeparator: "/", FolderSeparatorIcon: " > "},
		{Style: Unique, Expected: "~ > a > .a > abcd", HomePath: "/usr/home", Pwd: "/usr/home/ab/.abc/abcd", PathSeparator: "/", FolderSeparatorIcon: " > "},
		{Style: Unique, Expected: "~ > a > ab > abcd", HomePath: "/usr/home", Pwd: "/usr/home/ab/ab/abcd", PathSeparator: "/", FolderSeparatorIcon: " > "},

		{Style: AgnosterShort, Expected: "localhost > c$", HomePath: homeBillWindows, Pwd: "\\\\localhost\\c$", PathSeparator: "\\", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "localhost\\c$", HomePath: homeBillWindows, Pwd: "\\\\localhost\\c$", PathSeparator: "\\", FolderSeparatorIcon: "\\"},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("Home").Return(tc.HomePath)
		env.On("Pwd").Return(tc.Pwd)
		env.On("GOOS").Return(tc.GOOS)
		env.On("StackCount").Return(0)
		env.On("IsWsl").Return(false)
		env.On("DirIsWritable", tc.Pwd).Return(true)
		args := &environment.Flags{
			PSWD: tc.Pswd,
		}
		env.On("Flags").Return(args)
		path := &Path{
			env: env,
			props: properties.Map{
				FolderSeparatorIcon: tc.FolderSeparatorIcon,
				properties.Style:    tc.Style,
				MaxDepth:            tc.MaxDepth,
				HideRootLocation:    tc.HideRootLocation,
			},
		}
		_ = path.Enabled()
		got := renderTemplate(env, "{{ .Path }}", path)
		assert.Equal(t, tc.Expected, got)
	}
}

func TestGetFullPath(t *testing.T) {
	cases := []struct {
		Style                  string
		FolderSeparatorIcon    string
		Pwd                    string
		Pswd                   string
		Expected               string
		DisableMappedLocations bool
		GOOS                   string
		PathSeparator          string
		StackCount             int
		Template               string
	}{
		{Style: Full, Pwd: "/usr/home/abc", Template: "{{ .Path }}", StackCount: 2, Expected: "~/abc"},

		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", Expected: "/"},
		{Style: Full, Pwd: "", Expected: ""},
		{Style: Full, Pwd: "/", Expected: "/"},
		{Style: Full, Pwd: "/usr/home", Expected: "~"},
		{Style: Full, Pwd: "/usr/home/abc", Expected: "~/abc"},
		{Style: Full, Pwd: "/usr/home/abc", Expected: "/usr/home/abc", DisableMappedLocations: true},
		{Style: Full, Pwd: "/a/b/c/d", Expected: "/a/b/c/d"},

		{Style: Full, FolderSeparatorIcon: "|", Pwd: "", Expected: ""},
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/usr/home", Expected: "~"},
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/usr/home", Expected: "|usr|home", DisableMappedLocations: true},
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/usr/home/abc", Expected: "~|abc"},
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/a/b/c/d", Expected: "|a|b|c|d"},

		{Style: Folder, Pwd: "", Expected: ""},
		{Style: Folder, Pwd: "/", Expected: "/"},
		{Style: Folder, Pwd: "/usr/home", Expected: "~"},
		{Style: Folder, Pwd: "/usr/home", Expected: "home", DisableMappedLocations: true},
		{Style: Folder, Pwd: "/usr/home/abc", Expected: "abc"},
		{Style: Folder, Pwd: "/a/b/c/d", Expected: "d"},

		{Style: Folder, FolderSeparatorIcon: "|", Pwd: "", Expected: ""},
		{Style: Folder, FolderSeparatorIcon: "|", Pwd: "/", Expected: "/"},
		{Style: Folder, FolderSeparatorIcon: "|", Pwd: "/usr/home", Expected: "~"},
		{Style: Folder, FolderSeparatorIcon: "|", Pwd: "/usr/home", Expected: "home", DisableMappedLocations: true},
		{Style: Folder, FolderSeparatorIcon: "|", Pwd: "/usr/home/abc", Expected: "abc"},
		{Style: Folder, FolderSeparatorIcon: "|", Pwd: "/a/b/c/d", Expected: "d"},

		{Style: Folder, FolderSeparatorIcon: "\\", Pwd: "C:\\", Expected: "C:\\", PathSeparator: "\\", GOOS: environment.WINDOWS},
		{Style: Full, FolderSeparatorIcon: "\\", Pwd: "C:\\Users\\Jan", Expected: "C:\\Users\\Jan", PathSeparator: "\\", GOOS: environment.WINDOWS},

		// StackCountEnabled=true and StackCount=2
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", StackCount: 2, Expected: "2 /"},
		{Style: Full, Pwd: "", StackCount: 2, Expected: "2"},
		{Style: Full, Pwd: "/", StackCount: 2, Expected: "2 /"},
		{Style: Full, Pwd: "/usr/home", StackCount: 2, Expected: "2 ~"},
		{Style: Full, Pwd: "/usr/home/abc", StackCount: 2, Expected: "2 ~/abc"},
		{Style: Full, Pwd: "/usr/home/abc", StackCount: 2, Expected: "2 /usr/home/abc", DisableMappedLocations: true},
		{Style: Full, Pwd: "/a/b/c/d", StackCount: 2, Expected: "2 /a/b/c/d"},

		// StackCountEnabled=false and StackCount=2
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", Template: "{{ .Path }}", StackCount: 2, Expected: "/"},
		{Style: Full, Pwd: "", Template: "{{ .Path }}", StackCount: 2, Expected: ""},
		{Style: Full, Pwd: "/", Template: "{{ .Path }}", StackCount: 2, Expected: "/"},
		{Style: Full, Pwd: "/usr/home", Template: "{{ .Path }}", StackCount: 2, Expected: "~"},

		{Style: Full, Pwd: "/usr/home/abc", Template: "{{ .Path }}", StackCount: 2, Expected: "/usr/home/abc", DisableMappedLocations: true},
		{Style: Full, Pwd: "/a/b/c/d", Template: "{{ .Path }}", StackCount: 2, Expected: "/a/b/c/d"},

		// StackCountEnabled=true and StackCount=0
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", StackCount: 0, Expected: "/"},
		{Style: Full, Pwd: "", StackCount: 0, Expected: ""},
		{Style: Full, Pwd: "/", StackCount: 0, Expected: "/"},
		{Style: Full, Pwd: "/usr/home", StackCount: 0, Expected: "~"},
		{Style: Full, Pwd: "/usr/home/abc", StackCount: 0, Expected: "~/abc"},
		{Style: Full, Pwd: "/usr/home/abc", StackCount: 0, Expected: "/usr/home/abc", DisableMappedLocations: true},
		{Style: Full, Pwd: "/a/b/c/d", StackCount: 0, Expected: "/a/b/c/d"},

		// StackCountEnabled=true and StackCount<0
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", StackCount: -1, Expected: "/"},
		{Style: Full, Pwd: "", StackCount: -1, Expected: ""},
		{Style: Full, Pwd: "/", StackCount: -1, Expected: "/"},
		{Style: Full, Pwd: "/usr/home", StackCount: -1, Expected: "~"},
		{Style: Full, Pwd: "/usr/home/abc", StackCount: -1, Expected: "~/abc"},
		{Style: Full, Pwd: "/usr/home/abc", StackCount: -1, Expected: "/usr/home/abc", DisableMappedLocations: true},
		{Style: Full, Pwd: "/a/b/c/d", StackCount: -1, Expected: "/a/b/c/d"},

		// StackCountEnabled=true and StackCount not set
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", Expected: "/"},
		{Style: Full, Pwd: "", Expected: ""},
		{Style: Full, Pwd: "/", Expected: "/"},
		{Style: Full, Pwd: "/usr/home", Expected: "~"},
		{Style: Full, Pwd: "/usr/home/abc", Expected: "~/abc"},
		{Style: Full, Pwd: "/usr/home/abc", Expected: "/usr/home/abc", DisableMappedLocations: true},
		{Style: Full, Pwd: "/a/b/c/d", Expected: "/a/b/c/d"},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		if len(tc.PathSeparator) == 0 {
			tc.PathSeparator = "/"
		}
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("Home").Return("/usr/home")
		env.On("Pwd").Return(tc.Pwd)
		env.On("GOOS").Return(tc.GOOS)
		env.On("StackCount").Return(tc.StackCount)
		env.On("IsWsl").Return(false)
		env.On("DirIsWritable", tc.Pwd).Return(true)
		args := &environment.Flags{
			PSWD: tc.Pswd,
		}
		env.On("Flags").Return(args)
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
			env:   env,
			props: props,
		}
		_ = path.Enabled()
		got := renderTemplate(env, tc.Template, path)
		assert.Equal(t, tc.Expected, got)
	}
}

func TestGetFullPathCustomMappedLocations(t *testing.T) {
	cases := []struct {
		Pwd             string
		MappedLocations map[string]string
		Expected        string
	}{
		{Pwd: "/a/b/c/d", MappedLocations: map[string]string{"/a/b/c/d": "#"}, Expected: "#"},
		{Pwd: "/a/b/c/d", MappedLocations: map[string]string{"\\a\\b": "#"}, Expected: "#/c/d"},
		{Pwd: "\\a\\b\\c\\d", MappedLocations: map[string]string{"\\a\\b": "#"}, Expected: "#\\c\\d"},
		{Pwd: "/a/b/c/d", MappedLocations: map[string]string{"/a/b": "#"}, Expected: "#/c/d"},
		{Pwd: "/a/b/c/d", MappedLocations: map[string]string{"/a/b": "/e/f"}, Expected: "/e/f/c/d"},
		{Pwd: "/usr/home/a/b/c/d", MappedLocations: map[string]string{"~\\a\\b": "#"}, Expected: "#/c/d"},
		{Pwd: "/usr/home/a/b/c/d", MappedLocations: map[string]string{"~/a/b": "#"}, Expected: "#/c/d"},
		{Pwd: "/a/usr/home/b/c/d", MappedLocations: map[string]string{"/a~": "#"}, Expected: "/a/usr/home/b/c/d"},
		{Pwd: "/usr/home/a/b/c/d", MappedLocations: map[string]string{"/a/b": "#"}, Expected: "/usr/home/a/b/c/d"},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("PathSeparator").Return("/")
		env.On("Home").Return("/usr/home")
		env.On("Pwd").Return(tc.Pwd)
		env.On("GOOS").Return("")
		args := &environment.Flags{
			PSWD: tc.Pwd,
		}
		env.On("Flags").Return(args)
		path := &Path{
			env: env,
			props: properties.Map{
				MappedLocationsEnabled: false,
				MappedLocations:        tc.MappedLocations,
			},
		}
		got := path.getFullPath()
		assert.Equal(t, tc.Expected, got)
	}
}

func TestNormalizePath(t *testing.T) {
	cases := []struct {
		Input    string
		GOOS     string
		Expected string
	}{
		{Input: "C:\\Users\\Bob\\Foo", GOOS: environment.LINUX, Expected: "C:/Users/Bob/Foo"},
		{Input: "C:\\Users\\Bob\\Foo", GOOS: environment.WINDOWS, Expected: "c:/users/bob/foo"},
		{Input: "~\\Bob\\Foo", GOOS: environment.LINUX, Expected: "/usr/home/Bob/Foo"},
		{Input: "~\\Bob\\Foo", GOOS: environment.WINDOWS, Expected: "/usr/home/bob/foo"},
		{Input: "/foo/~/bar", GOOS: environment.LINUX, Expected: "/foo/~/bar"},
		{Input: "/foo/~/bar", GOOS: environment.WINDOWS, Expected: "/foo/~/bar"},
		{Input: "~/baz", GOOS: environment.LINUX, Expected: "/usr/home/baz"},
		{Input: "~/baz", GOOS: environment.WINDOWS, Expected: "/usr/home/baz"},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Home").Return("/usr/home")
		env.On("GOOS").Return(tc.GOOS)
		pt := &Path{
			env: env,
		}
		got := pt.normalize(tc.Input)
		assert.Equal(t, tc.Expected, got)
	}
}

func TestGetFolderPathCustomMappedLocations(t *testing.T) {
	pwd := "/a/b/c/d"
	env := new(mock.MockedEnvironment)
	env.On("PathSeparator").Return("/")
	env.On("Home").Return("/usr/home")
	env.On("Pwd").Return(pwd)
	env.On("GOOS").Return("")
	args := &environment.Flags{
		PSWD: pwd,
	}
	env.On("Flags").Return(args)
	path := &Path{
		env: env,
		props: properties.Map{
			MappedLocations: map[string]string{
				"/a/b/c/d": "#",
			},
		},
	}
	got := path.getFolderPath()
	assert.Equal(t, "#", got)
}

func TestAgnosterPath(t *testing.T) { // nolint:dupl
	cases := []struct {
		Case          string
		Expected      string
		Home          string
		PWD           string
		PathSeparator string
	}{
		{Case: "Windows outside home", Expected: "C: > f > f > location", Home: homeBillWindows, PWD: "C:\\Program Files\\Go\\location", PathSeparator: "\\"},
		{Case: "Windows oustide home", Expected: "~ > f > f > location", Home: homeBillWindows, PWD: homeBillWindows + "\\Documents\\Bill\\location", PathSeparator: "\\"},
		{Case: "Windows inside home zero levels", Expected: "C: > location", Home: homeBillWindows, PWD: "C:\\location", PathSeparator: "\\"},
		{Case: "Windows inside home one level", Expected: "C: > f > location", Home: homeBillWindows, PWD: "C:\\Program Files\\location", PathSeparator: "\\"},
		{Case: "Windows lower case drive letter", Expected: "C: > Windows", Home: homeBillWindows, PWD: "C:\\Windows\\", PathSeparator: "\\"},
		{Case: "Windows lower case drive letter (other)", Expected: "P: > Other", Home: homeBillWindows, PWD: "P:\\Other\\", PathSeparator: "\\"},
		{Case: "Windows lower word drive", Expected: "some: > some", Home: homeBillWindows, PWD: "some:\\some\\", PathSeparator: "\\"},
		{Case: "Windows lower word drive (ending with c)", Expected: "src: > source", Home: homeBillWindows, PWD: "src:\\source\\", PathSeparator: "\\"},
		{Case: "Windows lower word drive (arbitrary cases)", Expected: "sRc: > source", Home: homeBillWindows, PWD: "sRc:\\source\\", PathSeparator: "\\"},
		{Case: "Windows registry drive", Expected: "\uf013 > f > magnetic:test", Home: homeBillWindows, PWD: "HKLM:\\SOFTWARE\\magnetic:test\\", PathSeparator: "\\"},
		{Case: "Windows registry drive case sensitive", Expected: "\uf013 > f > magnetic:TOAST", Home: homeBillWindows, PWD: "HKLM:\\SOFTWARE\\magnetic:TOAST\\", PathSeparator: "\\"},
		{Case: "Unix outside home", Expected: "mnt > f > f > location", Home: homeJan, PWD: "/mnt/go/test/location", PathSeparator: "/"},
		{Case: "Unix inside home", Expected: "~ > f > f > location", Home: homeJan, PWD: homeJan + "/docs/jan/location", PathSeparator: "/"},
		{Case: "Unix outside home zero levels", Expected: "mnt > location", Home: homeJan, PWD: "/mnt/location", PathSeparator: "/"},
		{Case: "Unix outside home one level", Expected: "mnt > f > location", Home: homeJan, PWD: "/mnt/folder/location", PathSeparator: "/"},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Home").Return(tc.Home)
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("Pwd").Return(tc.PWD)
		env.On("GOOS").Return("")
		args := &environment.Flags{
			PSWD: tc.PWD,
		}
		env.On("Flags").Return(args)
		path := &Path{
			env: env,
			props: properties.Map{
				FolderSeparatorIcon: " > ",
				FolderIcon:          "f",
				HomeIcon:            "~",
			},
		}
		got := path.getAgnosterPath()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestAgnosterLeftPath(t *testing.T) { // nolint:dupl
	cases := []struct {
		Case          string
		Expected      string
		Home          string
		PWD           string
		PathSeparator string
	}{
		{Case: "Windows outside home", Expected: "C: > Program Files > f > f", Home: homeBillWindows, PWD: "C:\\Program Files\\Go\\location", PathSeparator: "\\"},
		{Case: "Windows inside home", Expected: "~ > Documents > f > f", Home: homeBillWindows, PWD: homeBillWindows + "\\Documents\\Bill\\location", PathSeparator: "\\"},
		{Case: "Windows inside home zero levels", Expected: "C: > location", Home: homeBillWindows, PWD: "C:\\location", PathSeparator: "\\"},
		{Case: "Windows inside home one level", Expected: "C: > Program Files > f", Home: homeBillWindows, PWD: "C:\\Program Files\\location", PathSeparator: "\\"},
		{Case: "Windows lower case drive letter", Expected: "C: > Windows", Home: homeBillWindows, PWD: "C:\\Windows\\", PathSeparator: "\\"},
		{Case: "Windows lower case drive letter (other)", Expected: "P: > Other", Home: homeBillWindows, PWD: "P:\\Other\\", PathSeparator: "\\"},
		{Case: "Windows lower word drive", Expected: "some: > some", Home: homeBillWindows, PWD: "some:\\some\\", PathSeparator: "\\"},
		{Case: "Windows lower word drive (ending with c)", Expected: "src: > source", Home: homeBillWindows, PWD: "src:\\source\\", PathSeparator: "\\"},
		{Case: "Windows lower word drive (arbitrary cases)", Expected: "sRc: > source", Home: homeBillWindows, PWD: "sRc:\\source\\", PathSeparator: "\\"},
		{Case: "Windows registry drive", Expected: "\uf013 > SOFTWARE > f", Home: homeBillWindows, PWD: "HKLM:\\SOFTWARE\\magnetic:test\\", PathSeparator: "\\"},
		{Case: "Windows registry drive case sensitive", Expected: "\uf013 > SOFTWARE > f", Home: homeBillWindows, PWD: "HKLM:\\SOFTWARE\\magnetic:TOAST\\", PathSeparator: "\\"},
		{Case: "Unix outside home", Expected: "mnt > go > f > f", Home: homeJan, PWD: "/mnt/go/test/location", PathSeparator: "/"},
		{Case: "Unix inside home", Expected: "~ > docs > f > f", Home: homeJan, PWD: homeJan + "/docs/jan/location", PathSeparator: "/"},
		{Case: "Unix outside home zero levels", Expected: "mnt > location", Home: homeJan, PWD: "/mnt/location", PathSeparator: "/"},
		{Case: "Unix outside home one level", Expected: "mnt > folder > f", Home: homeJan, PWD: "/mnt/folder/location", PathSeparator: "/"},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Home").Return(tc.Home)
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("Pwd").Return(tc.PWD)
		env.On("GOOS").Return("")
		args := &environment.Flags{
			PSWD: tc.PWD,
		}
		env.On("Flags").Return(args)
		path := &Path{
			env: env,
			props: properties.Map{
				FolderSeparatorIcon: " > ",
				FolderIcon:          "f",
				HomeIcon:            "~",
			},
		}
		got := path.getAgnosterLeftPath()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestGetPwd(t *testing.T) {
	cases := []struct {
		MappedLocationsEnabled bool
		Pwd                    string
		Pswd                   string
		Expected               string
	}{
		{MappedLocationsEnabled: true, Pwd: "", Expected: ""},
		{MappedLocationsEnabled: true, Pwd: "/usr", Expected: "/usr"},
		{MappedLocationsEnabled: true, Pwd: "/usr/home", Expected: "~"},
		{MappedLocationsEnabled: true, Pwd: "/usr/home/abc", Expected: "~/abc"},
		{MappedLocationsEnabled: true, Pwd: "/a/b/c/d", Expected: "#"},
		{MappedLocationsEnabled: true, Pwd: "/a/b/c/d/e/f/g", Expected: "#/e/f/g"},
		{MappedLocationsEnabled: true, Pwd: "/z/y/x/w", Expected: "/z/y/x/w"},

		{MappedLocationsEnabled: false, Pwd: "", Expected: ""},
		{MappedLocationsEnabled: false, Pwd: "/usr/home/abc", Expected: "/usr/home/abc"},
		{MappedLocationsEnabled: false, Pwd: "/a/b/c/d/e/f/g", Expected: "#/e/f/g"},
		{MappedLocationsEnabled: false, Pwd: "/usr/home/c/d/e/f/g", Expected: "/usr/home/c/d/e/f/g"},
		{MappedLocationsEnabled: true, Pwd: "/usr/home/c/d/e/f/g", Expected: "~/c/d/e/f/g"},

		{MappedLocationsEnabled: true, Pwd: "/w/d/x/w", Pswd: "/z/y/x/w", Expected: "/z/y/x/w"},
		{MappedLocationsEnabled: false, Pwd: "/f/g/k/d/e/f/g", Pswd: "/a/b/c/d/e/f/g", Expected: "#/e/f/g"},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("PathSeparator").Return("/")
		env.On("Home").Return("/usr/home")
		env.On("Pwd").Return(tc.Pwd)
		env.On("GOOS").Return("")
		args := &environment.Flags{
			PSWD: tc.Pswd,
		}
		env.On("Flags").Return(args)
		path := &Path{
			env: env,
			props: properties.Map{
				MappedLocationsEnabled: tc.MappedLocationsEnabled,
				MappedLocations: map[string]string{
					"/a/b/c/d": "#",
				},
			},
		}
		got := path.getPwd()
		assert.Equal(t, tc.Expected, got)
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
		{Case: "template empty", FolderSeparatorTemplate: "{{ if eq .Shell \"pwsh\" }}\\{{ end }}", Expected: ""},
		{Case: "invalid template", FolderSeparatorTemplate: "{{ if eq .Shell \"pwsh\" }}", Expected: ""},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("PathSeparator").Return("/")
		env.On("Log", mock2.Anything, mock2.Anything, mock2.Anything)
		path := &Path{
			env: env,
		}
		props := properties.Map{}
		if len(tc.FolderSeparatorTemplate) > 0 {
			props[FolderSeparatorTemplate] = tc.FolderSeparatorTemplate
		}
		if len(tc.FolderSeparatorIcon) > 0 {
			props[FolderSeparatorIcon] = tc.FolderSeparatorIcon
		}
		env.On("TemplateCache").Return(&environment.TemplateCache{
			Env:   make(map[string]string),
			Shell: "bash",
		})
		path.props = props
		got := path.getFolderSeparator()
		assert.Equal(t, tc.Expected, got)
	}
}
