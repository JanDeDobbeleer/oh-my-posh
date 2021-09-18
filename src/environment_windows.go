//go:build windows

package main

import (
	"os"
	"syscall"
	"time"

	"github.com/Azure/go-ansiterm/winterm"
	"golang.org/x/sys/windows"
)

func (env *environment) isRunningAsRoot() bool {
	defer env.tracer.trace(time.Now(), "isRunningAsRoot")
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
		return false
	}

	return member
}

func (env *environment) homeDir() string {
	home := os.Getenv("HOME")
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

func (env *environment) getWindowTitle(imageName, windowTitleRegex string) (string, error) {
	defer env.tracer.trace(time.Now(), "getWindowTitle", imageName, windowTitleRegex)
	return getWindowTitle(imageName, windowTitleRegex)
}

func (env *environment) isWsl() bool {
	defer env.tracer.trace(time.Now(), "isWsl")
	return false
}

func (env *environment) getTerminalWidth() (int, error) {
	defer env.tracer.trace(time.Now(), "getTerminalWidth")
	handle, err := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	if err != nil {
		return 0, err
	}
	info, err := winterm.GetConsoleScreenBufferInfo(uintptr(handle))
	if err != nil {
		return 0, err
	}
	// return int(float64(info.Size.X) * 0.57), nil
	return int(info.Size.X), nil
}

func (env *environment) getPlatform() string {
	return windowsPlatform
}
