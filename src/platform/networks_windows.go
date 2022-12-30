package platform

import (
	"errors"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	wlanapi             = syscall.NewLazyDLL("wlanapi.dll")
	hWlanOpenHandle     = wlanapi.NewProc("WlanOpenHandle")
	hWlanCloseHandle    = wlanapi.NewProc("WlanCloseHandle")
	hWlanQueryInterface = wlanapi.NewProc("WlanQueryInterface")
	hWlanEnumInterfaces = wlanapi.NewProc("WlanEnumInterfaces")
)

//nolint:revive
type MIN_IF_TABLE2 struct {
	NumEntries uint64
	Table      [256]MIB_IF_ROW2
}

//nolint:revive
const (
	IF_MAX_STRING_SIZE         uint64 = 256
	IF_MAX_PHYS_ADDRESS_LENGTH uint64 = 32
)

//nolint:revive
type MIB_IF_ROW2 struct {
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

//nolint:revive, unused
type WLAN_INTERFACE_INFO_LIST struct {
	dwNumberOfItems uint32
	dwIndex         uint32
	InterfaceInfo   [1]WLAN_INTERFACE_INFO
}

//nolint:revive
type WLAN_INTERFACE_INFO struct {
	InterfaceGuid           syscall.GUID
	strInterfaceDescription [256]uint16
	isState                 uint32
}

//nolint:revive
const (
	WLAN_MAX_NAME_LENGTH  int64 = 256
	DOT11_SSID_MAX_LENGTH int64 = 32
)

//nolint:revive, unused
type WLAN_CONNECTION_ATTRIBUTES struct {
	isState                   uint32
	wlanConnectionMode        uint32
	strProfileName            [WLAN_MAX_NAME_LENGTH]uint16
	wlanAssociationAttributes WLAN_ASSOCIATION_ATTRIBUTES
	wlanSecurityAttributes    WLAN_SECURITY_ATTRIBUTES
}

//nolint:revive, unused
type WLAN_ASSOCIATION_ATTRIBUTES struct {
	dot11Ssid         DOT11_SSID
	dot11BssType      uint32
	dot11Bssid        [6]uint8
	dot11PhyType      uint32
	uDot11PhyIndex    uint32
	wlanSignalQuality uint32
	ulRxRate          uint32
	ulTxRate          uint32
}

//nolint:revive, unused
type WLAN_SECURITY_ATTRIBUTES struct {
	bSecurityEnabled     uint32
	bOneXEnabled         uint32
	dot11AuthAlgorithm   uint32
	dot11CipherAlgorithm uint32
}

//nolint:revive
type DOT11_SSID struct {
	uSSIDLength uint32
	ucSSID      [DOT11_SSID_MAX_LENGTH]uint8
}

func (env *Shell) getConnections() []*Connection {
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

	if wifi, err := env.wifiNetwork(); err == nil {
		networks = append(networks, wifi)
	}

	return networks
}

func (env *Shell) wifiNetwork() (*Connection, error) {
	env.Trace(time.Now(), "wifiNetwork")
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

func (env *Shell) parseNetworkInterface(network *WLAN_INTERFACE_INFO, clientHandle uint32) (*Connection, error) {
	info := Connection{
		Type: WIFI,
	}

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
		env.Error("parseNetworkInterface", err)
		return &info, err
	}

	// SSID
	ssid := wlanAttr.wlanAssociationAttributes.dot11Ssid
	if ssid.uSSIDLength > 0 {
		info.SSID = string(ssid.ucSSID[0:ssid.uSSIDLength])
	}

	info.TransmitRate = uint64(wlanAttr.wlanAssociationAttributes.ulTxRate / 1024)
	info.ReceiveRate = uint64(wlanAttr.wlanAssociationAttributes.ulRxRate / 1024)

	return &info, nil
}
