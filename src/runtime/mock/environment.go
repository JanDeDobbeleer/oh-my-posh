package mock

import (
	"io"
	"io/fs"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/battery"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"

	mock "github.com/stretchr/testify/mock"
)

type Environment struct {
	mock.Mock
}

func (env *Environment) Getenv(key string) string {
	args := env.Called(key)
	return args.String(0)
}

func (env *Environment) Pwd() string {
	args := env.Called()
	return args.String(0)
}

func (env *Environment) Home() string {
	args := env.Called()
	return args.String(0)
}

func (env *Environment) HasFiles(pattern string) bool {
	args := env.Called(pattern)
	return args.Bool(0)
}

func (env *Environment) HasFilesInDir(dir, pattern string) bool {
	args := env.Called(dir, pattern)
	return args.Bool(0)
}

func (env *Environment) HasFolder(folder string) bool {
	args := env.Called(folder)
	return args.Bool(0)
}

func (env *Environment) ResolveSymlink(input string) (string, error) {
	args := env.Called(input)
	return args.String(0), args.Error(1)
}

func (env *Environment) FileContent(file string) string {
	args := env.Called(file)
	return args.String(0)
}

func (env *Environment) LsDir(input string) []fs.DirEntry {
	args := env.Called(input)
	return args.Get(0).([]fs.DirEntry)
}

func (env *Environment) User() string {
	args := env.Called()
	return args.String(0)
}

func (env *Environment) Host() (string, error) {
	args := env.Called()
	return args.String(0), args.Error(1)
}

func (env *Environment) GOOS() string {
	args := env.Called()
	return args.String(0)
}

func (env *Environment) Platform() string {
	args := env.Called()
	return args.String(0)
}

func (env *Environment) CommandPath(command string) string {
	args := env.Called(command)
	return args.String(0)
}

func (env *Environment) HasCommand(command string) bool {
	args := env.Called(command)
	return args.Bool(0)
}

func (env *Environment) RunCommand(command string, args ...string) (string, error) {
	arguments := env.Called(command, args)
	return arguments.String(0), arguments.Error(1)
}

func (env *Environment) RunShellCommand(shell, command string) string {
	args := env.Called(shell, command)
	return args.String(0)
}

func (env *Environment) StatusCodes() (int, string) {
	args := env.Called()
	return args.Int(0), args.String(1)
}

func (env *Environment) ExecutionTime() float64 {
	args := env.Called()
	return float64(args.Int(0))
}

func (env *Environment) Root() bool {
	args := env.Called()
	return args.Bool(0)
}

func (env *Environment) Flags() *runtime.Flags {
	arguments := env.Called()
	return arguments.Get(0).(*runtime.Flags)
}

func (env *Environment) BatteryState() (*battery.Info, error) {
	args := env.Called()
	return args.Get(0).(*battery.Info), args.Error(1)
}

func (env *Environment) Shell() string {
	args := env.Called()
	return args.String(0)
}

func (env *Environment) QueryWindowTitles(processName, windowTitleRegex string) (string, error) {
	args := env.Called(processName, windowTitleRegex)
	return args.String(0), args.Error(1)
}

func (env *Environment) WindowsRegistryKeyValue(path string) (*runtime.WindowsRegistryValue, error) {
	args := env.Called(path)
	return args.Get(0).(*runtime.WindowsRegistryValue), args.Error(1)
}

func (env *Environment) HTTPRequest(url string, _ io.Reader, _ int, _ ...http.RequestModifier) ([]byte, error) {
	args := env.Called(url)
	return args.Get(0).([]byte), args.Error(1)
}

func (env *Environment) HasParentFilePath(parent string, followSymlinks bool) (*runtime.FileInfo, error) {
	args := env.Called(parent, followSymlinks)
	return args.Get(0).(*runtime.FileInfo), args.Error(1)
}

func (env *Environment) StackCount() int {
	args := env.Called()
	return args.Int(0)
}

func (env *Environment) IsWsl() bool {
	args := env.Called()
	return args.Bool(0)
}

func (env *Environment) IsWsl2() bool {
	args := env.Called()
	return args.Bool(0)
}

func (env *Environment) IsCygwin() bool {
	args := env.Called()
	return args.Bool(0)
}

func (env *Environment) TerminalWidth() (int, error) {
	args := env.Called()
	return args.Int(0), args.Error(1)
}

func (env *Environment) CachePath() string {
	args := env.Called()
	return args.String(0)
}

func (env *Environment) Cache() cache.Cache {
	args := env.Called()
	return args.Get(0).(cache.Cache)
}

func (env *Environment) Session() cache.Cache {
	args := env.Called()
	return args.Get(0).(cache.Cache)
}

func (env *Environment) Close() {
	_ = env.Called()
}

func (env *Environment) Logs() string {
	args := env.Called()
	return args.String(0)
}

func (env *Environment) InWSLSharedDrive() bool {
	args := env.Called()
	return args.Bool(0)
}

func (env *Environment) ConvertToWindowsPath(_ string) string {
	args := env.Called()
	return args.String(0)
}

func (env *Environment) ConvertToLinuxPath(_ string) string {
	args := env.Called()
	return args.String(0)
}

func (env *Environment) Connection(connectionType runtime.ConnectionType) (*runtime.Connection, error) {
	args := env.Called(connectionType)
	return args.Get(0).(*runtime.Connection), args.Error(1)
}

func (env *Environment) MockGitCommand(dir, returnValue string, args ...string) {
	args = append([]string{"-C", dir, "--no-optional-locks", "-c", "core.quotepath=false", "-c", "color.status=false"}, args...)
	env.On("RunCommand", "git", args).Return(returnValue, nil)
}

func (env *Environment) MockHgCommand(dir, returnValue string, args ...string) {
	args = append([]string{"-R", dir}, args...)
	env.On("RunCommand", "hg", args).Return(returnValue, nil)
}

func (env *Environment) MockSvnCommand(dir, returnValue string, args ...string) {
	args = append([]string{"-C", dir, "--no-optional-locks", "-c", "core.quotepath=false", "-c", "color.status=false"}, args...)
	env.On("RunCommand", "svn", args).Return(returnValue, nil)
}

func (env *Environment) HasFileInParentDirs(pattern string, depth uint) bool {
	args := env.Called(pattern, depth)
	return args.Bool(0)
}

func (env *Environment) DirMatchesOneOf(dir string, regexes []string) bool {
	args := env.Called(dir, regexes)
	return args.Bool(0)
}

func (env *Environment) Trace(start time.Time, args ...string) {
	_ = env.Called(start, args)
}

func (env *Environment) Debug(message string) {
	_ = env.Called(message)
}

func (env *Environment) DebugF(format string, a ...any) {
	_ = env.Called(format, a)
}

func (env *Environment) Error(err error) {
	_ = env.Called(err)
}

func (env *Environment) DirIsWritable(path string) bool {
	args := env.Called(path)
	return args.Bool(0)
}

func (env *Environment) CursorPosition() (int, int) {
	args := env.Called()
	return args.Int(0), args.Int(1)
}

func (env *Environment) SystemInfo() (*runtime.SystemInfo, error) {
	args := env.Called()
	return args.Get(0).(*runtime.SystemInfo), args.Error(1)
}

func (env *Environment) Unset(name string) {
	for i := 0; i < len(env.ExpectedCalls); i++ {
		f := env.ExpectedCalls[i]
		if f.Method == name {
			f.Unset()
		}
	}
}
