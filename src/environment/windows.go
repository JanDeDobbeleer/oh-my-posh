//go:build windows

package environment

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"

	"github.com/Azure/go-ansiterm/winterm"
	"golang.org/x/sys/windows"
)

func (env *ShellEnvironment) Root() bool {
	defer env.Trace(time.Now(), "Root")
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
		env.Log(Error, "Root", err.Error())
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
		env.Log(Error, "Root", err.Error())
		return false
	}

	return member
}

func (env *ShellEnvironment) Home() string {
	home := os.Getenv("HOME")
	defer func() {
		env.Log(Debug, "Home", home)
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

func (env *ShellEnvironment) QueryWindowTitles(processName, windowTitleRegex string) (string, error) {
	defer env.Trace(time.Now(), "WindowTitle", windowTitleRegex)
	title, err := queryWindowTitles(processName, windowTitleRegex)
	if err != nil {
		env.Log(Error, "QueryWindowTitles", err.Error())
	}
	return title, err
}

func (env *ShellEnvironment) IsWsl() bool {
	defer env.Trace(time.Now(), "IsWsl")
	return false
}

func (env *ShellEnvironment) IsWsl2() bool {
	defer env.Trace(time.Now(), "IsWsl2")
	return false
}

func (env *ShellEnvironment) TerminalWidth() (int, error) {
	defer env.Trace(time.Now(), "TerminalWidth")
	if env.CmdFlags.TerminalWidth != 0 {
		return env.CmdFlags.TerminalWidth, nil
	}
	handle, err := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	if err != nil {
		env.Log(Error, "TerminalWidth", err.Error())
		return 0, err
	}
	info, err := winterm.GetConsoleScreenBufferInfo(uintptr(handle))
	if err != nil {
		env.Log(Error, "TerminalWidth", err.Error())
		return 0, err
	}
	// return int(float64(info.Size.X) * 0.57), nil
	return int(info.Size.X), nil
}

func (env *ShellEnvironment) Platform() string {
	return WindowsPlatform
}

func (env *ShellEnvironment) CachePath() string {
	defer env.Trace(time.Now(), "CachePath")
	// get LOCALAPPDATA if present
	if cachePath := returnOrBuildCachePath(env.Getenv("LOCALAPPDATA")); len(cachePath) != 0 {
		return cachePath
	}
	return env.Home()
}

func (env *ShellEnvironment) LookWinAppPath(file string) (string, error) {
	winAppPath := filepath.Join(env.Getenv("LOCALAPPDATA"), `\Microsoft\WindowsApps\`)
	command := file + ".exe"
	isWinStoreApp := func() bool {
		return env.HasFilesInDir(winAppPath, command)
	}
	if isWinStoreApp() {
		commandFile := filepath.Join(winAppPath, command)
		return readWinAppLink(commandFile)
	}
	return "", errors.New("no Windows Store App")
}

//
// Takes a registry path to a key like
//		"HKLM\Software\Microsoft\Windows NT\CurrentVersion\EditionID"
//
// The last part of the path is the key to retrieve.
//
// If the path ends in "\", the "(Default)" key in that path is retrieved.
//
// Returns a variant type if successful; nil and an error if not.
//
func (env *ShellEnvironment) WindowsRegistryKeyValue(path string) (*WindowsRegistryValue, error) {
	env.Trace(time.Now(), "WindowsRegistryKeyValue", path)

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
	regPathParts := strings.SplitN(path, "\\", 2)

	if len(regPathParts) < 2 {
		errorLogMsg := fmt.Sprintf("Error, malformed registry path: '%s'", path)
		env.Log(Error, "WindowsRegistryKeyValue", errorLogMsg)
		return nil, errors.New(errorLogMsg)
	}

	regRootHKeyHandle := getHKEYHandleFromAbbrString(regPathParts[0])
	if regRootHKeyHandle == 0 {
		errorLogMsg := fmt.Sprintf("Error, Supplied root HKEY value not valid: '%s'", regPathParts[0])
		env.Log(Error, "WindowsRegistryKeyValue", errorLogMsg)
		return nil, errors.New(errorLogMsg)
	}

	// Strip key off the end.
	lastSlash := strings.LastIndex(regPathParts[1], "\\")

	if lastSlash < 0 {
		errorLogMsg := fmt.Sprintf("Error, malformed registry path: '%s'", path)
		env.Log(Error, "WindowsRegistryKeyValue", errorLogMsg)
		return nil, errors.New(errorLogMsg)
	}

	regKey := regPathParts[1][lastSlash+1:]
	regPath := regPathParts[1][0:lastSlash]

	// Just for debug log display.
	regKeyLogged := regKey
	if len(regKeyLogged) == 0 {
		regKeyLogged = "(Default)"
	}
	env.Log(Debug, "WindowsRegistryKeyValue", fmt.Sprintf("WindowsRegistryKeyValue: root:\"%s\", path:\"%s\", key:\"%s\"", regPathParts[0], regPath, regKeyLogged))

	// Second part of split is registry path after HK part - needs to be UTF16 to pass to the windows. API
	regPathUTF16, err := windows.UTF16FromString(regPath)
	if err != nil {
		errorLogMsg := fmt.Sprintf("Error, Could not convert supplied path '%s' to UTF16, error: '%s'", regPath, err)
		env.Log(Error, "WindowsRegistryKeyValue", errorLogMsg)
		return nil, errors.New(errorLogMsg)
	}

	// Ok - open it..
	var hKeyHandle windows.Handle
	regOpenErr := windows.RegOpenKeyEx(regRootHKeyHandle, &regPathUTF16[0], 0, windows.KEY_READ, &hKeyHandle)
	if regOpenErr != nil {
		errorLogMsg := fmt.Sprintf("Error RegOpenKeyEx opening registry path to '%s', error: '%s'", regPath, regOpenErr)
		env.Log(Error, "WindowsRegistryKeyValue", errorLogMsg)
		return nil, errors.New(errorLogMsg)
	}
	// Success - from here on out, when returning make sure to close that reg key with a deferred call to close:
	defer func() {
		err := windows.RegCloseKey(hKeyHandle)
		if err != nil {
			env.Log(Error, "WindowsRegistryKeyValue", fmt.Sprintf("Error closing registry key: %s", err))
		}
	}()

	// Again - need UTF16 of the key for the API:
	regKeyUTF16, err := windows.UTF16FromString(regKey)
	if err != nil {
		errorLogMsg := fmt.Sprintf("Error, could not convert supplied key '%s' to UTF16, error: '%s'", regKey, err)
		env.Log(Error, "WindowsRegistryKeyValue", errorLogMsg)
		return nil, errors.New(errorLogMsg)
	}

	// Two stage way to get the key value - query once to get size - then allocate and query again to fill it.
	var keyBufType uint32
	var keyBufSize uint32

	regQueryErr := windows.RegQueryValueEx(hKeyHandle, &regKeyUTF16[0], nil, &keyBufType, nil, &keyBufSize)
	if regQueryErr != nil {
		errorLogMsg := fmt.Sprintf("Error calling RegQueryValueEx to retrieve key data size with error '%s'", regQueryErr)
		env.Log(Error, "WindowsRegistryKeyValue", errorLogMsg)
		return nil, errors.New(errorLogMsg)
	}

	// Alloc and fill...
	var keyBuf = make([]byte, keyBufSize)

	regQueryErr = windows.RegQueryValueEx(hKeyHandle, &regKeyUTF16[0], nil, &keyBufType, &keyBuf[0], &keyBufSize)
	if regQueryErr != nil {
		errorLogMsg := fmt.Sprintf("Error calling RegQueryValueEx to retrieve key data with error '%s'", regQueryErr)
		env.Log(Error, "WindowsRegistryKeyValue", errorLogMsg)
		return nil, errors.New(errorLogMsg)
	}

	// Format result into a string, depending on type.  (future refactor - move this out into it's own function)
	switch keyBufType {
	case windows.REG_SZ:
		var uint16p *uint16
		uint16p = (*uint16)(unsafe.Pointer(&keyBuf[0])) // nasty casty

		valueString := windows.UTF16PtrToString(uint16p)
		env.Log(Debug, "WindowsRegistryKeyValue", fmt.Sprintf("success, string: %s", valueString))

		return &WindowsRegistryValue{ValueType: RegString, Str: valueString}, nil
	case windows.REG_DWORD:
		var uint32p *uint32
		uint32p = (*uint32)(unsafe.Pointer(&keyBuf[0])) // more casting goodness

		env.Log(Debug, "WindowsRegistryKeyValue", fmt.Sprintf("success, DWORD, 0x%08X", *uint32p))
		return &WindowsRegistryValue{ValueType: RegDword, Dword: *uint32p}, nil
	case windows.REG_QWORD:
		var uint64p *uint64
		uint64p = (*uint64)(unsafe.Pointer(&keyBuf[0])) // more casting goodness

		env.Log(Debug, "WindowsRegistryKeyValue", fmt.Sprintf("success, QWORD, 0x%016X", *uint64p))
		return &WindowsRegistryValue{ValueType: RegQword, Qword: *uint64p}, nil
	default:
		errorLogMsg := fmt.Sprintf("Error, no formatter for REG_? type:%d, data size:%d bytes", keyBufType, keyBufSize)
		return nil, errors.New(errorLogMsg)
	}
}

func (env *ShellEnvironment) InWSLSharedDrive() bool {
	return false
}

func (env *ShellEnvironment) ConvertToWindowsPath(path string) string {
	return path
}

func (env *ShellEnvironment) ConvertToLinuxPath(path string) string {
	return path
}

var (
	hapi                = syscall.NewLazyDLL("wlanapi.dll")
	hWlanOpenHandle     = hapi.NewProc("WlanOpenHandle")
	hWlanCloseHandle    = hapi.NewProc("WlanCloseHandle")
	hWlanEnumInterfaces = hapi.NewProc("WlanEnumInterfaces")
	hWlanQueryInterface = hapi.NewProc("WlanQueryInterface")
)

const (
	FHSS   WifiType = "FHSS"
	DSSS   WifiType = "DSSS"
	IR     WifiType = "IR"
	A      WifiType = "802.11a"
	HRDSSS WifiType = "HRDSSS"
	G      WifiType = "802.11g"
	N      WifiType = "802.11n"
	AC     WifiType = "802.11ac"

	Infrastructure WifiType = "Infrastructure"
	Independent    WifiType = "Independent"
	Any            WifiType = "Any"

	OpenSystem WifiType = "802.11 Open System"
	SharedKey  WifiType = "802.11 Shared Key"
	WPA        WifiType = "WPA"
	WPAPSK     WifiType = "WPA PSK"
	WPANone    WifiType = "WPA NONE"
	WPA2       WifiType = "WPA2"
	WPA2PSK    WifiType = "WPA2 PSK"
	Disabled   WifiType = "disabled"

	None   WifiType = "None"
	WEP40  WifiType = "WEP40"
	TKIP   WifiType = "TKIP"
	CCMP   WifiType = "CCMP"
	WEP104 WifiType = "WEP104"
	WEP    WifiType = "WEP"
)

func (env *ShellEnvironment) WifiNetwork() (*WifiInfo, error) {
	env.Trace(time.Now(), "WifiNetwork")
	// Open handle
	var pdwNegotiatedVersion uint32
	var phClientHandle uint32
	e, _, err := hWlanOpenHandle.Call(uintptr(uint32(2)), uintptr(unsafe.Pointer(nil)), uintptr(unsafe.Pointer(&pdwNegotiatedVersion)), uintptr(unsafe.Pointer(&phClientHandle)))
	if e != 0 {
		return nil, err
	}

	// defer closing handle
	defer func() {
		_, _, _ = hWlanCloseHandle.Call(uintptr(phClientHandle), uintptr(unsafe.Pointer(nil)))
	}()

	// list interfaces
	var interfaceList *WLAN_INTERFACE_INFO_LIST
	e, _, err = hWlanEnumInterfaces.Call(uintptr(phClientHandle), uintptr(unsafe.Pointer(nil)), uintptr(unsafe.Pointer(&interfaceList)))
	if e != 0 {
		return nil, err
	}

	// use first interface that is connected
	numberOfInterfaces := int(interfaceList.dwNumberOfItems)
	infoSize := unsafe.Sizeof(interfaceList.InterfaceInfo[0])
	for i := 0; i < numberOfInterfaces; i++ {
		network := (*WLAN_INTERFACE_INFO)(unsafe.Pointer(uintptr(unsafe.Pointer(&interfaceList.InterfaceInfo[0])) + uintptr(i)*infoSize))
		if network.isState != 1 {
			continue
		}
		return env.parseNetworkInterface(network, phClientHandle)
	}
	return nil, errors.New("Not connected")
}

func (env *ShellEnvironment) parseNetworkInterface(network *WLAN_INTERFACE_INFO, clientHandle uint32) (*WifiInfo, error) {
	info := WifiInfo{}
	info.Interface = strings.TrimRight(string(utf16.Decode(network.strInterfaceDescription[:])), "\x00")

	// Query wifi connection state
	var dataSize uint16
	var wlanAttr *WLAN_CONNECTION_ATTRIBUTES
	e, _, err := hWlanQueryInterface.Call(uintptr(clientHandle),
		uintptr(unsafe.Pointer(&network.InterfaceGuid)),
		uintptr(7), // wlan_intf_opcode_current_connection
		uintptr(unsafe.Pointer(nil)),
		uintptr(unsafe.Pointer(&dataSize)),
		uintptr(unsafe.Pointer(&wlanAttr)),
		uintptr(unsafe.Pointer(nil)))
	if e != 0 {
		env.Log(Error, "parseNetworkInterface", "wlan_intf_opcode_current_connection error")
		return &info, err
	}

	// SSID
	ssid := wlanAttr.wlanAssociationAttributes.dot11Ssid
	if ssid.uSSIDLength > 0 {
		info.SSID = string(ssid.ucSSID[0:ssid.uSSIDLength])
	}

	// see https://docs.microsoft.com/en-us/windows/win32/nativewifi/dot11-phy-type
	switch wlanAttr.wlanAssociationAttributes.dot11PhyType {
	case 1:
		info.PhysType = FHSS
	case 2:
		info.PhysType = DSSS
	case 3:
		info.PhysType = IR
	case 4:
		info.PhysType = A
	case 5:
		info.PhysType = HRDSSS
	case 6:
		info.PhysType = G
	case 7:
		info.PhysType = N
	case 8:
		info.PhysType = AC
	default:
		info.PhysType = Unknown
	}

	// see https://docs.microsoft.com/en-us/windows/win32/nativewifi/dot11-bss-type
	switch wlanAttr.wlanAssociationAttributes.dot11BssType {
	case 1:
		info.RadioType = Infrastructure
	case 2:
		info.RadioType = Independent
	default:
		info.RadioType = Any
	}

	info.Signal = int(wlanAttr.wlanAssociationAttributes.wlanSignalQuality)
	info.TransmitRate = int(wlanAttr.wlanAssociationAttributes.ulTxRate) / 1024
	info.ReceiveRate = int(wlanAttr.wlanAssociationAttributes.ulRxRate) / 1024

	// Query wifi channel
	dataSize = 0
	var channel *uint32
	e, _, err = hWlanQueryInterface.Call(uintptr(clientHandle),
		uintptr(unsafe.Pointer(&network.InterfaceGuid)),
		uintptr(8), // wlan_intf_opcode_channel_number
		uintptr(unsafe.Pointer(nil)),
		uintptr(unsafe.Pointer(&dataSize)),
		uintptr(unsafe.Pointer(&channel)),
		uintptr(unsafe.Pointer(nil)))
	if e != 0 {
		env.Log(Error, "parseNetworkInterface", "wlan_intf_opcode_channel_number error")
		return &info, err
	}
	info.Channel = int(*channel)

	if wlanAttr.wlanSecurityAttributes.bSecurityEnabled <= 0 {
		info.Authentication = Disabled
		return &info, nil
	}

	// see https://docs.microsoft.com/en-us/windows/win32/nativewifi/dot11-auth-algorithm
	switch wlanAttr.wlanSecurityAttributes.dot11AuthAlgorithm {
	case 1:
		info.Authentication = OpenSystem
	case 2:
		info.Authentication = SharedKey
	case 3:
		info.Authentication = WPA
	case 4:
		info.Authentication = WPAPSK
	case 5:
		info.Authentication = WPANone
	case 6:
		info.Authentication = WPA2
	case 7:
		info.Authentication = WPA2PSK
	default:
		info.Authentication = Unknown
	}

	// see https://docs.microsoft.com/en-us/windows/win32/nativewifi/dot11-cipher-algorithm
	switch wlanAttr.wlanSecurityAttributes.dot11CipherAlgorithm {
	case 0:
		info.Cipher = None
	case 0x1:
		info.Cipher = WEP40
	case 0x2:
		info.Cipher = TKIP
	case 0x4:
		info.Cipher = CCMP
	case 0x5:
		info.Cipher = WEP104
	case 0x100:
		info.Cipher = WPA
	case 0x101:
		info.Cipher = WEP
	default:
		info.Cipher = Unknown
	}

	return &info, nil
}

type WLAN_INTERFACE_INFO_LIST struct { // nolint: revive
	dwNumberOfItems uint32
	dwIndex         uint32 // nolint: unused
	InterfaceInfo   [1]WLAN_INTERFACE_INFO
}

type WLAN_INTERFACE_INFO struct { // nolint: revive
	InterfaceGuid           syscall.GUID // nolint: revive
	strInterfaceDescription [256]uint16
	isState                 uint32
}

type WLAN_CONNECTION_ATTRIBUTES struct { // nolint: revive
	isState                   uint32      // nolint: unused
	wlanConnectionMode        uint32      // nolint: unused
	strProfileName            [256]uint16 // nolint: unused
	wlanAssociationAttributes WLAN_ASSOCIATION_ATTRIBUTES
	wlanSecurityAttributes    WLAN_SECURITY_ATTRIBUTES
}

type WLAN_ASSOCIATION_ATTRIBUTES struct { // nolint: revive
	dot11Ssid         DOT11_SSID
	dot11BssType      uint32
	dot11Bssid        [6]uint8 // nolint: unused
	dot11PhyType      uint32
	uDot11PhyIndex    uint32 // nolint: unused
	wlanSignalQuality uint32
	ulRxRate          uint32
	ulTxRate          uint32
}

type WLAN_SECURITY_ATTRIBUTES struct { // nolint: revive
	bSecurityEnabled     uint32
	bOneXEnabled         uint32 // nolint: unused
	dot11AuthAlgorithm   uint32
	dot11CipherAlgorithm uint32
}

type DOT11_SSID struct { // nolint: revive
	uSSIDLength uint32
	ucSSID      [32]uint8
}
