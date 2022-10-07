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

	iphlpapi     = syscall.NewLazyDLL("iphlpapi.dll")
	hGetIfTable2 = iphlpapi.NewProc("GetIfTable2")
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
	_, _, _ = procGetWindowThreadProcessID.Call(uintptr(handle), uintptr(unsafe.Pointer(&pid)))
	const query = windows.PROCESS_QUERY_INFORMATION | windows.PROCESS_VM_READ
	h, err := windows.OpenProcess(query, false, uint32(pid))
	if err != nil {
		return "", errors.New("unable to open window process")
	}
	buf := [1024]byte{}
	length, _, _ := getModuleBaseNameA.Call(uintptr(h), 0, uintptr(unsafe.Pointer(&buf)), 1024)
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
		var message string
		if err != nil {
			message = err.Error()
		}
		return "", errors.New("no matching window title found\n" + message)
	}
	return title, nil
}

type REPARSE_DATA_BUFFER struct { //nolint: revive
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
	defer syscall.CloseHandle(h) //nolint: errcheck

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
		return "", errors.New("unknown AppExecLink version")
	}
	return appExecLink.Path()
}

// networks

func (env *ShellEnvironment) getConnections() []*Connection {
	var pIFTable2 *MIN_IF_TABLE2
	_, _, _ = hGetIfTable2.Call(uintptr(unsafe.Pointer(&pIFTable2)))

	networks := make([]*Connection, 0)

	for i := 0; i < int(pIFTable2.NumEntries); i++ {
		networkInterface := pIFTable2.Table[i]
		alias := strings.TrimRight(syscall.UTF16ToString(networkInterface.Alias[:]), "\x00")
		description := strings.TrimRight(syscall.UTF16ToString(networkInterface.Description[:]), "\x00")

		if networkInterface.OperStatus != 1 || // not connected or functional
			!networkInterface.InterfaceAndOperStatusFlags.HardwareInterface || // rule out software interfaces
			strings.HasPrefix(alias, "Local Area Connection") || // not relevant
			strings.Index(alias, "-") >= 3 { // rule out parts of Ethernet filter interfaces
			// e.g. : "Ethernet-WFP Native MAC Layer LightWeight Filter-0000"
			continue
		}

		var connectionType ConnectionType
		var ssid string
		switch networkInterface.Type {
		case 6:
			connectionType = ETHERNET
		case 71:
			connectionType = WIFI
			ssid = env.getWiFiSSID(networkInterface.InterfaceGUID)
		case 237, 234, 244:
			connectionType = CELLULAR
		}

		if networkInterface.PhysicalMediumType == 10 {
			connectionType = BLUETOOTH
		}

		// skip connections which aren't relevant
		if len(connectionType) == 0 {
			continue
		}

		network := &Connection{
			Type:         connectionType,
			Name:         description, // we want a relatable name, alias isn't that
			TransmitRate: networkInterface.TransmitLinkSpeed,
			ReceiveRate:  networkInterface.ReceiveLinkSpeed,
			SSID:         ssid,
		}

		networks = append(networks, network)
	}
	return networks
}

type MIN_IF_TABLE2 struct { //nolint: revive
	NumEntries uint64
	Table      [256]MIB_IF_ROW2
}

const (
	IF_MAX_STRING_SIZE         uint64 = 256 //nolint: revive
	IF_MAX_PHYS_ADDRESS_LENGTH uint64 = 32  //nolint: revive
)

type MIB_IF_ROW2 struct { //nolint: revive
	InterfaceLuid            uint64
	InterfaceIndex           uint32
	InterfaceGUID            windows.GUID
	Alias                    [IF_MAX_STRING_SIZE + 1]uint16
	Description              [IF_MAX_STRING_SIZE + 1]uint16
	PhysicalAddressLength    uint32
	PhysicalAddress          [IF_MAX_PHYS_ADDRESS_LENGTH]uint8
	PermanentPhysicalAddress [IF_MAX_PHYS_ADDRESS_LENGTH]uint8

	Mtu                uint32
	Type               uint32
	TunnelType         uint32
	MediaType          uint32
	PhysicalMediumType uint32
	AccessType         uint32
	DirectionType      uint32

	InterfaceAndOperStatusFlags struct {
		HardwareInterface bool
		FilterInterface   bool
		ConnectorPresent  bool
		NotAuthenticated  bool
		NotMediaConnected bool
		Paused            bool
		LowPower          bool
		EndPointInterface bool
	}

	OperStatus        uint32
	AdminStatus       uint32
	MediaConnectState uint32
	NetworkGUID       windows.GUID
	ConnectionType    uint32

	TransmitLinkSpeed uint64
	ReceiveLinkSpeed  uint64

	InOctets           uint64
	InUcastPkts        uint64
	InNUcastPkts       uint64
	InDiscards         uint64
	InErrors           uint64
	InUnknownProtos    uint64
	InUcastOctets      uint64
	InMulticastOctets  uint64
	InBroadcastOctets  uint64
	OutOctets          uint64
	OutUcastPkts       uint64
	OutNUcastPkts      uint64
	OutDiscards        uint64
	OutErrors          uint64
	OutUcastOctets     uint64
	OutMulticastOctets uint64
	OutBroadcastOctets uint64
	OutQLen            uint64
}

var (
	wlanapi             = syscall.NewLazyDLL("wlanapi.dll")
	hWlanOpenHandle     = wlanapi.NewProc("WlanOpenHandle")
	hWlanCloseHandle    = wlanapi.NewProc("WlanCloseHandle")
	hWlanQueryInterface = wlanapi.NewProc("WlanQueryInterface")
)

func (env *ShellEnvironment) getWiFiSSID(guid windows.GUID) string {
	// Query wifi connection state
	var pdwNegotiatedVersion uint32
	var phClientHandle uint32
	e, _, err := hWlanOpenHandle.Call(uintptr(uint32(2)), uintptr(unsafe.Pointer(nil)), uintptr(unsafe.Pointer(&pdwNegotiatedVersion)), uintptr(unsafe.Pointer(&phClientHandle)))
	if e != 0 {
		env.Log(Error, "getAllWifiSSID", err.Error())
		return ""
	}

	// defer closing handle
	defer func() {
		_, _, _ = hWlanCloseHandle.Call(uintptr(phClientHandle), uintptr(unsafe.Pointer(nil)))
	}()

	var dataSize uint16
	var wlanAttr *WLAN_CONNECTION_ATTRIBUTES

	e, _, _ = hWlanQueryInterface.Call(uintptr(phClientHandle),
		uintptr(unsafe.Pointer(&guid)),
		uintptr(7), // wlan_intf_opcode_current_connection
		uintptr(unsafe.Pointer(nil)),
		uintptr(unsafe.Pointer(&dataSize)),
		uintptr(unsafe.Pointer(&wlanAttr)),
		uintptr(unsafe.Pointer(nil)))
	if e != 0 {
		env.Log(Error, "parseWlanInterface", "wlan_intf_opcode_current_connection error")
		return ""
	}

	ssid := wlanAttr.wlanAssociationAttributes.dot11Ssid
	if ssid.uSSIDLength <= 0 {
		return ""
	}
	return string(ssid.ucSSID[0:ssid.uSSIDLength])
}

type WLAN_CONNECTION_ATTRIBUTES struct { //nolint: revive
	isState                   uint32      //nolint: unused
	wlanConnectionMode        uint32      //nolint: unused
	strProfileName            [256]uint16 //nolint: unused
	wlanAssociationAttributes WLAN_ASSOCIATION_ATTRIBUTES
	wlanSecurityAttributes    WLAN_SECURITY_ATTRIBUTES //nolint: unused
}

type WLAN_ASSOCIATION_ATTRIBUTES struct { //nolint: revive
	dot11Ssid         DOT11_SSID
	dot11BssType      uint32   //nolint: unused
	dot11Bssid        [6]uint8 //nolint: unused
	dot11PhyType      uint32   //nolint: unused
	uDot11PhyIndex    uint32   //nolint: unused
	wlanSignalQuality uint32   //nolint: unused
	ulRxRate          uint32   //nolint: unused
	ulTxRate          uint32   //nolint: unused
}

type WLAN_SECURITY_ATTRIBUTES struct { //nolint: revive
	bSecurityEnabled     uint32 //nolint: unused
	bOneXEnabled         uint32 //nolint: unused
	dot11AuthAlgorithm   uint32 //nolint: unused
	dot11CipherAlgorithm uint32 //nolint: unused
}

type DOT11_SSID struct { //nolint: revive
	uSSIDLength uint32
	ucSSID      [32]uint8
}
