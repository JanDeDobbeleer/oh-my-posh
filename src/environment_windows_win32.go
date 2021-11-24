//go:build windows

package main

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// WindowsProcess is an implementation of Process for Windows.
type WindowsProcess struct {
	pid  int
	ppid int
	exe  string
}

// getImagePid returns the
func getImagePid(imageName string) ([]int, error) {
	processes, err := processes()
	if err != nil {
		return nil, err
	}
	var pids []int
	for i := 0; i < len(processes); i++ {
		if strings.ToLower(processes[i].exe) == imageName {
			pids = append(pids, processes[i].pid)
		}
	}
	return pids, nil
}

// getWindowTitle returns the title of a window linked to a process name
func getWindowTitle(imageName, windowTitleRegex string) (string, error) {
	processPid, err := getImagePid(imageName)
	if err != nil {
		return "", nil
	}

	// is a spotify process running?
	// no: returns an empty string
	if len(processPid) == 0 {
		return "", nil
	}

	// returns the first window of the first pid
	_, windowTitle := GetWindowTitle(processPid[0], windowTitleRegex)

	return windowTitle, nil
}

func newWindowsProcess(e *windows.ProcessEntry32) *WindowsProcess {
	// Find when the string ends for decoding
	end := 0
	for {
		if e.ExeFile[end] == 0 {
			break
		}
		end++
	}

	return &WindowsProcess{
		pid:  int(e.ProcessID),
		ppid: int(e.ParentProcessID),
		exe:  syscall.UTF16ToString(e.ExeFile[:end]),
	}
}

// Processes returns a snapshot of all the processes
// Taken and adapted from https://github.com/mitchellh/go-ps
func processes() ([]WindowsProcess, error) {
	// get process table snapshot
	handle, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, syscall.GetLastError()
	}
	defer func() {
		_ = windows.CloseHandle(handle)
	}()

	// get process infor by looping through the snapshot
	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	err = windows.Process32First(handle, &entry)
	if err != nil {
		return nil, fmt.Errorf("error retrieving process info")
	}

	results := make([]WindowsProcess, 0, 50)
	for {
		results = append(results, *newWindowsProcess(&entry))
		err := windows.Process32Next(handle, &entry)
		if err != nil {
			if err == syscall.ERROR_NO_MORE_FILES {
				break
			}
			return nil, fmt.Errorf("Fail to syscall Process32Next: %v", err)
		}
	}

	return results, nil
}

// win32 specific code

// win32 dll load and function definitions
var (
	user32                       = syscall.NewLazyDLL("user32.dll")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowThreadProcessID = user32.NewProc("GetWindowThreadProcessId")
)

// EnumWindows call EnumWindows from user32 and returns all active windows
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-enumwindows
func EnumWindows(enumFunc, lparam uintptr) (err error) {
	r1, _, e1 := syscall.Syscall(procEnumWindows.Addr(), 2, enumFunc, lparam, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

// GetWindowText returns the title and text of a window from a window handle
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getwindowtextw
func GetWindowText(hwnd syscall.Handle, str *uint16, maxCount int32) (length int32, err error) {
	r0, _, e1 := syscall.Syscall(procGetWindowTextW.Addr(), 3, uintptr(hwnd), uintptr(unsafe.Pointer(str)), uintptr(maxCount))
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

// GetWindowTitle searches for a window attached to the pid
func GetWindowTitle(pid int, windowTitleRegex string) (syscall.Handle, string) {
	var hwnd syscall.Handle
	var title string

	// callback fro EnumWindows
	cb := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {
		var prcsID int
		// get pid
		_, _, _ = procGetWindowThreadProcessID.Call(uintptr(h), uintptr(unsafe.Pointer(&prcsID)))
		// check if pid matches spotify pid
		if prcsID == pid {
			b := make([]uint16, 200)
			_, err := GetWindowText(h, &b[0], int32(len(b)))
			if err != nil {
				// ignore the error
				return 1 // continue enumeration
			}
			title = syscall.UTF16ToString(b)
			if matchString(windowTitleRegex, title) {
				// will cause EnumWindows to return 0 (error)
				// but we don't want to enumerate all windows since we got what we want
				hwnd = h
				return 0
			}
		}

		return 1 // continue enumeration
	})
	// Enumerates all top-level windows on the screen
	// The error is not checked because if EnumWindows is stopped bofere enumerating all windows
	// it returns 0(error occurred) instead of 1(success)
	// In our case, title will equal "" or the title of the window anyway
	_ = EnumWindows(cb, 0)
	return hwnd, title
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
