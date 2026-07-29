package runtime

import (
	"encoding/json"
	"io"
	"io/fs"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/battery"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"

	disk "github.com/shirou/gopsutil/v4/disk"
)

const (
	UNKNOWN = "unknown"
	WINDOWS = "windows"
	DARWIN  = "darwin"
	LINUX   = "linux"
	FREEBSD = "freebsd"
	CMD     = "cmd"
	ANDROID = "android"

	PRIMARY = "primary"
)

// Player-agnostic: the OS media-transport layer (e.g. Windows SMTC) can surface any app that publishes a session.
type MediaInfo struct {
	// Status is the lowercased playback status: playing/paused/stopped/closed/opened/changing.
	Status      string
	Title       string
	Artist      string
	Album       string
	TrackNumber int
}

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
	RunCommandWithEnv(command string, envs []string, args ...string) (string, error)
	RunShellCommand(shell, command string) string
	ExecutionTime() float64
	Flags() *Flags
	BatteryState() (*battery.Info, error)
	QueryWindowTitles(processName, windowTitleRegex string) (string, error)
	QueryMediaPlayer(player string) (*MediaInfo, error)
	WindowsRegistryKeyValue(key string) (*WindowsRegistryValue, error)
	HTTPRequest(url string, body io.Reader, timeout int, requestModifiers ...http.RequestModifier) ([]byte, error)
	IsWsl() bool
	IsWsl2() bool
	IsCygwin() bool
	StackCount() int
	TerminalWidth() (int, error)
	Logs() string
	InWSLSharedDrive() bool
	ConvertToLinuxPath(input string) string
	ConvertToWindowsPath(input string) string
	Connection(connectionType ConnectionType) (*Connection, error)
	CursorPosition() (row, col int)
	SystemInfo() (*SystemInfo, error)
}

type Flags struct {
	SegmentData   map[string]json.RawMessage
	Type          string
	PipeStatus    string
	ConfigPath    string
	PSWD          string
	Shell         string
	ShellVersion  string
	PWD           string
	AbsolutePWD   string
	EnvData       json.RawMessage
	ExecutionTime float64
	PromptCount   int
	Column        int
	TerminalWidth int
	ErrorCode     int
	StackCount    int
	ConfigHash    uint64
	JobCount      int
	Cleared       bool
	Strict        bool
	Debug         bool
	HasExtra      bool
	NoExitCode    bool
	Init          bool
	Migrate       bool
	Eval          bool
	Escape        bool
	IsPrimary     bool
	Plain         bool
	Force         bool
	Streaming     bool
	Interrupted   bool
	// DataOnly cuts this environment off from the machine: every method that
	// would read a file, list a directory, resolve a symlink, run a command,
	// make a request or read an OS variable answers empty or errDataOnly (see
	// Terminal's own methods). A segment still executes and still decides for
	// itself whether it is enabled - it just finds nothing when it reaches
	// out, so anything that depends on a real machine reports itself absent
	// while anything derivable from the recorded data renders normally.
	//
	// Gating the environment rather than the segments is what makes the
	// guarantee hold for code nobody audited: git's StashCount reads
	// logs/refs/stash from a method the template calls, long after the
	// recorded data has been restored, and no rule about recorded keys would
	// have caught it.
	//
	// Off for an ordinary prompt render, where reaching the machine is the
	// entire point; on for a caller rendering a config it was handed rather
	// than one the user owns, such as the studio.
	DataOnly bool
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
