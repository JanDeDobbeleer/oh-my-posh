//go:build windows

package environment

import (
	"errors"
	"oh-my-posh/regex"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

// win32 specific code

// win32 dll load and function definitions
var (
	user32                       = syscall.NewLazyDLL("user32.dll")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowThreadProcessID = user32.NewProc("GetWindowThreadProcessId")

	psapi              = syscall.NewLazyDLL("psapi.dll")
	getModuleBaseNameA = psapi.NewProc("GetModuleBaseNameA")
)

// enumWindows call enumWindows from user32 and returns all active windows
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-enumwindows
func enumWindows(enumFunc, lparam uintptr) (err error) {
	r1, _, e1 := syscall.SyscallN(procEnumWindows.Addr(), enumFunc, lparam, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

// getWindowText returns the title and text of a window from a window handle
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getwindowtextw
func getWindowText(hwnd syscall.Handle, str *uint16, maxCount int32) (length int32, err error) {
	r0, _, e1 := syscall.SyscallN(procGetWindowTextW.Addr(), uintptr(hwnd), uintptr(unsafe.Pointer(str)), uintptr(maxCount))
	length = int32(r0)
	if length == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func getWindowFileName(handle syscall.Handle) (string, error) {
	var pid int
	// get pid
	_, _, err := procGetWindowThreadProcessID.Call(uintptr(handle), uintptr(unsafe.Pointer(&pid)))
	if err != nil && err.Error() != "The operation completed successfully." {
		return "", errors.New("unable to get window process pid")
	}
	const query = windows.PROCESS_QUERY_INFORMATION | windows.PROCESS_VM_READ
	h, err := windows.OpenProcess(query, false, uint32(pid))
	if err != nil {
		return "", errors.New("unable to open window process")
	}
	buf := [1024]byte{}
	length, _, err := getModuleBaseNameA.Call(uintptr(h), 0, uintptr(unsafe.Pointer(&buf)), 1024)
	if err != nil && err.Error() != "The operation completed successfully." {
		return "", errors.New("unable to get window process name")
	}
	filename := string(buf[:length])
	return strings.ToLower(filename), nil
}

// GetWindowTitle searches for a window attached to the pid
func queryWindowTitles(processName, windowTitleRegex string) (string, error) {
	var title string
	// callback for EnumWindows
	cb := syscall.NewCallback(func(handle syscall.Handle, pointer uintptr) uintptr {
		fileName, err := getWindowFileName(handle)
		if err != nil {
			// ignore the error and continue enumeration
			return 1
		}
		if processName != fileName {
			// ignore the error and continue enumeration
			return 1
		}
		b := make([]uint16, 200)
		_, err = getWindowText(handle, &b[0], int32(len(b)))
		if err != nil {
			// ignore the error and continue enumeration
			return 1
		}
		title = syscall.UTF16ToString(b)
		if regex.MatchString(windowTitleRegex, title) {
			// will cause EnumWindows to return 0 (error)
			// but we don't want to enumerate all windows since we got what we want
			return 0
		}
		return 1 // continue enumeration
	})
	// Enumerates all top-level windows on the screen
	// The error is not checked because if EnumWindows is stopped bofere enumerating all windows
	// it returns 0(error occurred) instead of 1(success)
	// In our case, title will equal "" or the title of the window anyway
	err := enumWindows(cb, 0)
	if len(title) == 0 {
		return "", errors.New("no matching window title found\n" + err.Error())
	}
	return title, nil
}

// Return the windows handles corresponding to the names of the root registry keys.
// A returned value of 0 means there was no match.
func getHKEYHandleFromAbbrString(abbr string) windows.Handle {
	switch abbr {
	case "HKCR", "HKEY_CLASSES_ROOT":
		return windows.HKEY_CLASSES_ROOT
	case "HKCC", "HKEY_CURRENT_CONFIG":
		return windows.HKEY_CURRENT_CONFIG
	case "HKCU", "HKEY_CURRENT_USER":
		return windows.HKEY_CURRENT_USER
	case "HKLM", "HKEY_LOCAL_MACHINE":
		return windows.HKEY_LOCAL_MACHINE
	case "HKU", "HKEY_USERS":
		return windows.HKEY_USERS
	}

	return 0
}

type REPARSE_DATA_BUFFER struct { // nolint: revive
	ReparseTag        uint32
	ReparseDataLength uint16
	Reserved          uint16
	DUMMYUNIONNAME    byte
}

type GenericDataBuffer struct {
	DataBuffer [1]uint8
}

type AppExecLinkReparseBuffer struct {
	Version    uint32
	StringList [1]uint16
}

func (rb *AppExecLinkReparseBuffer) Path() (string, error) {
	UTF16ToStringPosition := func(s []uint16) (string, int) {
		for i, v := range s {
			if v == 0 {
				s = s[0:i]
				return string(utf16.Decode(s)), i
			}
		}
		return "", 0
	}
	stringList := (*[0xffff]uint16)(unsafe.Pointer(&rb.StringList[0]))[0:]
	var link string
	var position int
	for i := 0; i <= 2; i++ {
		link, position = UTF16ToStringPosition(stringList)
		position++
		if position >= len(stringList) {
			return "", errors.New("invalid AppExecLinkReparseBuffer")
		}
		stringList = stringList[position:]
	}
	return link, nil
}

// openSymlink calls CreateFile Windows API with FILE_FLAG_OPEN_REPARSE_POINT
// parameter, so that Windows does not follow symlink, if path is a symlink.
// openSymlink returns opened file handle.
func openSymlink(path string) (syscall.Handle, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	attrs := uint32(syscall.FILE_FLAG_BACKUP_SEMANTICS)
	// Use FILE_FLAG_OPEN_REPARSE_POINT, otherwise CreateFile will follow symlink.
	// See https://docs.microsoft.com/en-us/windows/desktop/FileIO/symbolic-link-effects-on-file-systems-functions#createfile-and-createfiletransacted
	attrs |= syscall.FILE_FLAG_OPEN_REPARSE_POINT
	h, err := syscall.CreateFile(p, 0, 0, nil, syscall.OPEN_EXISTING, attrs, 0)
	if err != nil {
		return 0, err
	}
	return h, nil
}

func readWinAppLink(path string) (string, error) {
	h, err := openSymlink(path)
	if err != nil {
		return "", err
	}
	defer syscall.CloseHandle(h) // nolint: errcheck

	rdbbuf := make([]byte, syscall.MAXIMUM_REPARSE_DATA_BUFFER_SIZE)
	var bytesReturned uint32
	err = syscall.DeviceIoControl(h, syscall.FSCTL_GET_REPARSE_POINT, nil, 0, &rdbbuf[0], uint32(len(rdbbuf)), &bytesReturned, nil)
	if err != nil {
		return "", err
	}

	rdb := (*REPARSE_DATA_BUFFER)(unsafe.Pointer(&rdbbuf[0]))
	rb := (*GenericDataBuffer)(unsafe.Pointer(&rdb.DUMMYUNIONNAME))
	appExecLink := (*AppExecLinkReparseBuffer)(unsafe.Pointer(&rb.DataBuffer))
	if appExecLink.Version != 3 {
		return " ", errors.New("unknown AppExecLink version")
	}
	return appExecLink.Path()
}
