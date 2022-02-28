package mock

import (
	"oh-my-posh/environment"

	"github.com/distatus/battery"
	mock "github.com/stretchr/testify/mock"
)

type MockedEnvironment struct {
	mock.Mock
}

func (env *MockedEnvironment) Getenv(key string) string {
	args := env.Called(key)
	return args.String(0)
}

func (env *MockedEnvironment) Pwd() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) Home() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) HasFiles(pattern string) bool {
	args := env.Called(pattern)
	return args.Bool(0)
}

func (env *MockedEnvironment) HasFilesInDir(dir, pattern string) bool {
	args := env.Called(dir, pattern)
	return args.Bool(0)
}

func (env *MockedEnvironment) HasFolder(folder string) bool {
	args := env.Called(folder)
	return args.Bool(0)
}

func (env *MockedEnvironment) FileContent(file string) string {
	args := env.Called(file)
	return args.String(0)
}

func (env *MockedEnvironment) FolderList(path string) []string {
	args := env.Called(path)
	return args.Get(0).([]string)
}

func (env *MockedEnvironment) PathSeparator() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) User() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) Host() (string, error) {
	args := env.Called()
	return args.String(0), args.Error(1)
}

func (env *MockedEnvironment) GOOS() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) Platform() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) HasCommand(command string) bool {
	args := env.Called(command)
	return args.Bool(0)
}

func (env *MockedEnvironment) RunCommand(command string, args ...string) (string, error) {
	arguments := env.Called(command, args)
	return arguments.String(0), arguments.Error(1)
}

func (env *MockedEnvironment) RunShellCommand(shell, command string) string {
	args := env.Called(shell, command)
	return args.String(0)
}

func (env *MockedEnvironment) ErrorCode() int {
	args := env.Called()
	return args.Int(0)
}

func (env *MockedEnvironment) ExecutionTime() float64 {
	args := env.Called()
	return float64(args.Int(0))
}

func (env *MockedEnvironment) Root() bool {
	args := env.Called()
	return args.Bool(0)
}

func (env *MockedEnvironment) Args() *environment.Args {
	arguments := env.Called()
	return arguments.Get(0).(*environment.Args)
}

func (env *MockedEnvironment) BatteryInfo() ([]*battery.Battery, error) {
	args := env.Called()
	return args.Get(0).([]*battery.Battery), args.Error(1)
}

func (env *MockedEnvironment) Shell() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) WindowTitle(imageName, windowTitleRegex string) (string, error) {
	args := env.Called(imageName)
	return args.String(0), args.Error(1)
}

func (env *MockedEnvironment) WindowsRegistryKeyValue(path string) (*environment.WindowsRegistryValue, error) {
	args := env.Called(path)
	return args.Get(0).(*environment.WindowsRegistryValue), args.Error(1)
}

func (env *MockedEnvironment) HTTPRequest(url string, timeout int, requestModifiers ...environment.HTTPRequestModifier) ([]byte, error) {
	args := env.Called(url)
	return args.Get(0).([]byte), args.Error(1)
}

func (env *MockedEnvironment) HasParentFilePath(path string) (*environment.FileInfo, error) {
	args := env.Called(path)
	return args.Get(0).(*environment.FileInfo), args.Error(1)
}

func (env *MockedEnvironment) StackCount() int {
	args := env.Called()
	return args.Int(0)
}

func (env *MockedEnvironment) IsWsl() bool {
	args := env.Called()
	return args.Bool(0)
}

func (env *MockedEnvironment) IsWsl2() bool {
	args := env.Called()
	return args.Bool(0)
}

func (env *MockedEnvironment) TerminalWidth() (int, error) {
	args := env.Called()
	return args.Int(0), args.Error(1)
}

func (env *MockedEnvironment) CachePath() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) Cache() environment.Cache {
	args := env.Called()
	return args.Get(0).(environment.Cache)
}

func (env *MockedEnvironment) Close() {
	_ = env.Called()
}

func (env *MockedEnvironment) Logs() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) InWSLSharedDrive() bool {
	args := env.Called()
	return args.Bool(0)
}

func (env *MockedEnvironment) ConvertToWindowsPath(path string) string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) ConvertToLinuxPath(path string) string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) WifiNetwork() (*environment.WifiInfo, error) {
	args := env.Called()
	return args.Get(0).(*environment.WifiInfo), args.Error(1)
}

func (env *MockedEnvironment) TemplateCache() *environment.TemplateCache {
	args := env.Called()
	return args.Get(0).(*environment.TemplateCache)
}

func (env *MockedEnvironment) MockGitCommand(dir, returnValue string, args ...string) {
	args = append([]string{"-C", dir, "--no-optional-locks", "-c", "core.quotepath=false", "-c", "color.status=false"}, args...)
	env.On("RunCommand", "git", args).Return(returnValue, nil)
}

func (env *MockedEnvironment) HasFileInParentDirs(pattern string, depth uint) bool {
	args := env.Called(pattern, depth)
	return args.Bool(0)
}
