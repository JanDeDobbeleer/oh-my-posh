package main

import (
	"testing"

	"github.com/distatus/battery"
	"github.com/gookit/config/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockedEnvironment struct {
	mock.Mock
}

func (env *MockedEnvironment) getenv(key string) string {
	args := env.Called(key)
	return args.String(0)
}

func (env *MockedEnvironment) pwd() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) homeDir() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) hasFiles(pattern string) bool {
	args := env.Called(pattern)
	return args.Bool(0)
}

func (env *MockedEnvironment) hasFilesInDir(dir, pattern string) bool {
	args := env.Called(dir, pattern)
	return args.Bool(0)
}

func (env *MockedEnvironment) hasFolder(folder string) bool {
	args := env.Called(folder)
	return args.Bool(0)
}

func (env *MockedEnvironment) getFileContent(file string) string {
	args := env.Called(file)
	return args.String(0)
}

func (env *MockedEnvironment) getFoldersList(path string) []string {
	args := env.Called(path)
	return args.Get(0).([]string)
}

func (env *MockedEnvironment) getPathSeperator() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) getCurrentUser() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) getHostName() (string, error) {
	args := env.Called()
	return args.String(0), args.Error(1)
}

func (env *MockedEnvironment) getRuntimeGOOS() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) getPlatform() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) hasCommand(command string) bool {
	args := env.Called(command)
	return args.Bool(0)
}

func (env *MockedEnvironment) runCommand(command string, args ...string) (string, error) {
	arguments := env.Called(command, args)
	return arguments.String(0), arguments.Error(1)
}

func (env *MockedEnvironment) runShellCommand(shell, command string) string {
	args := env.Called(shell, command)
	return args.String(0)
}

func (env *MockedEnvironment) lastErrorCode() int {
	args := env.Called()
	return args.Int(0)
}

func (env *MockedEnvironment) executionTime() float64 {
	args := env.Called()
	return float64(args.Int(0))
}

func (env *MockedEnvironment) isRunningAsRoot() bool {
	args := env.Called()
	return args.Bool(0)
}

func (env *MockedEnvironment) getArgs() *args {
	arguments := env.Called()
	return arguments.Get(0).(*args)
}

func (env *MockedEnvironment) getBatteryInfo() ([]*battery.Battery, error) {
	args := env.Called()
	return args.Get(0).([]*battery.Battery), args.Error(1)
}

func (env *MockedEnvironment) getShellName() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) getWindowTitle(imageName, windowTitleRegex string) (string, error) {
	args := env.Called(imageName)
	return args.String(0), args.Error(1)
}

func (env *MockedEnvironment) getWindowsRegistryKeyValue(path string) (*windowsRegistryValue, error) {
	args := env.Called(path)
	return args.Get(0).(*windowsRegistryValue), args.Error(1)
}

func (env *MockedEnvironment) HTTPRequest(url string, timeout int, requestModifiers ...HTTPRequestModifier) ([]byte, error) {
	args := env.Called(url)
	return args.Get(0).([]byte), args.Error(1)
}

func (env *MockedEnvironment) hasParentFilePath(path string) (*fileInfo, error) {
	args := env.Called(path)
	return args.Get(0).(*fileInfo), args.Error(1)
}

func (env *MockedEnvironment) stackCount() int {
	args := env.Called()
	return args.Int(0)
}

func (env *MockedEnvironment) isWsl() bool {
	args := env.Called()
	return args.Bool(0)
}

func (env *MockedEnvironment) isWsl2() bool {
	args := env.Called()
	return args.Bool(0)
}

func (env *MockedEnvironment) getTerminalWidth() (int, error) {
	args := env.Called()
	return args.Int(0), args.Error(1)
}

func (env *MockedEnvironment) getCachePath() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) cache() cache {
	args := env.Called()
	return args.Get(0).(cache)
}

func (env *MockedEnvironment) close() {
	_ = env.Called()
}

func (env *MockedEnvironment) logs() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) inWSLSharedDrive() bool {
	args := env.Called()
	return args.Bool(0)
}

func (env *MockedEnvironment) convertToWindowsPath(path string) string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) convertToLinuxPath(path string) string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) getWifiNetwork() (*wifiInfo, error) {
	args := env.Called()
	return args.Get(0).(*wifiInfo), args.Error(1)
}

func (env *MockedEnvironment) templateCache() *templateCache {
	args := env.Called()
	return args.Get(0).(*templateCache)
}

func (env *MockedEnvironment) onTemplate() {
	for _, call := range env.Mock.ExpectedCalls {
		if call.Method == "templateCache" {
			return
		}
	}
	env.On("templateCache").Return(&templateCache{
		Env: make(map[string]string),
	})
}

const (
	homeBill        = "/home/bill"
	homeJan         = "/usr/home/jan"
	homeBillWindows = "C:\\Users\\Bill"
	levelDir        = "/level"
)

func TestIsInHomeDirTrue(t *testing.T) {
	home := homeBill
	env := new(MockedEnvironment)
	env.On("homeDir").Return(home)
	path := &path{
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
	env := new(MockedEnvironment)
	env.On("homeDir").Return(home)
	path := &path{
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
		PathSeperator string
		HomeIcon      string
		RegistryIcon  string
	}{
		{Expected: "~", HomeIcon: "~", HomePath: "/home/bill/", Pwd: "/home/bill/", PathSeperator: "/"},
		{Expected: "usr", HomePath: "/home/bill/", Pwd: "/usr/error/what", PathSeperator: "/"},
		{Expected: "C:", HomePath: "C:\\Users\\Bill", Pwd: "C:\\Program Files\\Go", PathSeperator: "\\"},
		{Expected: "REG", RegistryIcon: "REG", HomePath: "C:\\Users\\Bill", Pwd: "HKCU:\\Program Files\\Go", PathSeperator: "\\"},
		{Expected: "~", HomeIcon: "~", HomePath: "C:\\Users\\Bill", Pwd: "Microsoft.PowerShell.Core\\FileSystem::C:\\Users\\Bill", PathSeperator: "\\"},
		{Expected: "C:", HomePath: "C:\\Users\\Jack", Pwd: "Microsoft.PowerShell.Core\\FileSystem::C:\\Users\\Bill", PathSeperator: "\\"},
		{Expected: "", HomePath: "C:\\Users\\Jack", Pwd: "", PathSeperator: "\\"},
		{Expected: "DRIVE:", HomePath: "/home/bill/", Pwd: "/usr/error/what", Pswd: "DRIVE:", PathSeperator: "/"},
	}
	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("homeDir").Return(tc.HomePath)
		env.On("pwd").Return(tc.Pwd)
		args := &args{
			PSWD: &tc.Pswd,
		}
		env.On("getArgs").Return(args)
		env.On("getPathSeperator").Return(tc.PathSeperator)
		env.On("getRuntimeGOOS").Return("")
		path := &path{
			env: env,
			props: properties{
				HomeIcon:            tc.HomeIcon,
				WindowsRegistryIcon: tc.RegistryIcon,
			},
		}
		got := path.rootLocation()
		assert.EqualValues(t, tc.Expected, got)
	}
}

func TestIsInHomeDirFalse(t *testing.T) {
	home := homeBill
	env := new(MockedEnvironment)
	env.On("homeDir").Return(home)
	path := &path{
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
	env := new(MockedEnvironment)
	env.On("getPathSeperator").Return("/")
	env.On("getRunteGOOS").Return("")
	path := &path{
		env: env,
	}
	got := path.pathDepth(pwd)
	assert.Equal(t, 99, got)
}

func TestPathDepthZeroLevelsDeep(t *testing.T) {
	pwd := "/usr/"
	env := new(MockedEnvironment)
	env.On("getPathSeperator").Return("/")
	path := &path{
		env: env,
	}
	got := path.pathDepth(pwd)
	assert.Equal(t, 0, got)
}

func TestPathDepthOneLevelDeep(t *testing.T) {
	pwd := "/usr/location"
	env := new(MockedEnvironment)
	env.On("getPathSeperator").Return("/")
	path := &path{
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
		PathSeperator       string
		HomeIcon            string
		FolderSeparatorIcon string
		Style               string
		GOOS                string
		MaxDepth            int
	}{
		{Style: AgnosterFull, Expected: "usr > location > whatever", HomePath: "/usr/home", Pwd: "/usr/location/whatever", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "usr > .. > man", HomePath: "/usr/home", Pwd: "/usr/location/whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "~ > .. > man", HomePath: "/usr/home", Pwd: "/usr/home/whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "~ > projects", HomePath: "/usr/home", Pwd: "/usr/home/projects", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "C:", HomePath: homeBillWindows, Pwd: "C:", PathSeperator: "\\", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "/", HomePath: homeBillWindows, Pwd: "/", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "foo", HomePath: homeBillWindows, Pwd: "/foo", PathSeperator: "/", FolderSeparatorIcon: " > "},

		{Style: AgnosterShort, Expected: "usr > .. > bar > man", HomePath: "/usr/home", Pwd: "/usr/foo/bar/man", PathSeperator: "/", FolderSeparatorIcon: " > ", MaxDepth: 2},
		{Style: AgnosterShort, Expected: "usr > foo > bar > man", HomePath: "/usr/home", Pwd: "/usr/foo/bar/man", PathSeperator: "/", FolderSeparatorIcon: " > ", MaxDepth: 3},
		{Style: AgnosterShort, Expected: "~ > .. > bar > man", HomePath: "/usr/home", Pwd: "/usr/home/foo/bar/man", PathSeperator: "/", FolderSeparatorIcon: " > ", MaxDepth: 2},
		{Style: AgnosterShort, Expected: "~ > foo > bar > man", HomePath: "/usr/home", Pwd: "/usr/home/foo/bar/man", PathSeperator: "/", FolderSeparatorIcon: " > ", MaxDepth: 3},

		{
			Style:               AgnosterShort,
			Expected:            "C: > .. > bar > man",
			HomePath:            homeBillWindows,
			Pwd:                 "C:\\usr\\foo\\bar\\man",
			PathSeperator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            2,
		},
		{
			Style:               AgnosterShort,
			Expected:            "C: > .. > foo > bar > man",
			HomePath:            homeBillWindows,
			Pwd:                 "C:\\usr\\foo\\bar\\man",
			PathSeperator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            3,
		},
		{
			Style:               AgnosterShort,
			Expected:            "~ > .. > bar > man",
			HomePath:            homeBillWindows,
			Pwd:                 "C:\\Users\\Bill\\foo\\bar\\man",
			PathSeperator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            2,
		},
		{
			Style:               AgnosterShort,
			Expected:            "~ > foo > bar > man",
			HomePath:            homeBillWindows,
			Pwd:                 "C:\\Users\\Bill\\foo\\bar\\man",
			PathSeperator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            3,
		},

		{Style: AgnosterFull, Expected: "PSDRIVE: | src", HomePath: homeBillWindows, Pwd: "/foo", Pswd: "PSDRIVE:/src", PathSeperator: "/", FolderSeparatorIcon: " | "},
		{Style: AgnosterShort, Expected: "PSDRIVE: | .. | init", HomePath: homeBillWindows, Pwd: "/foo", Pswd: "PSDRIVE:/src/init", PathSeperator: "/", FolderSeparatorIcon: " | "},

		{Style: Mixed, Expected: "~ > .. > man", HomePath: "/usr/home", Pwd: "/usr/home/whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: Mixed, Expected: "~ > ab > .. > man", HomePath: "/usr/home", Pwd: "/usr/home/ab/whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},

		{Style: Letter, Expected: "~ > a > w > man", HomePath: "/usr/home", Pwd: "/usr/home/ab/whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "u > b > a > w > man", HomePath: "/usr/home", Pwd: "/usr/burp/ab/whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "u > .b > a > w > man", HomePath: "/usr/home", Pwd: "/usr/.burp/ab/whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "u > .b > a > .w > man", HomePath: "/usr/home", Pwd: "/usr/.burp/ab/.whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "u > .b > a > ._w > man", HomePath: "/usr/home", Pwd: "/usr/.burp/ab/._whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "u > .ä > ū > .w > man", HomePath: "/usr/home", Pwd: "/usr/.äufbau/ūmgebung/.whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "u > .b > 1 > .w > man", HomePath: "/usr/home", Pwd: "/usr/.burp/12345/.whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "u > .b > 1 > .w > man", HomePath: "/usr/home", Pwd: "/usr/.burp/12345abc/.whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "u > .b > __p > .w > man", HomePath: "/usr/home", Pwd: "/usr/.burp/__pycache__/.whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "➼ > .w > man", HomePath: "/usr/home", Pwd: "➼/.whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: Letter, Expected: "➼ s > .w > man", HomePath: "/usr/home", Pwd: "➼ something/.whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},
	}
	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getPathSeperator").Return(tc.PathSeperator)
		env.On("homeDir").Return(tc.HomePath)
		env.On("pwd").Return(tc.Pwd)
		env.On("getRuntimeGOOS").Return(tc.GOOS)
		env.On("stackCount").Return(0)
		args := &args{
			PSWD: &tc.Pswd,
		}
		env.On("getArgs").Return(args)
		env.onTemplate()
		path := &path{
			env: env,
			props: properties{
				FolderSeparatorIcon: tc.FolderSeparatorIcon,
				Style:               tc.Style,
				MaxDepth:            tc.MaxDepth,
				SegmentTemplate:     "{{ .Path }}",
			},
		}
		got := path.string()
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

		{Style: Folder, FolderSeparatorIcon: "\\", Pwd: "C:\\", Expected: "C:\\", PathSeparator: "\\", GOOS: windowsPlatform},
		{Style: Full, FolderSeparatorIcon: "\\", Pwd: "C:\\Users\\Jan", Expected: "C:\\Users\\Jan", PathSeparator: "\\", GOOS: windowsPlatform},

		// StackCountEnabled=true and StackCount=2
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", StackCount: 2, Expected: "2 /"},
		{Style: Full, Pwd: "", StackCount: 2, Expected: "2 "},
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
		env := new(MockedEnvironment)
		if len(tc.PathSeparator) == 0 {
			tc.PathSeparator = "/"
		}
		env.On("getPathSeperator").Return(tc.PathSeparator)
		env.On("homeDir").Return("/usr/home")
		env.On("pwd").Return(tc.Pwd)
		env.On("getRuntimeGOOS").Return(tc.GOOS)
		env.On("stackCount").Return(tc.StackCount)
		args := &args{
			PSWD: &tc.Pswd,
		}
		env.On("getArgs").Return(args)
		env.onTemplate()
		if len(tc.Template) == 0 {
			tc.Template = "{{ if gt .StackCount 0 }}{{ .StackCount }} {{ end }}{{ .Path }}"
		}
		props := properties{
			Style:           tc.Style,
			SegmentTemplate: tc.Template,
		}
		if tc.FolderSeparatorIcon != "" {
			props[FolderSeparatorIcon] = tc.FolderSeparatorIcon
		}
		if tc.DisableMappedLocations {
			props[MappedLocationsEnabled] = false
		}
		path := &path{
			env:   env,
			props: props,
		}
		got := path.string()
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
		env := new(MockedEnvironment)
		env.On("getPathSeperator").Return("/")
		env.On("homeDir").Return("/usr/home")
		env.On("pwd").Return(tc.Pwd)
		env.On("getRuntimeGOOS").Return("")
		args := &args{
			PSWD: &tc.Pwd,
		}
		env.On("getArgs").Return(args)
		path := &path{
			env: env,
			props: properties{
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
		{Input: "C:\\Users\\Bob\\Foo", GOOS: linuxPlatform, Expected: "C:/Users/Bob/Foo"},
		{Input: "C:\\Users\\Bob\\Foo", GOOS: windowsPlatform, Expected: "c:/users/bob/foo"},
		{Input: "~\\Bob\\Foo", GOOS: linuxPlatform, Expected: "/usr/home/Bob/Foo"},
		{Input: "~\\Bob\\Foo", GOOS: windowsPlatform, Expected: "/usr/home/bob/foo"},
		{Input: "/foo/~/bar", GOOS: linuxPlatform, Expected: "/foo/~/bar"},
		{Input: "/foo/~/bar", GOOS: windowsPlatform, Expected: "/foo/~/bar"},
		{Input: "~/baz", GOOS: linuxPlatform, Expected: "/usr/home/baz"},
		{Input: "~/baz", GOOS: windowsPlatform, Expected: "/usr/home/baz"},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("homeDir").Return("/usr/home")
		env.On("getRuntimeGOOS").Return(tc.GOOS)
		pt := &path{
			env: env,
		}
		got := pt.normalize(tc.Input)
		assert.Equal(t, tc.Expected, got)
	}
}

func TestGetFolderPathCustomMappedLocations(t *testing.T) {
	pwd := "/a/b/c/d"
	env := new(MockedEnvironment)
	env.On("getPathSeperator").Return("/")
	env.On("homeDir").Return("/usr/home")
	env.On("pwd").Return(pwd)
	env.On("getRuntimeGOOS").Return("")
	args := &args{
		PSWD: &pwd,
	}
	env.On("getArgs").Return(args)
	path := &path{
		env: env,
		props: properties{
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
		env := new(MockedEnvironment)
		env.On("homeDir").Return(tc.Home)
		env.On("getPathSeperator").Return(tc.PathSeparator)
		env.On("pwd").Return(tc.PWD)
		env.On("getRuntimeGOOS").Return("")
		args := &args{
			PSWD: &tc.PWD,
		}
		env.On("getArgs").Return(args)
		path := &path{
			env: env,
			props: properties{
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
		env := new(MockedEnvironment)
		env.On("homeDir").Return(tc.Home)
		env.On("getPathSeperator").Return(tc.PathSeparator)
		env.On("pwd").Return(tc.PWD)
		env.On("getRuntimeGOOS").Return("")
		args := &args{
			PSWD: &tc.PWD,
		}
		env.On("getArgs").Return(args)
		path := &path{
			env: env,
			props: properties{
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
		env := new(MockedEnvironment)
		env.On("getPathSeperator").Return("/")
		env.On("homeDir").Return("/usr/home")
		env.On("pwd").Return(tc.Pwd)
		env.On("getRuntimeGOOS").Return("")
		args := &args{
			PSWD: &tc.Pswd,
		}
		env.On("getArgs").Return(args)
		path := &path{
			env: env,
			props: properties{
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

func TestParseMappedLocations(t *testing.T) {
	defer testClearDefaultConfig()
	cases := []struct {
		Case string
		JSON string
	}{
		{Case: "new format", JSON: `{ "properties": { "mapped_locations": {"folder1": "one","folder2": "two"} } }`},
		{Case: "old format", JSON: `{ "properties": { "mapped_locations": [["folder1", "one"], ["folder2", "two"]] } }`},
	}
	for _, tc := range cases {
		config.ClearAll()
		config.WithOptions(func(opt *config.Options) {
			opt.DecoderConfig = &mapstructure.DecoderConfig{
				TagName: "config",
			}
		})
		err := config.LoadStrings(config.JSON, tc.JSON)
		assert.NoError(t, err)
		var segment Segment
		err = config.BindStruct("", &segment)
		assert.NoError(t, err)
		mappedLocations := segment.Properties.getKeyValueMap(MappedLocations, make(map[string]string))
		assert.Equal(t, "two", mappedLocations["folder2"])
	}
}
