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
	"golang.org/x/sys/windows/registry"
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
	return WINDOWS
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

// Takes a registry path to a key like
//
//	"HKLM\Software\Microsoft\Windows NT\CurrentVersion\EditionID"
//
// The last part of the path is the key to retrieve.
//
// If the path ends in "\", the "(Default)" key in that path is retrieved.
//
// Returns a variant type if successful; nil and an error if not.
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
	rootKey, regPath, found := strings.Cut(path, `\`)
	if !found {
		errorLogMsg := fmt.Sprintf("Error, malformed registry path: '%s'", path)
		env.Log(Error, "WindowsRegistryKeyValue", errorLogMsg)
		return nil, errors.New(errorLogMsg)
	}

	regKey := Base(env, regPath)
	if len(regKey) != 0 {
		regPath = strings.TrimSuffix(regPath, `\`+regKey)
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
		errorLogMsg := fmt.Sprintf("Error, unknown registry key: '%s'", rootKey)
		env.Log(Error, "WindowsRegistryKeyValue", errorLogMsg)
		return nil, errors.New(errorLogMsg)
	}

	k, err := registry.OpenKey(key, regPath, registry.READ)
	if err != nil {
		env.Log(Error, "WindowsRegistryKeyValue", err.Error())
		return nil, err
	}
	_, valType, err := k.GetValue(regKey, nil)
	if err != nil {
		env.Log(Error, "WindowsRegistryKeyValue", err.Error())
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
	env.Log(Debug, "WindowsRegistryKeyValue", fmt.Sprintf("%s(%s): %s", regKey, regValue.ValueType, regValue.String))
	return regValue, nil
}

func (env *ShellEnvironment) InWSLSharedDrive() bool {
	return false
}

func (env *ShellEnvironment) ConvertToWindowsPath(path string) string {
	return strings.ReplaceAll(path, `\`, "/")
}

func (env *ShellEnvironment) ConvertToLinuxPath(path string) string {
	return path
}

var (
	iphlpapi     = syscall.NewLazyDLL("iphlpapi.dll")
	hGetIfTable2 = iphlpapi.NewProc("GetIfTable2")
)

const (
	// see https://docs.microsoft.com/en-us/windows/win32/api/netioapi/ns-netioapi-mib_if_row2

	// InterfaceType
	IF_TYPE_OTHER              IFTYPE = "Other"            // 1
	IF_TYPE_ETHERNET_CSMACD    IFTYPE = "Ethernet/802.3"   // 6
	IF_TYPE_ISO88025_TOKENRING IFTYPE = "Token Ring/802.5" // 9
	IF_TYPE_FDDI               IFTYPE = "FDDI"             // 15
	IF_TYPE_PPP                IFTYPE = "PPP"              // 23
	IF_TYPE_SOFTWARE_LOOPBACK  IFTYPE = "Loopback"         // 24
	IF_TYPE_ATM                IFTYPE = "ATM"              // 37
	IF_TYPE_IEEE80211          IFTYPE = "Wi-Fi/802.11"     // 71
	IF_TYPE_TUNNEL             IFTYPE = "Tunnel"           // 131
	IF_TYPE_IEEE1394           IFTYPE = "FireWire/1394"    // 144
	IF_TYPE_IEEE80216_WMAN     IFTYPE = "WMAN/802.16"      // 237 WiMax
	IF_TYPE_WWANPP             IFTYPE = "WWANPP/GSM"       // 243 GSM
	IF_TYPE_WWANPP2            IFTYPE = "WWANPP/CDMA"      // 244 CDMA
	IF_TYPE_UNKNOWN            IFTYPE = "Unknown"

	// NDISMediaType
	NdisMedium802_3        NDIS_MEDIUM = "802.3"         // 0
	NdisMedium802_5        NDIS_MEDIUM = "802.5"         // 1
	NdisMediumFddi         NDIS_MEDIUM = "FDDI"          // 2
	NdisMediumWan          NDIS_MEDIUM = "WAN"           // 3
	NdisMediumLocalTalk    NDIS_MEDIUM = "LocalTalk"     // 4
	NdisMediumDix          NDIS_MEDIUM = "DIX"           // 5
	NdisMediumArcnetRaw    NDIS_MEDIUM = "ARCNET"        // 6
	NdisMediumArcnet878_2  NDIS_MEDIUM = "ARCNET(878.2)" // 7
	NdisMediumAtm          NDIS_MEDIUM = "ATM"           // 8
	NdisMediumWirelessWan  NDIS_MEDIUM = "WWAN"          // 9
	NdisMediumIrda         NDIS_MEDIUM = "IrDA"          // 10
	NdisMediumBpc          NDIS_MEDIUM = "Broadcast"     // 11
	NdisMediumCoWan        NDIS_MEDIUM = "CO WAN"        // 12
	NdisMedium1394         NDIS_MEDIUM = "1394"          // 13
	NdisMediumInfiniBand   NDIS_MEDIUM = "InfiniBand"    // 14
	NdisMediumTunnel       NDIS_MEDIUM = "Tunnel"        // 15
	NdisMediumNative802_11 NDIS_MEDIUM = "Native 802.11" // 16
	NdisMediumLoopback     NDIS_MEDIUM = "Loopback"      // 17
	NdisMediumWiMax        NDIS_MEDIUM = "WiMax"         // 18
	NdisMediumUnknown      NDIS_MEDIUM = "Unknown"

	// NDISPhysicalMeidaType
	NdisPhysicalMediumUnspecified  NDIS_PHYSICAL_MEDIUM = "Unspecified"   // 0
	NdisPhysicalMediumWirelessLan  NDIS_PHYSICAL_MEDIUM = "Wireless LAN"  // 1
	NdisPhysicalMediumCableModem   NDIS_PHYSICAL_MEDIUM = "Cable Modem"   // 2
	NdisPhysicalMediumPhoneLine    NDIS_PHYSICAL_MEDIUM = "Phone Line"    // 3
	NdisPhysicalMediumPowerLine    NDIS_PHYSICAL_MEDIUM = "Power Line"    // 4
	NdisPhysicalMediumDSL          NDIS_PHYSICAL_MEDIUM = "DSL"           // 5
	NdisPhysicalMediumFibreChannel NDIS_PHYSICAL_MEDIUM = "Fibre Channel" // 6
	NdisPhysicalMedium1394         NDIS_PHYSICAL_MEDIUM = "1394"          // 7
	NdisPhysicalMediumWirelessWan  NDIS_PHYSICAL_MEDIUM = "Wireless WAN"  // 8
	NdisPhysicalMediumNative802_11 NDIS_PHYSICAL_MEDIUM = "Native 802.11" // 9
	NdisPhysicalMediumBluetooth    NDIS_PHYSICAL_MEDIUM = "Bluetooth"     // 10
	NdisPhysicalMediumInfiniband   NDIS_PHYSICAL_MEDIUM = "Infini Band"   // 11
	NdisPhysicalMediumWiMax        NDIS_PHYSICAL_MEDIUM = "WiMax"         // 12
	NdisPhysicalMediumUWB          NDIS_PHYSICAL_MEDIUM = "UWB"           // 13
	NdisPhysicalMedium802_3        NDIS_PHYSICAL_MEDIUM = "802.3"         // 14
	NdisPhysicalMedium802_5        NDIS_PHYSICAL_MEDIUM = "802.5"         // 15
	NdisPhysicalMediumIrda         NDIS_PHYSICAL_MEDIUM = "IrDA"          // 16
	NdisPhysicalMediumWiredWAN     NDIS_PHYSICAL_MEDIUM = "Wired WAN"     // 17
	NdisPhysicalMediumWiredCoWan   NDIS_PHYSICAL_MEDIUM = "Wired CO WAN"  // 18
	NdisPhysicalMediumOther        NDIS_PHYSICAL_MEDIUM = "Other"         // 19
	NdisPhysicalMediumUnknown      NDIS_PHYSICAL_MEDIUM = "Unknown"
)

func (env *ShellEnvironment) GetAllNetworkInterfaces() (*[]NetworkInfo, error) {

	var pIFTable2 *MIN_IF_TABLE2
	hGetIfTable2.Call(uintptr(unsafe.Pointer(&pIFTable2)))

	SSIDs, _ := env.GetAllWifiSSID()
	networks := make([]NetworkInfo, 0)

	for i := 0; i < int(pIFTable2.NumEntries); i++ {
		_if := pIFTable2.Table[i]
		_Alias := strings.TrimRight(syscall.UTF16ToString(_if.Alias[:]), "\x00")
		if _if.PhysicalMediumType != 0 && _if.OperStatus == 1 &&
			strings.LastIndex(_Alias, "-") <= 4 &&
			!strings.HasPrefix(_Alias, "Local Area Connection") {
			network := NetworkInfo{}
			network.Alias = _Alias
			network.Interface = strings.TrimRight(syscall.UTF16ToString(_if.Description[:]), "\x00")
			network.TransmitLinkSpeed = _if.TransmitLinkSpeed
			network.ReceiveLinkSpeed = _if.ReceiveLinkSpeed

			switch _if.Type {
			case 1:
				network.InterfaceType = IF_TYPE_OTHER
			case 6:
				network.InterfaceType = IF_TYPE_ETHERNET_CSMACD
			case 9:
				network.InterfaceType = IF_TYPE_ISO88025_TOKENRING
			case 15:
				network.InterfaceType = IF_TYPE_FDDI
			case 23:
				network.InterfaceType = IF_TYPE_PPP
			case 24:
				network.InterfaceType = IF_TYPE_SOFTWARE_LOOPBACK
			case 37:
				network.InterfaceType = IF_TYPE_ATM
			case 71:
				network.InterfaceType = IF_TYPE_IEEE80211
			case 131:
				network.InterfaceType = IF_TYPE_TUNNEL
			case 144:
				network.InterfaceType = IF_TYPE_IEEE1394
			case 237:
				network.InterfaceType = IF_TYPE_IEEE80216_WMAN
			case 243:
				network.InterfaceType = IF_TYPE_WWANPP
			case 244:
				network.InterfaceType = IF_TYPE_WWANPP2
			default:
				network.InterfaceType = IF_TYPE_UNKNOWN
			}

			switch _if.MediaType {
			case 0:
				network.NDISMediaType = NdisMedium802_3
			case 1:
				network.NDISMediaType = NdisMedium802_5
			case 2:
				network.NDISMediaType = NdisMediumFddi
			case 3:
				network.NDISMediaType = NdisMediumWan
			case 4:
				network.NDISMediaType = NdisMediumLocalTalk
			case 5:
				network.NDISMediaType = NdisMediumDix
			case 6:
				network.NDISMediaType = NdisMediumArcnetRaw
			case 7:
				network.NDISMediaType = NdisMediumArcnet878_2
			case 8:
				network.NDISMediaType = NdisMediumAtm
			case 9:
				network.NDISMediaType = NdisMediumWirelessWan
			case 10:
				network.NDISMediaType = NdisMediumIrda
			case 11:
				network.NDISMediaType = NdisMediumBpc
			case 12:
				network.NDISMediaType = NdisMediumCoWan
			case 13:
				network.NDISMediaType = NdisMedium1394
			case 14:
				network.NDISMediaType = NdisMediumInfiniBand
			case 15:
				network.NDISMediaType = NdisMediumTunnel
			case 16:
				network.NDISMediaType = NdisMediumNative802_11
			case 17:
				network.NDISMediaType = NdisMediumLoopback
			case 18:
				network.NDISMediaType = NdisMediumWiMax
			default:
				network.NDISMediaType = NdisMediumUnknown
			}

			switch _if.PhysicalMediumType {
			case 0:
				network.NDISPhysicalMeidaType = NdisPhysicalMediumUnspecified
			case 1:
				network.NDISPhysicalMeidaType = NdisPhysicalMediumWirelessLan
			case 2:
				network.NDISPhysicalMeidaType = NdisPhysicalMediumCableModem
			case 3:
				network.NDISPhysicalMeidaType = NdisPhysicalMediumPhoneLine
			case 4:
				network.NDISPhysicalMeidaType = NdisPhysicalMediumPowerLine
			case 5:
				network.NDISPhysicalMeidaType = NdisPhysicalMediumDSL
			case 6:
				network.NDISPhysicalMeidaType = NdisPhysicalMediumFibreChannel
			case 7:
				network.NDISPhysicalMeidaType = NdisPhysicalMedium1394
			case 8:
				network.NDISPhysicalMeidaType = NdisPhysicalMediumWirelessWan
			case 9:
				network.NDISPhysicalMeidaType = NdisPhysicalMediumNative802_11
			case 10:
				network.NDISPhysicalMeidaType = NdisPhysicalMediumBluetooth
			case 11:
				network.NDISPhysicalMeidaType = NdisPhysicalMediumInfiniband
			case 12:
				network.NDISPhysicalMeidaType = NdisPhysicalMediumWiMax
			case 13:
				network.NDISPhysicalMeidaType = NdisPhysicalMediumUWB
			case 14:
				network.NDISPhysicalMeidaType = NdisPhysicalMedium802_3
			case 15:
				network.NDISPhysicalMeidaType = NdisPhysicalMedium802_5
			case 16:
				network.NDISPhysicalMeidaType = NdisPhysicalMediumIrda
			case 17:
				network.NDISPhysicalMeidaType = NdisPhysicalMediumWiredWAN
			case 18:
				network.NDISPhysicalMeidaType = NdisPhysicalMediumWiredCoWan
			case 19:
				network.NDISPhysicalMeidaType = NdisPhysicalMediumOther
			default:
				network.NDISPhysicalMeidaType = NdisPhysicalMediumUnknown
			}

			if SSID, OK := SSIDs[network.Interface]; OK {
				network.SSID = SSID
			}

			networks = append(networks, network)
		}
	}
	return &networks, nil
}

type MIN_IF_TABLE2 struct {
	NumEntries uint64
	Table      [256]MIB_IF_ROW2
}

const (
	IF_MAX_STRING_SIZE         uint64 = 256
	IF_MAX_PHYS_ADDRESS_LENGTH uint64 = 32
)

type MIB_IF_ROW2 struct {
	InterfaceLuid            uint64
	InterfaceIndex           uint32
	InterfaceGuid            windows.GUID
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

	InterfaceAndOperStatusFlags uint8

	OperStatus        uint32
	AdminStatus       uint32
	MediaConnectState uint32
	NetworkGuid       windows.GUID
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

func (env *ShellEnvironment) GetAllWifiSSID() (map[string]string, error) {
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

	ssid := make(map[string]string)
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
		if network.isState == 1 {
			wifi, _ := env.parseWlanInterface(network, phClientHandle)
			ssid[wifi.Interface] = wifi.SSID
		}
	}
	return ssid, nil
}

var (
	wlanapi             = syscall.NewLazyDLL("wlanapi.dll")
	hWlanOpenHandle     = wlanapi.NewProc("WlanOpenHandle")
	hWlanCloseHandle    = wlanapi.NewProc("WlanCloseHandle")
	hWlanEnumInterfaces = wlanapi.NewProc("WlanEnumInterfaces")
	hWlanQueryInterface = wlanapi.NewProc("WlanQueryInterface")
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
		return env.parseWlanInterface(network, phClientHandle)
	}
	return nil, errors.New("Not connected")
}

func (env *ShellEnvironment) parseWlanInterface(network *WLAN_INTERFACE_INFO, clientHandle uint32) (*WifiInfo, error) {
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
		env.Log(Error, "parseWlanInterface", "wlan_intf_opcode_current_connection error")
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
		info.PhysType = UNKNOWN
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
		env.Log(Error, "parseWlanInterface", "wlan_intf_opcode_channel_number error")
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
		info.Authentication = UNKNOWN
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
		info.Cipher = UNKNOWN
	}

	return &info, nil
}

type WLAN_INTERFACE_INFO_LIST struct { //nolint: revive
	dwNumberOfItems uint32
	dwIndex         uint32 //nolint: unused
	InterfaceInfo   [256]WLAN_INTERFACE_INFO
}

type WLAN_INTERFACE_INFO struct { //nolint: revive
	InterfaceGuid           syscall.GUID //nolint: revive
	strInterfaceDescription [256]uint16
	isState                 uint32
}

type WLAN_CONNECTION_ATTRIBUTES struct { //nolint: revive
	isState                   uint32      //nolint: unused
	wlanConnectionMode        uint32      //nolint: unused
	strProfileName            [256]uint16 //nolint: unused
	wlanAssociationAttributes WLAN_ASSOCIATION_ATTRIBUTES
	wlanSecurityAttributes    WLAN_SECURITY_ATTRIBUTES
}

type WLAN_ASSOCIATION_ATTRIBUTES struct { //nolint: revive
	dot11Ssid         DOT11_SSID
	dot11BssType      uint32
	dot11Bssid        [6]uint8 //nolint: unused
	dot11PhyType      uint32
	uDot11PhyIndex    uint32 //nolint: unused
	wlanSignalQuality uint32
	ulRxRate          uint32
	ulTxRate          uint32
}

type WLAN_SECURITY_ATTRIBUTES struct { //nolint: revive
	bSecurityEnabled     uint32
	bOneXEnabled         uint32 //nolint: unused
	dot11AuthAlgorithm   uint32
	dot11CipherAlgorithm uint32
}

type DOT11_SSID struct { //nolint: revive
	uSSIDLength uint32
	ucSSID      [32]uint8
}

func (env *ShellEnvironment) DirIsWritable(path string) bool {
	defer env.Trace(time.Now(), "DirIsWritable")
	info, err := os.Stat(path)
	if err != nil {
		env.Log(Error, "DirIsWritable", err.Error())
		return false
	}

	if !info.IsDir() {
		env.Log(Error, "DirIsWritable", "Path isn't a directory")
		return false
	}

	// Check if the user bit is enabled in file permission
	if info.Mode().Perm()&(1<<(uint(7))) == 0 {
		env.Log(Error, "DirIsWritable", "Write permission bit is not set on this file for user")
		return false
	}

	return true
}
