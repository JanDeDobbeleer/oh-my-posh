package main

import (
	"math/rand"
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

func (env *MockedEnvironment) runShellCommand(shell string, command string) string {
	args := env.Called(shell, command)
	return args.String(0)
}

func (env *MockedEnvironment) lastErrorCode() int {
	args := env.Called(nil)
	return args.Int(0)
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

func (env *MockedEnvironment) getWindowTitle(imageName string, windowTitleRegex string) (string, error) {
	args := env.Called(imageName)
	return args.String(0), args.Error(1)
}

func TestIsInHomeDirTrue(t *testing.T) {
	home := "/home/bill"
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return(home)
	path := &path{
		env: env,
	}
	got := path.inHomeDir(home)
	assert.True(t, got)
}

func TestIsInHomeDirLevelTrue(t *testing.T) {
	level := rand.Intn(100)
	home := "/home/bill"
	pwd := home
	for i := 0; i < level; i++ {
		pwd += "/level"
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
	env.On("homeDir", nil).Return("/home/bill")
	env.On("getPathSeperator", nil).Return("/")
	path := &path{
		env:   env,
		props: props,
	}
	got := path.rootLocation(home)
	assert.EqualValues(t, expected, got)
}

func TestRootLocationOutsideHome(t *testing.T) {
	props := &properties{
		values: map[Property]interface{}{HomeIcon: "~"},
	}
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return("/home/bill")
	env.On("getPathSeperator", nil).Return("/")
	path := &path{
		env:   env,
		props: props,
	}
	got := path.rootLocation("/usr/error/what")
	assert.EqualValues(t, "usr", got)
}

func TestRootLocationWindowsDrive(t *testing.T) {
	props := &properties{
		values: map[Property]interface{}{HomeIcon: "~"},
	}
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return("C:\\Users\\Bill")
	env.On("getPathSeperator", nil).Return("\\")
	path := &path{
		env:   env,
		props: props,
	}
	got := path.rootLocation("C:\\Program Files\\Go")
	assert.EqualValues(t, "C:", got)
}

func TestRootLocationWindowsRegistry(t *testing.T) {
	expected := "REG"
	props := &properties{
		values: map[Property]interface{}{WindowsRegistryIcon: expected},
	}
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return("C:\\Users\\Bill")
	env.On("getPathSeperator", nil).Return("\\")
	path := &path{
		env:   env,
		props: props,
	}
	got := path.rootLocation("HKCU:\\Program Files\\Go")
	assert.EqualValues(t, expected, got)
}

func TestRootLocationWindowsPowerShellHome(t *testing.T) {
	expected := "~"
	props := &properties{
		values: map[Property]interface{}{HomeIcon: expected},
	}
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return("C:\\Users\\Bill")
	env.On("getPathSeperator", nil).Return("\\")
	path := &path{
		env:   env,
		props: props,
	}
	got := path.rootLocation("Microsoft.PowerShell.Core\\FileSystem::C:\\Users\\Bill")
	assert.EqualValues(t, expected, got)
}

func TestRootLocationWindowsPowerShellOutsideHome(t *testing.T) {
	props := &properties{
		values: map[Property]interface{}{WindowsRegistryIcon: "REG"},
	}
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return("C:\\Program Files\\Go")
	env.On("getPathSeperator", nil).Return("\\")
	path := &path{
		env:   env,
		props: props,
	}
	got := path.rootLocation("Microsoft.PowerShell.Core\\FileSystem::C:\\Users\\Bill")
	assert.EqualValues(t, "C:", got)
}

func TestRootLocationEmptyDir(t *testing.T) {
	props := &properties{
		values: map[Property]interface{}{WindowsRegistryIcon: "REG"},
	}
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return("/home/bill")
	env.On("getPathSeperator", nil).Return("/")
	path := &path{
		env:   env,
		props: props,
	}
	got := path.rootLocation("")
	assert.EqualValues(t, "", got)
}

func TestIsInHomeDirFalse(t *testing.T) {
	home := "/home/bill"
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return(home)
	path := &path{
		env: env,
	}
	got := path.inHomeDir("/usr/error")
	assert.False(t, got)
}

func TestPathDepthInHome(t *testing.T) {
	home := "/home/bill"
	pwd := home
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return(home)
	env.On("getPathSeperator", nil).Return("/")
	path := &path{
		env: env,
	}
	got := path.pathDepth(pwd)
	assert.Equal(t, 0, got)
}

func TestPathDepthInHomeTrailing(t *testing.T) {
	home := "/home/bill/"
	pwd := home + "/"
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return(home)
	env.On("getPathSeperator", nil).Return("/")
	path := &path{
		env: env,
	}
	got := path.pathDepth(pwd)
	assert.Equal(t, 0, got)
}

func TestPathDepthInHomeMultipleLevelsDeep(t *testing.T) {
	level := rand.Intn(100)
	home := "/home/bill"
	pwd := home
	for i := 0; i < level; i++ {
		pwd += "/level"
	}
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return(home)
	env.On("getPathSeperator", nil).Return("/")
	path := &path{
		env: env,
	}
	got := path.pathDepth(pwd)
	assert.Equal(t, level, got)
}

func TestPathDepthOutsideHomeMultipleLevelsDeep(t *testing.T) {
	level := rand.Intn(100)
	home := "/home/gates"
	pwd := "/usr"
	for i := 0; i < level; i++ {
		pwd += "/level"
	}
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return(home)
	env.On("getPathSeperator", nil).Return("/")
	path := &path{
		env: env,
	}
	got := path.pathDepth(pwd)
	assert.Equal(t, level, got)
}

func TestPathDepthOutsideHomeZeroLevelsDeep(t *testing.T) {
	home := "/home/gates"
	pwd := "/usr/"
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return(home)
	env.On("getPathSeperator", nil).Return("/")
	path := &path{
		env: env,
	}
	got := path.pathDepth(pwd)
	assert.Equal(t, 0, got)
}

func TestPathDepthOutsideHomeOneLevelDeep(t *testing.T) {
	home := "/home/gates"
	pwd := "/usr/location"
	env := new(MockedEnvironment)
	env.On("homeDir", nil).Return(home)
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

func testWritePathInfo(home string, pwd string, pathSeparator string) string {
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
	home := "C:\\Users\\Bill"
	want := "C: > f > f > location"
	got := testWritePathInfo(home, "C:\\Program Files\\Go\\location", "\\")
	assert.Equal(t, want, got)
}

func TestWritePathInfoWindowsInsideHome(t *testing.T) {
	home := "C:\\Users\\Bill"
	location := home + "\\Documents\\Bill\\location"
	want := "~ > f > f > location"
	got := testWritePathInfo(home, location, "\\")
	assert.Equal(t, want, got)
}

func TestWritePathInfoWindowsOutsideHomeZeroLevels(t *testing.T) {
	home := "C:\\Users\\Bill"
	want := "C: > location"
	got := testWritePathInfo(home, "C:\\location", "\\")
	assert.Equal(t, want, got)
}

func TestWritePathInfoWindowsOutsideHomeOneLevels(t *testing.T) {
	home := "C:\\Users\\Bill"
	want := "C: > f > location"
	got := testWritePathInfo(home, "C:\\Program Files\\location", "\\")
	assert.Equal(t, want, got)
}

func TestWritePathInfoUnixOutsideHome(t *testing.T) {
	home := "/usr/home/jan"
	want := "mnt > f > f > location"
	got := testWritePathInfo(home, "/mnt/go/test/location", "/")
	assert.Equal(t, want, got)
}

func TestWritePathInfoUnixInsideHome(t *testing.T) {
	home := "/usr/home/jan"
	location := home + "/docs/jan/location"
	want := "~ > f > f > location"
	got := testWritePathInfo(home, location, "/")
	assert.Equal(t, want, got)
}

func TestWritePathInfoUnixOutsideHomeZeroLevels(t *testing.T) {
	home := "/usr/home/jan"
	want := "mnt > location"
	got := testWritePathInfo(home, "/mnt/location", "/")
	assert.Equal(t, want, got)
}

func TestWritePathInfoUnixOutsideHomeOneLevels(t *testing.T) {
	home := "/usr/home/jan"
	want := "mnt > f > location"
	got := testWritePathInfo(home, "/mnt/folder/location", "/")
	assert.Equal(t, want, got)
}
