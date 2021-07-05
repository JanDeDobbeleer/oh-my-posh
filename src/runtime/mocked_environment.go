// +build !release

package runtime

import (
	"oh-my-posh/engine"

	"github.com/distatus/battery"
	"github.com/stretchr/testify/mock"
)

type MockedEnvironment struct {
	mock.Mock
}

func (env *MockedEnvironment) Getenv(key string) string {
	args := env.Called(key)
	return args.String(0)
}

func (env *MockedEnvironment) Getcwd() string {
	args := env.Called(nil)
	return args.String(0)
}

func (env *MockedEnvironment) HomeDir() string {
	args := env.Called(nil)
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

func (env *MockedEnvironment) GetFileContent(file string) string {
	args := env.Called(file)
	return args.String(0)
}

func (env *MockedEnvironment) GetPathSeperator() string {
	args := env.Called(nil)
	return args.String(0)
}

func (env *MockedEnvironment) GetCurrentUser() string {
	args := env.Called(nil)
	return args.String(0)
}

func (env *MockedEnvironment) GetHostName() (string, error) {
	args := env.Called(nil)
	return args.String(0), args.Error(1)
}

func (env *MockedEnvironment) GetRuntimeGOOS() string {
	args := env.Called(nil)
	return args.String(0)
}

func (env *MockedEnvironment) GetPlatform() string {
	args := env.Called(nil)
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

func (env *MockedEnvironment) LastErrorCode() int {
	args := env.Called(nil)
	return args.Int(0)
}

func (env *MockedEnvironment) ExecutionTime() float64 {
	args := env.Called(nil)
	return float64(args.Int(0))
}

func (env *MockedEnvironment) IsRunningAsRoot() bool {
	args := env.Called(nil)
	return args.Bool(0)
}

func (env *MockedEnvironment) GetArgs() *engine.Args {
	arguments := env.Called(nil)
	return arguments.Get(0).(*engine.Args)
}

func (env *MockedEnvironment) GetBatteryInfo() ([]*battery.Battery, error) {
	args := env.Called(nil)
	return args.Get(0).([]*battery.Battery), args.Error(1)
}

func (env *MockedEnvironment) GetShellName() string {
	args := env.Called(nil)
	return args.String(0)
}

func (env *MockedEnvironment) GetWindowTitle(imageName, windowTitleRegex string) (string, error) {
	args := env.Called(imageName)
	return args.String(0), args.Error(1)
}

func (env *MockedEnvironment) DoGet(url string) ([]byte, error) {
	args := env.Called(url)
	return args.Get(0).([]byte), args.Error(1)
}

func (env *MockedEnvironment) HasParentFilePath(path string) (*File, error) {
	args := env.Called(path)
	return args.Get(0).(*File), args.Error(1)
}

func (env *MockedEnvironment) StackCount() int {
	args := env.Called(nil)
	return args.Int(0)
}

func (env *MockedEnvironment) IsWsl() bool {
	return false
}

func (env *MockedEnvironment) GetTerminalWidth() (int, error) {
	args := env.Called(nil)
	return args.Int(0), args.Error(1)
}
