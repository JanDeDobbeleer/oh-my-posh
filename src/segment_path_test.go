package main

import (
	"testing"

	"github.com/distatus/battery"
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

func (env *MockedEnvironment) getcwd() string {
	args := env.Called(nil)
	return args.String(0)
}

func (env *MockedEnvironment) homeDir() string {
	args := env.Called(nil)
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

func (env *MockedEnvironment) getPathSeperator() string {
	args := env.Called(nil)
	return args.String(0)
}

func (env *MockedEnvironment) getCurrentUser() string {
	args := env.Called(nil)
	return args.String(0)
}

func (env *MockedEnvironment) getHostName() (string, error) {
	args := env.Called(nil)
	return args.String(0), args.Error(1)
}

func (env *MockedEnvironment) getRuntimeGOOS() string {
	args := env.Called(nil)
	return args.String(0)
}

func (env *MockedEnvironment) getPlatform() string {
	args := env.Called(nil)
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
	args := env.Called(nil)
	return args.Int(0)
}

func (env *MockedEnvironment) executionTime() float64 {
	args := env.Called(nil)
	return float64(args.Int(0))
}

func (env *MockedEnvironment) isRunningAsRoot() bool {
	args := env.Called(nil)
	return args.Bool(0)
}

func (env *MockedEnvironment) getArgs() *args {
	arguments := env.Called(nil)
	return arguments.Get(0).(*args)
}

func (env *MockedEnvironment) getBatteryInfo() (*battery.Battery, error) {
	args := env.Called(nil)
	return args.Get(0).(*battery.Battery), args.Error(1)
}

func (env *MockedEnvironment) getShellName() string {
	args := env.Called(nil)
	return args.String(0)
}

func (env *MockedEnvironment) getWindowTitle(imageName, windowTitleRegex string) (string, error) {
	args := env.Called(imageName)
	return args.String(0), args.Error(1)
}

func (env *MockedEnvironment) doGet(url string) ([]byte, error) {
	args := env.Called(url)
	return args.Get(0).([]byte), args.Error(1)
}

func (env *MockedEnvironment) hasParentFilePath(path string) (*fileInfo, error) {
	args := env.Called(path)
	return args.Get(0).(*fileInfo), args.Error(1)
}

func (env *MockedEnvironment) stackCount() int {
	args := env.Called(nil)
	return args.Int(0)
}

func (env *MockedEnvironment) isWsl() bool {
	return false
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
	env.On("homeDir", nil).Return(home)
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
	env.On("homeDir", nil).Return(home)
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
		props := &properties{
			values: map[Property]interface{}{
				HomeIcon:            tc.HomeIcon,
				WindowsRegistryIcon: tc.RegistryIcon,
			},
		}
		env := new(MockedEnvironment)
		env.On("homeDir", nil).Return(tc.HomePath)
		env.On("getcwd", nil).Return(tc.Pwd)
		args := &args{
			PSWD: &tc.Pswd,
		}
		env.On("getArgs", nil).Return(args)
		env.On("getPathSeperator", nil).Return(tc.PathSeperator)
		path := &path{
			env:   env,
			props: props,
		}
		got := path.rootLocation()
		assert.EqualValues(t, tc.Expected, got)
	}
}

func TestIsInHomeDirFalse(t *testing.T) {
	home := homeBill
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return(home)
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
	env.On("getPathSeperator", nil).Return("/")
	path := &path{
		env: env,
	}
	got := path.pathDepth(pwd)
	assert.Equal(t, 99, got)
}

func TestPathDepthZeroLevelsDeep(t *testing.T) {
	pwd := "/usr/"
	env := new(MockedEnvironment)
	env.On("getPathSeperator", nil).Return("/")
	path := &path{
		env: env,
	}
	got := path.pathDepth(pwd)
	assert.Equal(t, 0, got)
}

func TestPathDepthOneLevelDeep(t *testing.T) {
	pwd := "/usr/location"
	env := new(MockedEnvironment)
	env.On("getPathSeperator", nil).Return("/")
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
	}{
		{Style: AgnosterFull, Expected: "usr > location > whatever", HomePath: "/usr/home", Pwd: "/usr/location/whatever", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "usr > .. > man", HomePath: "/usr/home", Pwd: "/usr/location/whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "~ > .. > man", HomePath: "/usr/home", Pwd: "/usr/home/whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "~ > projects", HomePath: "/usr/home", Pwd: "/usr/home/projects", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "C:", HomePath: homeBillWindows, Pwd: "C:", PathSeperator: "\\", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "", HomePath: homeBillWindows, Pwd: "/", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: AgnosterShort, Expected: "foo", HomePath: homeBillWindows, Pwd: "/foo", PathSeperator: "/", FolderSeparatorIcon: " > "},

		{Style: AgnosterFull, Expected: "PSDRIVE: | src", HomePath: homeBillWindows, Pwd: "/foo", Pswd: "PSDRIVE:/src", PathSeperator: "/", FolderSeparatorIcon: " | "},
		{Style: AgnosterShort, Expected: "PSDRIVE: | .. | init", HomePath: homeBillWindows, Pwd: "/foo", Pswd: "PSDRIVE:/src/init", PathSeperator: "/", FolderSeparatorIcon: " | "},

		{Style: Mixed, Expected: "~ > .. > man", HomePath: "/usr/home", Pwd: "/usr/home/whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},
		{Style: Mixed, Expected: "~ > ab > .. > man", HomePath: "/usr/home", Pwd: "/usr/home/ab/whatever/man", PathSeperator: "/", FolderSeparatorIcon: " > "},
	}
	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getPathSeperator", nil).Return(tc.PathSeperator)
		env.On("homeDir", nil).Return(tc.HomePath)
		env.On("getcwd", nil).Return(tc.Pwd)
		env.On("getRuntimeGOOS", nil).Return(tc.GOOS)
		args := &args{
			PSWD: &tc.Pswd,
		}
		env.On("getArgs", nil).Return(args)
		path := &path{
			env: env,
			props: &properties{
				values: map[Property]interface{}{
					FolderSeparatorIcon: tc.FolderSeparatorIcon,
					Style:               tc.Style,
				},
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
		StackCountEnabled      bool
	}{
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
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", StackCountEnabled: true, StackCount: 2, Expected: "2 /"},
		{Style: Full, Pwd: "", StackCountEnabled: true, StackCount: 2, Expected: "2 "},
		{Style: Full, Pwd: "/", StackCountEnabled: true, StackCount: 2, Expected: "2 /"},
		{Style: Full, Pwd: "/usr/home", StackCountEnabled: true, StackCount: 2, Expected: "2 ~"},
		{Style: Full, Pwd: "/usr/home/abc", StackCountEnabled: true, StackCount: 2, Expected: "2 ~/abc"},
		{Style: Full, Pwd: "/usr/home/abc", StackCountEnabled: true, StackCount: 2, Expected: "2 /usr/home/abc", DisableMappedLocations: true},
		{Style: Full, Pwd: "/a/b/c/d", StackCountEnabled: true, StackCount: 2, Expected: "2 /a/b/c/d"},

		// StackCountEnabled=false and StackCount=2
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", StackCountEnabled: false, StackCount: 2, Expected: "/"},
		{Style: Full, Pwd: "", StackCountEnabled: false, StackCount: 2, Expected: ""},
		{Style: Full, Pwd: "/", StackCountEnabled: false, StackCount: 2, Expected: "/"},
		{Style: Full, Pwd: "/usr/home", StackCountEnabled: false, StackCount: 2, Expected: "~"},
		{Style: Full, Pwd: "/usr/home/abc", StackCountEnabled: false, StackCount: 2, Expected: "~/abc"},
		{Style: Full, Pwd: "/usr/home/abc", StackCountEnabled: false, StackCount: 2, Expected: "/usr/home/abc", DisableMappedLocations: true},
		{Style: Full, Pwd: "/a/b/c/d", StackCountEnabled: false, StackCount: 2, Expected: "/a/b/c/d"},

		// StackCountEnabled=true and StackCount=0
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", StackCountEnabled: true, StackCount: 0, Expected: "/"},
		{Style: Full, Pwd: "", StackCountEnabled: true, StackCount: 0, Expected: ""},
		{Style: Full, Pwd: "/", StackCountEnabled: true, StackCount: 0, Expected: "/"},
		{Style: Full, Pwd: "/usr/home", StackCountEnabled: true, StackCount: 0, Expected: "~"},
		{Style: Full, Pwd: "/usr/home/abc", StackCountEnabled: true, StackCount: 0, Expected: "~/abc"},
		{Style: Full, Pwd: "/usr/home/abc", StackCountEnabled: true, StackCount: 0, Expected: "/usr/home/abc", DisableMappedLocations: true},
		{Style: Full, Pwd: "/a/b/c/d", StackCountEnabled: true, StackCount: 0, Expected: "/a/b/c/d"},

		// StackCountEnabled=true and StackCount<0
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", StackCountEnabled: true, StackCount: -1, Expected: "/"},
		{Style: Full, Pwd: "", StackCountEnabled: true, StackCount: -1, Expected: ""},
		{Style: Full, Pwd: "/", StackCountEnabled: true, StackCount: -1, Expected: "/"},
		{Style: Full, Pwd: "/usr/home", StackCountEnabled: true, StackCount: -1, Expected: "~"},
		{Style: Full, Pwd: "/usr/home/abc", StackCountEnabled: true, StackCount: -1, Expected: "~/abc"},
		{Style: Full, Pwd: "/usr/home/abc", StackCountEnabled: true, StackCount: -1, Expected: "/usr/home/abc", DisableMappedLocations: true},
		{Style: Full, Pwd: "/a/b/c/d", StackCountEnabled: true, StackCount: -1, Expected: "/a/b/c/d"},

		// StackCountEnabled=true and StackCount not set
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", StackCountEnabled: true, Expected: "/"},
		{Style: Full, Pwd: "", StackCountEnabled: true, Expected: ""},
		{Style: Full, Pwd: "/", StackCountEnabled: true, Expected: "/"},
		{Style: Full, Pwd: "/usr/home", StackCountEnabled: true, Expected: "~"},
		{Style: Full, Pwd: "/usr/home/abc", StackCountEnabled: true, Expected: "~/abc"},
		{Style: Full, Pwd: "/usr/home/abc", StackCountEnabled: true, Expected: "/usr/home/abc", DisableMappedLocations: true},
		{Style: Full, Pwd: "/a/b/c/d", StackCountEnabled: true, Expected: "/a/b/c/d"},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		if len(tc.PathSeparator) == 0 {
			tc.PathSeparator = "/"
		}
		env.On("getPathSeperator", nil).Return(tc.PathSeparator)
		env.On("homeDir", nil).Return("/usr/home")
		env.On("getcwd", nil).Return(tc.Pwd)
		env.On("getRuntimeGOOS", nil).Return(tc.GOOS)
		env.On("stackCount", nil).Return(tc.StackCount)
		args := &args{
			PSWD: &tc.Pswd,
		}
		env.On("getArgs", nil).Return(args)
		props := &properties{
			values: map[Property]interface{}{
				Style:             tc.Style,
				StackCountEnabled: tc.StackCountEnabled,
			},
		}
		if tc.FolderSeparatorIcon != "" {
			props.values[FolderSeparatorIcon] = tc.FolderSeparatorIcon
		}
		if tc.DisableMappedLocations {
			props.values[MappedLocationsEnabled] = false
		}
		path := &path{
			env:   env,
			props: props,
		}
		got := path.string()
		assert.Equal(t, tc.Expected, got)
	}
}

func TestGetFolderPathCustomMappedLocations(t *testing.T) {
	pwd := "/a/b/c/d"
	env := new(MockedEnvironment)
	env.On("getPathSeperator", nil).Return("/")
	env.On("homeDir", nil).Return("/usr/home")
	env.On("getcwd", nil).Return(pwd)
	args := &args{
		PSWD: &pwd,
	}
	env.On("getArgs", nil).Return(args)
	path := &path{
		env: env,
		props: &properties{
			values: map[Property]interface{}{
				MappedLocations: map[string]string{
					"/a/b/c/d": "#",
				},
			},
		},
	}
	got := path.getFolderPath()
	assert.Equal(t, "#", got)
}

func testWritePathInfo(home, pwd, pathSeparator string) string {
	props := &properties{
		values: map[Property]interface{}{
			FolderSeparatorIcon: " > ",
			FolderIcon:          "f",
			HomeIcon:            "~",
		},
	}
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return(home)
	env.On("getPathSeperator", nil).Return(pathSeparator)
	env.On("getcwd", nil).Return(pwd)
	args := &args{
		PSWD: &pwd,
	}
	env.On("getArgs", nil).Return(args)
	path := &path{
		env:   env,
		props: props,
	}
	return path.getAgnosterPath()
}

func TestWritePathInfoWindowsOutsideHome(t *testing.T) {
	home := homeBillWindows
	want := "C: > f > f > location"
	got := testWritePathInfo(home, "C:\\Program Files\\Go\\location", "\\")
	assert.Equal(t, want, got)
}

func TestWritePathInfoWindowsInsideHome(t *testing.T) {
	home := homeBillWindows
	location := home + "\\Documents\\Bill\\location"
	want := "~ > f > f > location"
	got := testWritePathInfo(home, location, "\\")
	assert.Equal(t, want, got)
}

func TestWritePathInfoWindowsOutsideHomeZeroLevels(t *testing.T) {
	home := homeBillWindows
	want := "C: > location"
	got := testWritePathInfo(home, "C:\\location", "\\")
	assert.Equal(t, want, got)
}

func TestWritePathInfoWindowsOutsideHomeOneLevels(t *testing.T) {
	home := homeBillWindows
	want := "C: > f > location"
	got := testWritePathInfo(home, "C:\\Program Files\\location", "\\")
	assert.Equal(t, want, got)
}

func TestWritePathInfoUnixOutsideHome(t *testing.T) {
	home := homeJan
	want := "mnt > f > f > location"
	got := testWritePathInfo(home, "/mnt/go/test/location", "/")
	assert.Equal(t, want, got)
}

func TestWritePathInfoUnixInsideHome(t *testing.T) {
	home := homeJan
	location := home + "/docs/jan/location"
	want := "~ > f > f > location"
	got := testWritePathInfo(home, location, "/")
	assert.Equal(t, want, got)
}

func TestWritePathInfoUnixOutsideHomeZeroLevels(t *testing.T) {
	home := homeJan
	want := "mnt > location"
	got := testWritePathInfo(home, "/mnt/location", "/")
	assert.Equal(t, want, got)
}

func TestWritePathInfoUnixOutsideHomeOneLevels(t *testing.T) {
	home := homeJan
	want := "mnt > f > location"
	got := testWritePathInfo(home, "/mnt/folder/location", "/")
	assert.Equal(t, want, got)
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
		{MappedLocationsEnabled: false, Pwd: "/a/b/c/d/e/f/g", Expected: "/a/b/c/d/e/f/g"},

		{MappedLocationsEnabled: true, Pwd: "/w/d/x/w", Pswd: "/z/y/x/w", Expected: "/z/y/x/w"},
		{MappedLocationsEnabled: false, Pwd: "/f/g/k/d/e/f/g", Pswd: "/a/b/c/d/e/f/g", Expected: "/a/b/c/d/e/f/g"},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getPathSeperator", nil).Return("/")
		env.On("homeDir", nil).Return("/usr/home")
		env.On("getcwd", nil).Return(tc.Pwd)
		args := &args{
			PSWD: &tc.Pswd,
		}
		env.On("getArgs", nil).Return(args)
		path := &path{
			env: env,
			props: &properties{
				values: map[Property]interface{}{
					MappedLocationsEnabled: tc.MappedLocationsEnabled,
					MappedLocations: map[string]string{
						"/a/b/c/d": "#",
					},
				},
			},
		}
		got := path.getPwd()
		assert.Equal(t, tc.Expected, got)
	}
}
