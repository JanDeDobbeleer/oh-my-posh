//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/Azure/go-ansiterm/winterm"
	"golang.org/x/sys/windows"
)

func (env *environment) isRunningAsRoot() bool {
	defer env.trace(time.Now(), "isRunningAsRoot")
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
		env.log(Error, "isRunningAsRoot", err.Error())
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
		env.log(Error, "isRunningAsRoot", err.Error())
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
	defer env.trace(time.Now(), "getWindowTitle", imageName, windowTitleRegex)
	return getWindowTitle(imageName, windowTitleRegex)
}

func (env *environment) isWsl() bool {
	defer env.trace(time.Now(), "isWsl")
	return false
}

func (env *environment) getTerminalWidth() (int, error) {
	defer env.trace(time.Now(), "getTerminalWidth")
	handle, err := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	if err != nil {
		env.log(Error, "getTerminalWidth", err.Error())
		return 0, err
	}
	info, err := winterm.GetConsoleScreenBufferInfo(uintptr(handle))
	if err != nil {
		env.log(Error, "getTerminalWidth", err.Error())
		return 0, err
	}
	// return int(float64(info.Size.X) * 0.57), nil
	return int(info.Size.X), nil
}

func (env *environment) getPlatform() string {
	return windowsPlatform
}

func (env *environment) getCachePath() string {
	defer env.trace(time.Now(), "getCachePath")
	// get LOCALAPPDATA if present
	if cachePath := returnOrBuildCachePath(env.getenv("LOCALAPPDATA")); len(cachePath) != 0 {
		return cachePath
	}
	return env.homeDir()
}

//
// Takes a registry path like "HKLM\Software\Microsoft\Windows NT\CurrentVersion" and a key under that path like "CurrentVersion" (or "" if the (Default) key is required).
// Returns a bool and string:
//
//   true and the retrieved value formatted into a string if successful.
//   false and the string will be the error
//
func (env *environment) getWindowsRegistryKeyValue(regPath, regKey string) (string, error) {
	env.trace(time.Now(), "getWindowsRegistryKeyValue", regPath, regKey)

	// Extract root HK value and turn it into a windows.Handle to open the key.
	regPathParts := strings.SplitN(regPath, "\\", 2)

	regRootHKeyHandle := getHKEYHandleFromAbbrString(regPathParts[0])
	if regRootHKeyHandle == 0 {
		errorLogMsg := fmt.Sprintf("Error, Supplied root HKEY value not valid: '%s'", regPathParts[0])
		env.log(Error, "getWindowsRegistryKeyValue", errorLogMsg)
		return "", errors.New(errorLogMsg)
	}

	// Second part of split is registry path after HK part - needs to be UTF16 to pass to the windows. API
	regPathUTF16, regPathUTF16ConversionErr := windows.UTF16FromString(regPathParts[1])
	if regPathUTF16ConversionErr != nil {
		errorLogMsg := fmt.Sprintf("Error, Could not convert supplied path '%s' to UTF16, error: '%s'", regPathParts[1], regPathUTF16ConversionErr)
		env.log(Error, "getWindowsRegistryKeyValue", errorLogMsg)
		return "", errors.New(errorLogMsg)
	}

	// Ok - open it..
	var hKeyHandle windows.Handle
	regOpenErr := windows.RegOpenKeyEx(regRootHKeyHandle, &regPathUTF16[0], 0, windows.KEY_READ, &hKeyHandle)
	if regOpenErr != nil {
		errorLogMsg := fmt.Sprintf("Error RegOpenKeyEx opening registry path to '%s', error: '%s'", regPath, regOpenErr)
		env.log(Error, "getWindowsRegistryKeyValue", errorLogMsg)
		return "", errors.New(errorLogMsg)
	}
	// Success - from here on out, when returning make sure to close that reg key with a deferred call to close:
	defer func() {
		err := windows.RegCloseKey(hKeyHandle)
		if err != nil {
			env.log(Error, "getWindowsRegistryKeyValue", fmt.Sprintf("Error closing registry key: %s", err))
		}
	}()

	// Again - need UTF16 of the key for the API:
	regKeyUTF16, regKeyUTF16ConversionErr := windows.UTF16FromString(regKey)
	if regKeyUTF16ConversionErr != nil {
		errorLogMsg := fmt.Sprintf("Error, could not convert supplied key '%s' to UTF16, error: '%s'", regKey, regKeyUTF16ConversionErr)
		env.log(Error, "getWindowsRegistryKeyValue", errorLogMsg)
		return "", errors.New(errorLogMsg)
	}

	// Two stage way to get the key value - query once to get size - then allocate and query again to fill it.
	var keyBufType uint32
	var keyBufSize uint32

	regQueryErr := windows.RegQueryValueEx(hKeyHandle, &regKeyUTF16[0], nil, &keyBufType, nil, &keyBufSize)
	if regQueryErr != nil {
		errorLogMsg := fmt.Sprintf("Error calling RegQueryValueEx to retrieve key data size with error '%s'", regQueryErr)
		env.log(Error, "getWindowsRegistryKeyValue", errorLogMsg)
		return "", errors.New(errorLogMsg)
	}

	// Alloc and fill...
	var keyBuf = make([]byte, keyBufSize)

	regQueryErr = windows.RegQueryValueEx(hKeyHandle, &regKeyUTF16[0], nil, &keyBufType, &keyBuf[0], &keyBufSize)
	if regQueryErr != nil {
		errorLogMsg := fmt.Sprintf("Error calling RegQueryValueEx to retrieve key data with error '%s'", regQueryErr)
		env.log(Error, "getWindowsRegistryKeyValue", errorLogMsg)
		return "", errors.New(errorLogMsg)
	}

	// Format result into a string, depending on type.  (future refactor - move this out into it's own function)
	switch keyBufType {
	case windows.REG_SZ:
		var uint16p *uint16
		uint16p = (*uint16)(unsafe.Pointer(&keyBuf[0])) // nasty casty

		valueString := windows.UTF16PtrToString(uint16p)
		env.log(Debug, "getWindowsRegistryKeyValue", fmt.Sprintf("success, string: %s", valueString))
		return valueString, nil
	case windows.REG_DWORD:
		var uint32p *uint32
		uint32p = (*uint32)(unsafe.Pointer(&keyBuf[0])) // more casting goodness

		valueString := fmt.Sprintf("0x%08X", *uint32p)
		env.log(Debug, "getWindowsRegistryKeyValue", fmt.Sprintf("success, DWORD, formatted as string: %s", valueString))
		return valueString, nil
	default:
		errorLogMsg := fmt.Sprintf("Error, no formatter for REG_? type:%d, data size:%d bytes", keyBufType, keyBufSize)
		return "", errors.New(errorLogMsg)
	}
}
