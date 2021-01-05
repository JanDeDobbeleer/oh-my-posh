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
	expected := "~"
	props := &properties{
		values: map[Property]interface{}{HomeIcon: expected},
	}
	home := "/home/bill/"
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return(home)
	env.On("getcwd", nil).Return(home)
	env.On("getPathSeperator", nil).Return("/")
	path := &path{
		env:   env,
		props: props,
	}
	got := path.rootLocation()
	assert.EqualValues(t, expected, got)
}

func TestRootLocationOutsideHome(t *testing.T) {
	props := &properties{
		values: map[Property]interface{}{HomeIcon: "~"},
	}
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return("/home/bill")
	env.On("getcwd", nil).Return("/usr/error/what")
	env.On("getPathSeperator", nil).Return("/")
	path := &path{
		env:   env,
		props: props,
	}
	got := path.rootLocation()
	assert.EqualValues(t, "usr", got)
}

func TestRootLocationWindowsDrive(t *testing.T) {
	props := &properties{
		values: map[Property]interface{}{HomeIcon: "~"},
	}
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return("C:\\Users\\Bill")
	env.On("getcwd", nil).Return("C:\\Program Files\\Go")
	env.On("getPathSeperator", nil).Return("\\")
	path := &path{
		env:   env,
		props: props,
	}
	got := path.rootLocation()
	assert.EqualValues(t, "C:", got)
}

func TestRootLocationWindowsRegistry(t *testing.T) {
	expected := "REG"
	props := &properties{
		values: map[Property]interface{}{WindowsRegistryIcon: expected},
	}
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return("C:\\Users\\Bill")
	env.On("getcwd", nil).Return("HKCU:\\Program Files\\Go")
	env.On("getPathSeperator", nil).Return("\\")
	path := &path{
		env:   env,
		props: props,
	}
	got := path.rootLocation()
	assert.EqualValues(t, expected, got)
}

func TestRootLocationWindowsPowerShellHome(t *testing.T) {
	expected := "~"
	props := &properties{
		values: map[Property]interface{}{HomeIcon: expected},
	}
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return("C:\\Users\\Bill")
	env.On("getcwd", nil).Return("Microsoft.PowerShell.Core\\FileSystem::C:\\Users\\Bill")
	env.On("getPathSeperator", nil).Return("\\")
	path := &path{
		env:   env,
		props: props,
	}
	got := path.rootLocation()
	assert.EqualValues(t, expected, got)
}

func TestRootLocationWindowsPowerShellOutsideHome(t *testing.T) {
	props := &properties{
		values: map[Property]interface{}{WindowsRegistryIcon: "REG"},
	}
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return("C:\\Program Files\\Go")
	env.On("getcwd", nil).Return("Microsoft.PowerShell.Core\\FileSystem::C:\\Users\\Bill")
	env.On("getPathSeperator", nil).Return("\\")
	path := &path{
		env:   env,
		props: props,
	}
	got := path.rootLocation()
	assert.EqualValues(t, "C:", got)
}

func TestRootLocationEmptyDir(t *testing.T) {
	props := &properties{
		values: map[Property]interface{}{WindowsRegistryIcon: "REG"},
	}
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return("/home/bill")
	env.On("getcwd", nil).Return("")
	env.On("getPathSeperator", nil).Return("/")
	path := &path{
		env:   env,
		props: props,
	}
	got := path.rootLocation()
	assert.EqualValues(t, "", got)
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

func TestGetAgnosterFullPath(t *testing.T) {
	pwd := "/usr/location/whatever"
	env := new(MockedEnvironment)
	env.On("getPathSeperator", nil).Return("/")
	env.On("homeDir", nil).Return("/usr/home")
	env.On("getcwd", nil).Return(pwd)
	path := &path{
		env: env,
		props: &properties{
			values: map[Property]interface{}{
				FolderSeparatorIcon: " > ",
			},
		},
	}
	got := path.getAgnosterFullPath()
	assert.Equal(t, "usr > location > whatever", got)
}

func TestGetAgnosterShortPath(t *testing.T) {
	pwd := "/usr/location/whatever/man"
	env := new(MockedEnvironment)
	env.On("getPathSeperator", nil).Return("/")
	env.On("homeDir", nil).Return("/usr/home")
	env.On("getcwd", nil).Return(pwd)
	path := &path{
		env: env,
		props: &properties{
			values: map[Property]interface{}{
				FolderSeparatorIcon: " > ",
			},
		},
	}
	got := path.getAgnosterShortPath()
	assert.Equal(t, "usr > .. > man", got)
}

func TestGetAgnosterShortPathInsideHome(t *testing.T) {
	pwd := "/usr/home/whatever/man"
	env := new(MockedEnvironment)
	env.On("getPathSeperator", nil).Return("/")
	env.On("homeDir", nil).Return("/usr/home")
	env.On("getcwd", nil).Return(pwd)
	path := &path{
		env: env,
		props: &properties{
			values: map[Property]interface{}{
				FolderSeparatorIcon: " > ",
			},
		},
	}
	got := path.getAgnosterShortPath()
	assert.Equal(t, "~ > .. > man", got)
}

func TestGetAgnosterShortPathInsideHomeOneLevel(t *testing.T) {
	pwd := "/usr/home/projects"
	env := new(MockedEnvironment)
	env.On("getPathSeperator", nil).Return("/")
	env.On("homeDir", nil).Return("/usr/home")
	env.On("getcwd", nil).Return(pwd)
	path := &path{
		env: env,
		props: &properties{
			values: map[Property]interface{}{
				FolderSeparatorIcon: " > ",
			},
		},
	}
	got := path.getAgnosterShortPath()
	assert.Equal(t, "~ > projects", got)
}

func TestGetAgnosterShortPathZeroLevelsWindows(t *testing.T) {
	pwd := "C:"
	env := new(MockedEnvironment)
	env.On("getPathSeperator", nil).Return("\\")
	env.On("homeDir", nil).Return(homeBillWindows)
	env.On("getcwd", nil).Return(pwd)
	path := &path{
		env: env,
		props: &properties{
			values: map[Property]interface{}{
				FolderSeparatorIcon: " > ",
			},
		},
	}
	got := path.getAgnosterShortPath()
	assert.Equal(t, "C:", got)
}

func TestGetAgnosterShortPathZeroLevelsLinux(t *testing.T) {
	pwd := "/"
	env := new(MockedEnvironment)
	env.On("getPathSeperator", nil).Return("/")
	env.On("homeDir", nil).Return(homeBillWindows)
	env.On("getcwd", nil).Return(pwd)
	path := &path{
		env: env,
		props: &properties{
			values: map[Property]interface{}{
				FolderSeparatorIcon: " > ",
			},
		},
	}
	got := path.getAgnosterShortPath()
	assert.Equal(t, "", got)
}

func TestGetAgnosterShortPathOneLevel(t *testing.T) {
	pwd := "/foo"
	env := new(MockedEnvironment)
	env.On("getPathSeperator", nil).Return("/")
	env.On("homeDir", nil).Return(homeBillWindows)
	env.On("getcwd", nil).Return(pwd)
	path := &path{
		env: env,
		props: &properties{
			values: map[Property]interface{}{
				FolderSeparatorIcon: " > ",
			},
		},
	}
	got := path.getAgnosterShortPath()
	assert.Equal(t, "foo", got)
}

// nolint:dupl // false positive
func TestGetFullPath(t *testing.T) {
	cases := []struct {
		UseFolderSeparatorIcon bool
		Pwd                    string
		Expected               string
	}{
		{UseFolderSeparatorIcon: false, Pwd: "", Expected: ""},
		{UseFolderSeparatorIcon: false, Pwd: "/", Expected: "/"},
		{UseFolderSeparatorIcon: false, Pwd: "/usr/home", Expected: "~"},
		{UseFolderSeparatorIcon: false, Pwd: "/usr/home/abc", Expected: "~/abc"},
		{UseFolderSeparatorIcon: false, Pwd: "/a/b/c/d", Expected: "/a/b/c/d"},

		{UseFolderSeparatorIcon: true, Pwd: "", Expected: ""},
		{UseFolderSeparatorIcon: true, Pwd: "/", Expected: "|"},
		{UseFolderSeparatorIcon: true, Pwd: "/usr/home", Expected: "~"},
		{UseFolderSeparatorIcon: true, Pwd: "/usr/home/abc", Expected: "~|abc"},
		{UseFolderSeparatorIcon: true, Pwd: "/a/b/c/d", Expected: "|a|b|c|d"},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getPathSeperator", nil).Return("/")
		env.On("homeDir", nil).Return("/usr/home")
		env.On("getcwd", nil).Return(tc.Pwd)
		props := map[Property]interface{}{}
		if tc.UseFolderSeparatorIcon {
			props[FolderSeparatorIcon] = "|"
		}
		path := &path{
			env: env,
			props: &properties{
				values: props,
			},
		}
		got := path.getFullPath()
		assert.Equal(t, tc.Expected, got)
	}
}

// nolint:dupl // false positive
func TestGetFolderPath(t *testing.T) {
	cases := []struct {
		UseFolderSeparatorIcon bool
		Pwd                    string
		Expected               string
	}{
		{UseFolderSeparatorIcon: false, Pwd: "", Expected: "."},
		{UseFolderSeparatorIcon: false, Pwd: "/", Expected: "/"},
		{UseFolderSeparatorIcon: false, Pwd: "/usr/home", Expected: "~"},
		{UseFolderSeparatorIcon: false, Pwd: "/usr/home/abc", Expected: "abc"},
		{UseFolderSeparatorIcon: false, Pwd: "/a/b/c/d", Expected: "d"},

		{UseFolderSeparatorIcon: true, Pwd: "", Expected: "."},
		{UseFolderSeparatorIcon: true, Pwd: "/", Expected: "|"},
		{UseFolderSeparatorIcon: true, Pwd: "/usr/home", Expected: "~"},
		{UseFolderSeparatorIcon: true, Pwd: "/usr/home/abc", Expected: "abc"},
		{UseFolderSeparatorIcon: true, Pwd: "/a/b/c/d", Expected: "d"},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getPathSeperator", nil).Return("/")
		env.On("homeDir", nil).Return("/usr/home")
		env.On("getcwd", nil).Return(tc.Pwd)
		props := map[Property]interface{}{}
		if tc.UseFolderSeparatorIcon {
			props[FolderSeparatorIcon] = "|"
		}
		path := &path{
			env: env,
			props: &properties{
				values: props,
			},
		}
		got := path.getFolderPath()
		assert.Equal(t, tc.Expected, got)
	}
}

func TestGetFolderPathCustomMappedLocations(t *testing.T) {
	pwd := "/a/b/c/d"
	env := new(MockedEnvironment)
	env.On("getPathSeperator", nil).Return("/")
	env.On("homeDir", nil).Return("/usr/home")
	env.On("getcwd", nil).Return(pwd)
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
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getPathSeperator", nil).Return("/")
		env.On("homeDir", nil).Return("/usr/home")
		env.On("getcwd", nil).Return(tc.Pwd)
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
