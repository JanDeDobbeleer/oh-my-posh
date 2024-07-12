package runtime

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Azure/go-ansiterm/winterm"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func (term *Terminal) Root() bool {
	defer term.Trace(time.Now())
	var sid *windows.SID

	// Although this looks scary, it is directly copied from the
	// official windows documentation. The Go API for this is a
	// direct wrap around the official C++ API.
	// See https://docs.microsoft.com/en-us/windows/desktop/api/securitybaseapi/nf-securitybaseapi-checktokenmembership
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		term.Error(err)
		return false
	}
	defer func() {
		_ = windows.FreeSid(sid)
	}()

	// This appears to cast a null pointer so I'm not sure why this
	// works, but this guy says it does and it Works for Me™:
	// https://github.com/golang/go/issues/28804#issuecomment-438838144
	token := windows.Token(0)

	member, err := token.IsMember(sid)
	if err != nil {
		term.Error(err)
		return false
	}

	return member
}

func (term *Terminal) Home() string {
	home := os.Getenv("HOME")
	defer func() {
		term.Debug(home)
	}()
	if len(home) > 0 {
		return home
	}
	// fallback to older implemenations on Windows
	home = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}
	return home
}

func (term *Terminal) QueryWindowTitles(processName, windowTitleRegex string) (string, error) {
	defer term.Trace(time.Now(), windowTitleRegex)
	title, err := queryWindowTitles(processName, windowTitleRegex)
	if err != nil {
		term.Error(err)
	}
	return title, err
}

func (term *Terminal) IsWsl() bool {
	defer term.Trace(time.Now())
	return false
}

func (term *Terminal) IsWsl2() bool {
	defer term.Trace(time.Now())
	return false
}

func (term *Terminal) TerminalWidth() (int, error) {
	defer term.Trace(time.Now())

	if term.CmdFlags.TerminalWidth > 0 {
		term.DebugF("terminal width: %d", term.CmdFlags.TerminalWidth)
		return term.CmdFlags.TerminalWidth, nil
	}

	handle, err := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	if err != nil {
		term.Error(err)
		return 0, err
	}

	info, err := winterm.GetConsoleScreenBufferInfo(uintptr(handle))
	if err != nil {
		term.Error(err)
		return 0, err
	}

	term.CmdFlags.TerminalWidth = int(info.Size.X)
	term.DebugF("terminal width: %d", term.CmdFlags.TerminalWidth)
	return term.CmdFlags.TerminalWidth, nil
}

func (term *Terminal) Platform() string {
	return WINDOWS
}

func (term *Terminal) CachePath() string {
	defer term.Trace(time.Now())
	// get LOCALAPPDATA if present
	if cachePath := returnOrBuildCachePath(term.Getenv("LOCALAPPDATA")); len(cachePath) != 0 {
		return cachePath
	}
	return term.Home()
}

// Takes a registry path to a key like
//
//	"HKLM\Software\Microsoft\Windows NT\CurrentVersion\EditionID"
//
// The last part of the path is the key to retrieve.
//
// If the path ends in "\", the "(Default)" key in that path is retrieved.
//
// Returns a variant type if successful; nil and an error if not.
func (term *Terminal) WindowsRegistryKeyValue(path string) (*WindowsRegistryValue, error) {
	term.Trace(time.Now(), path)

	// Format:
	// "HKLM\Software\Microsoft\Windows NT\CurrentVersion\EditionID"
	//   1  |                  2                         |   3
	//
	// Split into:
	//
	// 1. Root key - extract the root HKEY string and turn this into a handle to get started
	// 2. Path - open this path
	// 3. Key - get this key value
	//
	// If 3 is "" (i.e. the path ends with "\"), then get (Default) key.
	//
	rootKey, regPath, found := strings.Cut(path, `\`)
	if !found {
		err := fmt.Errorf("Error, malformed registry path: '%s'", path)
		term.Error(err)
		return nil, err
	}

	var regKey string
	if !strings.HasSuffix(regPath, `\`) {
		regKey = Base(term, regPath)
		if len(regKey) != 0 {
			regPath = strings.TrimSuffix(regPath, `\`+regKey)
		}
	}

	var key registry.Key
	switch rootKey {
	case "HKCR", "HKEY_CLASSES_ROOT":
		key = windows.HKEY_CLASSES_ROOT
	case "HKCC", "HKEY_CURRENT_CONFIG":
		key = windows.HKEY_CURRENT_CONFIG
	case "HKCU", "HKEY_CURRENT_USER":
		key = windows.HKEY_CURRENT_USER
	case "HKLM", "HKEY_LOCAL_MACHINE":
		key = windows.HKEY_LOCAL_MACHINE
	case "HKU", "HKEY_USERS":
		key = windows.HKEY_USERS
	default:
		err := fmt.Errorf("Error, unknown registry key: '%s", rootKey)
		term.Error(err)
		return nil, err
	}

	k, err := registry.OpenKey(key, regPath, registry.READ)
	if err != nil {
		term.Error(err)
		return nil, err
	}

	_, valType, err := k.GetValue(regKey, nil)
	if err != nil {
		term.Error(err)
		return nil, err
	}

	var regValue *WindowsRegistryValue

	switch valType {
	case windows.REG_SZ, windows.REG_EXPAND_SZ:
		value, _, _ := k.GetStringValue(regKey)
		regValue = &WindowsRegistryValue{ValueType: STRING, String: value}
	case windows.REG_DWORD:
		value, _, _ := k.GetIntegerValue(regKey)
		regValue = &WindowsRegistryValue{ValueType: DWORD, DWord: value, String: fmt.Sprintf("0x%08X", value)}
	case windows.REG_QWORD:
		value, _, _ := k.GetIntegerValue(regKey)
		regValue = &WindowsRegistryValue{ValueType: QWORD, QWord: value, String: fmt.Sprintf("0x%016X", value)}
	case windows.REG_BINARY:
		value, _, _ := k.GetBinaryValue(regKey)
		regValue = &WindowsRegistryValue{ValueType: BINARY, String: string(value)}
	}

	if regValue == nil {
		errorLogMsg := fmt.Sprintf("Error, no formatter for type: %d", valType)
		return nil, errors.New(errorLogMsg)
	}

	term.Debug(fmt.Sprintf("%s(%s): %s", regKey, regValue.ValueType, regValue.String))
	return regValue, nil
}

func (term *Terminal) InWSLSharedDrive() bool {
	return false
}

func (term *Terminal) ConvertToWindowsPath(path string) string {
	return strings.ReplaceAll(path, `\`, "/")
}

func (term *Terminal) ConvertToLinuxPath(path string) string {
	return path
}

func (term *Terminal) DirIsWritable(path string) bool {
	defer term.Trace(time.Now())
	return term.isWriteable(path)
}

func (term *Terminal) Connection(connectionType ConnectionType) (*Connection, error) {
	if term.networks == nil {
		networks := term.getConnections()
		if len(networks) == 0 {
			return nil, errors.New("No connections found")
		}
		term.networks = networks
	}
	for _, network := range term.networks {
		if network.Type == connectionType {
			return network, nil
		}
	}
	term.Error(fmt.Errorf("Network type '%s' not found", connectionType))
	return nil, &NotImplemented{}
}

func (term *Terminal) LookPath(command string) (string, error) {
	winAppPath := filepath.Join(term.Getenv("LOCALAPPDATA"), `\Microsoft\WindowsApps\`, command)
	if !strings.HasSuffix(winAppPath, ".exe") {
		winAppPath += ".exe"
	}

	path, err := exec.LookPath(command)
	if err == nil && path != winAppPath {
		return path, nil
	}

	return readWinAppLink(winAppPath)
}
