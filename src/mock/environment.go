package mock

import (
	"io"
	"io/fs"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/platform/battery"

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

func (env *MockedEnvironment) ResolveSymlink(path string) (string, error) {
	args := env.Called(path)
	return args.String(0), args.Error(1)
}

func (env *MockedEnvironment) FileContent(file string) string {
	args := env.Called(file)
	return args.String(0)
}

func (env *MockedEnvironment) LsDir(path string) []fs.DirEntry {
	args := env.Called(path)
	return args.Get(0).([]fs.DirEntry)
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

func (env *MockedEnvironment) CommandPath(command string) string {
	args := env.Called(command)
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

func (env *MockedEnvironment) StatusCodes() (int, string) {
	args := env.Called()
	return args.Int(0), args.String(1)
}

func (env *MockedEnvironment) ExecutionTime() float64 {
	args := env.Called()
	return float64(args.Int(0))
}

func (env *MockedEnvironment) Root() bool {
	args := env.Called()
	return args.Bool(0)
}

func (env *MockedEnvironment) Flags() *platform.Flags {
	arguments := env.Called()
	return arguments.Get(0).(*platform.Flags)
}

func (env *MockedEnvironment) BatteryState() (*battery.Info, error) {
	args := env.Called()
	return args.Get(0).(*battery.Info), args.Error(1)
}

func (env *MockedEnvironment) Shell() string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) QueryWindowTitles(processName, windowTitleRegex string) (string, error) {
	args := env.Called(processName, windowTitleRegex)
	return args.String(0), args.Error(1)
}

func (env *MockedEnvironment) WindowsRegistryKeyValue(path string) (*platform.WindowsRegistryValue, error) {
	args := env.Called(path)
	return args.Get(0).(*platform.WindowsRegistryValue), args.Error(1)
}

func (env *MockedEnvironment) HTTPRequest(url string, _ io.Reader, _ int, _ ...platform.HTTPRequestModifier) ([]byte, error) {
	args := env.Called(url)
	return args.Get(0).([]byte), args.Error(1)
}

func (env *MockedEnvironment) HasParentFilePath(path string) (*platform.FileInfo, error) {
	args := env.Called(path)
	return args.Get(0).(*platform.FileInfo), args.Error(1)
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

func (env *MockedEnvironment) Cache() platform.Cache {
	args := env.Called()
	return args.Get(0).(platform.Cache)
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

func (env *MockedEnvironment) ConvertToWindowsPath(_ string) string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) ConvertToLinuxPath(_ string) string {
	args := env.Called()
	return args.String(0)
}

func (env *MockedEnvironment) Connection(connectionType platform.ConnectionType) (*platform.Connection, error) {
	args := env.Called(connectionType)
	return args.Get(0).(*platform.Connection), args.Error(1)
}

func (env *MockedEnvironment) TemplateCache() *platform.TemplateCache {
	args := env.Called()
	return args.Get(0).(*platform.TemplateCache)
}

func (env *MockedEnvironment) LoadTemplateCache() {
	_ = env.Called()
}

func (env *MockedEnvironment) MockGitCommand(dir, returnValue string, args ...string) {
	args = append([]string{"-C", dir, "--no-optional-locks", "-c", "core.quotepath=false", "-c", "color.status=false"}, args...)
	env.On("RunCommand", "git", args).Return(returnValue, nil)
}

func (env *MockedEnvironment) MockHgCommand(dir, returnValue string, args ...string) {
	args = append([]string{"-R", dir}, args...)
	env.On("RunCommand", "hg", args).Return(returnValue, nil)
}

func (env *MockedEnvironment) MockSvnCommand(dir, returnValue string, args ...string) {
	args = append([]string{"-C", dir, "--no-optional-locks", "-c", "core.quotepath=false", "-c", "color.status=false"}, args...)
	env.On("RunCommand", "svn", args).Return(returnValue, nil)
}

func (env *MockedEnvironment) HasFileInParentDirs(pattern string, depth uint) bool {
	args := env.Called(pattern, depth)
	return args.Bool(0)
}

func (env *MockedEnvironment) DirMatchesOneOf(dir string, regexes []string) bool {
	args := env.Called(dir, regexes)
	return args.Bool(0)
}

func (env *MockedEnvironment) Trace(start time.Time, args ...string) {
	_ = env.Called(start, args)
}

func (env *MockedEnvironment) Debug(message string) {
	_ = env.Called(message)
}

func (env *MockedEnvironment) DebugF(format string, a ...any) {
	_ = env.Called(format, a)
}

func (env *MockedEnvironment) Error(err error) {
	_ = env.Called(err)
}

func (env *MockedEnvironment) DirIsWritable(path string) bool {
	args := env.Called(path)
	return args.Bool(0)
}

func (env *MockedEnvironment) SetPromptCount() {
	_ = env.Called()
}

func (env *MockedEnvironment) CursorPosition() (int, int) {
	args := env.Called()
	return args.Int(0), args.Int(1)
}

func (env *MockedEnvironment) SystemInfo() (*platform.SystemInfo, error) {
	args := env.Called()
	return args.Get(0).(*platform.SystemInfo), args.Error(1)
}

func (env *MockedEnvironment) Unset(name string) {
	for i := 0; i < len(env.ExpectedCalls); i++ {
		f := env.ExpectedCalls[i]
		if f.Method == name {
			f.Unset()
		}
	}
}
