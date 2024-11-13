package runtime

import (
	"io"
	"io/fs"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/battery"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"

	disk "github.com/shirou/gopsutil/v3/disk"
)

const (
	UNKNOWN = "unknown"
	WINDOWS = "windows"
	DARWIN  = "darwin"
	LINUX   = "linux"
	CMD     = "cmd"

	PRIMARY = "primary"
)

type Environment interface {
	Getenv(key string) string
	Pwd() string
	Home() string
	User() string
	Root() bool
	Host() (string, error)
	GOOS() string
	Shell() string
	Platform() string
	StatusCodes() (int, string)
	HasFiles(pattern string) bool
	HasFilesInDir(dir, pattern string) bool
	HasFolder(folder string) bool
	HasParentFilePath(input string, followSymlinks bool) (fileInfo *FileInfo, err error)
	HasFileInParentDirs(pattern string, depth uint) bool
	ResolveSymlink(input string) (string, error)
	DirMatchesOneOf(dir string, regexes []string) bool
	DirIsWritable(input string) bool
	CommandPath(command string) string
	HasCommand(command string) bool
	FileContent(file string) string
	LsDir(input string) []fs.DirEntry
	RunCommand(command string, args ...string) (string, error)
	RunShellCommand(shell, command string) string
	ExecutionTime() float64
	Flags() *Flags
	BatteryState() (*battery.Info, error)
	QueryWindowTitles(processName, windowTitleRegex string) (string, error)
	WindowsRegistryKeyValue(key string) (*WindowsRegistryValue, error)
	HTTPRequest(url string, body io.Reader, timeout int, requestModifiers ...http.RequestModifier) ([]byte, error)
	IsWsl() bool
	IsWsl2() bool
	IsCygwin() bool
	StackCount() int
	TerminalWidth() (int, error)
	Cache() cache.Cache
	Session() cache.Cache
	Close()
	Logs() string
	InWSLSharedDrive() bool
	ConvertToLinuxPath(input string) string
	ConvertToWindowsPath(input string) string
	Connection(connectionType ConnectionType) (*Connection, error)
	CursorPosition() (row, col int)
	SystemInfo() (*SystemInfo, error)
}

type Flags struct {
	PSWD          string
	PipeStatus    string
	Config        string
	Shell         string
	ShellVersion  string
	PWD           string
	AbsolutePWD   string
	Type          string
	ErrorCode     int
	PromptCount   int
	StackCount    int
	Column        int
	TerminalWidth int
	ExecutionTime float64
	JobCount      int
	IsPrimary     bool
	HasExtra      bool
	Debug         bool
	Plain         bool
	Strict        bool
	Cleared       bool
	NoExitCode    bool
	SaveCache     bool
	Init          bool
	Migrate       bool
	Eval          bool
}

type CommandError struct {
	Err      string
	ExitCode int
}

func (e *CommandError) Error() string {
	return e.Err
}

type FileInfo struct {
	ParentFolder string
	Path         string
	IsDir        bool
}

type WindowsRegistryValueType string

const (
	DWORD  = "DWORD"
	QWORD  = "QWORD"
	BINARY = "BINARY"
	STRING = "STRING"
)

type WindowsRegistryValue struct {
	ValueType WindowsRegistryValueType
	String    string
	DWord     uint64
	QWord     uint64
}

type NotImplemented struct{}

func (n *NotImplemented) Error() string {
	return "not implemented"
}

type ConnectionType string

const (
	ETHERNET  ConnectionType = "ethernet"
	WIFI      ConnectionType = "wifi"
	CELLULAR  ConnectionType = "cellular"
	BLUETOOTH ConnectionType = "bluetooth"
)

type Connection struct {
	Name         string
	Type         ConnectionType
	SSID         string
	TransmitRate uint64
	ReceiveRate  uint64
}

type Memory struct {
	PhysicalTotalMemory     uint64
	PhysicalAvailableMemory uint64
	PhysicalFreeMemory      uint64
	PhysicalPercentUsed     float64
	SwapTotalMemory         uint64
	SwapFreeMemory          uint64
	SwapPercentUsed         float64
}

type SystemInfo struct {
	Disks map[string]disk.IOCountersStat
	Memory
	Load1  float64
	Load5  float64
	Load15 float64
}
