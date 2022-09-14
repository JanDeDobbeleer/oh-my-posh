package segments

import (
	"fmt"
	"math"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"oh-my-posh/regex"
	"strconv"
	"strings"
)

type Networks struct {
	props properties.Properties
	env   environment.Environment

	Error string

	NetworksInfo []environment.NetworkInfo
	Networks     string
	Status       string
}

type Unit string

const (
	Auto Unit = "Auto"
	A    Unit = "A"
	Hide Unit = "Hide"
	bps  Unit = "bps"
	b    Unit = "b"
	Kbps Unit = "Kbps"
	K    Unit = "K"
	Mbps Unit = "Mbps"
	M    Unit = "M"
	Gbps Unit = "Gbps"
	G    Unit = "G"
	Tbps Unit = "Tbps"
	T    Unit = "T"
)

func (n *Networks) Template() string {
	return "{{ if eq n.Status \"Connected\" }} {{ .Networks }} | {{ n.IconConnected }} {{ else }} {{ n.IconDisconnected }}"
}

func (n *Networks) Enabled() bool {
	// This segment only supports Windows/WSL for now
	if n.env.Platform() != environment.WINDOWS && !n.env.IsWsl() {
		return false
	}
	Spliter := n.props.GetString("Spliter", "|")
	// IconConnected := n.props.GetString("IconConnected", "")
	// IconDisconnected := n.props.GetString("IconDisconnected", "")
	networks, err := n.env.GetAllNetworkInterfaces()
	displayError := n.props.GetBool(properties.DisplayError, false)
	if err != nil && displayError {
		n.Error = err.Error()
		return true
	}
	if err != nil || networks == nil {
		return false
	}
	if len(*networks) == 0 {
		n.Status = "Disconnected"
	} else {
		n.Status = "Connected"
		networkstrs := make([]string, 0)
		for _, network := range *networks {
			networkstrs = append(networkstrs, n.ConstructNetworkInfo(network))
		}
		n.Networks = strings.Join(networkstrs, Spliter)
	}
	return true
}

func (n *Networks) Init(props properties.Properties, env environment.Environment) {
	n.props = props
	n.env = env
}

func (n *Networks) ConstructNetworkInfo(network environment.NetworkInfo) string {
	str := ""
	IconEthernet := n.props.GetString("IconEthernet", "")
	IconWiFi := n.props.GetString("IconWiFi", "")
	IconBluetooth := n.props.GetString("IconBluetooth", "")
	IconCellular := n.props.GetString("IconCellular", "")
	IconOther := n.props.GetString("IconOther", "")
	NDISPhysicalMeidaTypeMap := make(map[environment.NDIS_PHYSICAL_MEDIUM]string)
	NDISPhysicalMeidaTypeMap[environment.NdisPhysicalMedium802_3] = IconEthernet
	NDISPhysicalMeidaTypeMap[environment.NdisPhysicalMediumNative802_11] = IconWiFi
	NDISPhysicalMeidaTypeMap[environment.NdisPhysicalMediumBluetooth] = IconBluetooth
	NDISPhysicalMeidaTypeMap[environment.NdisPhysicalMediumWirelessWan] = IconCellular
	NameMap := make(map[environment.NDIS_PHYSICAL_MEDIUM]string)
	NameMap[environment.NdisPhysicalMedium802_3] = "Ethernet"
	NameMap[environment.NdisPhysicalMediumNative802_11] = "Wi-Fi"
	NameMap[environment.NdisPhysicalMediumBluetooth] = "Bluetooth"
	NameMap[environment.NdisPhysicalMediumWirelessWan] = "Cellular"

	IconAsAT := n.props.GetBool("IconAsAT", false)
	ShowType := n.props.GetBool("ShowType", true)
	ShowSSID := n.props.GetBool("ShowSSID", true)
	SSIDAbbr := n.props.GetInt("SSIDAbbr", 0)
	LinkSpeedFull := n.props.GetBool("LinkSpeedFull", false)
	LinkSpeedUnit := Unit(n.props.GetString("LinkSpeedUnit", "Auto"))

	icon, OK := NDISPhysicalMeidaTypeMap[network.NDISPhysicalMeidaType]
	AT := "@"
	if !OK {
		icon = IconOther
	}
	if IconAsAT {
		AT = icon
	} else {
		str += icon
		if !ShowType && !(ShowSSID && network.NDISPhysicalMeidaType == environment.NdisPhysicalMediumNative802_11) {
			AT = ""
		}
	}

	if ShowSSID && network.NDISPhysicalMeidaType == environment.NdisPhysicalMediumNative802_11 {
		if SSIDAbbr > 0 {
			str += regex.ReplaceAllString(fmt.Sprintf("(.{0,%d}[^ #_-])[ #_-].*", SSIDAbbr-1), network.SSID, "$1")
		} else {
			str += network.SSID
		}
	}

	if ShowType && !(ShowSSID && network.NDISPhysicalMeidaType == environment.NdisPhysicalMediumNative802_11) {
		if name, OK := NameMap[network.NDISPhysicalMeidaType]; OK {
			str += name
		} else {
			str += "Unknown"
		}
	}

	if LinkSpeedUnit != Hide {
		var TransmitLinkSpeed string
		var TransmitLinkSpeedUnit Unit
		var ReceiveLinkSpeed string
		var ReceiveLinkSpeedUnit Unit

		switch LinkSpeedUnit {
		case bps:
			TransmitLinkSpeed = fmt.Sprintf("%d", network.TransmitLinkSpeed)
			TransmitLinkSpeedUnit = bps
			ReceiveLinkSpeed = fmt.Sprintf("%d", network.ReceiveLinkSpeed)
			ReceiveLinkSpeedUnit = bps
		case b:
			TransmitLinkSpeed = fmt.Sprintf("%d", network.TransmitLinkSpeed)
			TransmitLinkSpeedUnit = b
			ReceiveLinkSpeed = fmt.Sprintf("%d", network.ReceiveLinkSpeed)
			ReceiveLinkSpeedUnit = b
		case Kbps:
			TransmitLinkSpeed = strconv.FormatFloat(float64(network.TransmitLinkSpeed)/math.Pow10(3), 'f', -1, 64)
			TransmitLinkSpeedUnit = Kbps
			ReceiveLinkSpeed = strconv.FormatFloat(float64(network.ReceiveLinkSpeed)/math.Pow10(3), 'f', -1, 64)
			ReceiveLinkSpeedUnit = Kbps
		case K:
			TransmitLinkSpeed = strconv.FormatFloat(float64(network.TransmitLinkSpeed)/math.Pow10(3), 'f', -1, 64)
			TransmitLinkSpeedUnit = K
			ReceiveLinkSpeed = strconv.FormatFloat(float64(network.ReceiveLinkSpeed)/math.Pow10(3), 'f', -1, 64)
			ReceiveLinkSpeedUnit = K
		case Mbps:
			TransmitLinkSpeed = strconv.FormatFloat(float64(network.TransmitLinkSpeed)/math.Pow10(6), 'f', -1, 64)
			TransmitLinkSpeedUnit = Mbps
			ReceiveLinkSpeed = strconv.FormatFloat(float64(network.ReceiveLinkSpeed)/math.Pow10(6), 'f', -1, 64)
			ReceiveLinkSpeedUnit = Mbps
		case M:
			TransmitLinkSpeed = strconv.FormatFloat(float64(network.TransmitLinkSpeed)/math.Pow10(6), 'f', -1, 64)
			TransmitLinkSpeedUnit = M
			ReceiveLinkSpeed = strconv.FormatFloat(float64(network.ReceiveLinkSpeed)/math.Pow10(6), 'f', -1, 64)
			ReceiveLinkSpeedUnit = M
		case Gbps:
			TransmitLinkSpeed = strconv.FormatFloat(float64(network.TransmitLinkSpeed)/math.Pow10(9), 'f', -1, 64)
			TransmitLinkSpeedUnit = Gbps
			ReceiveLinkSpeed = strconv.FormatFloat(float64(network.ReceiveLinkSpeed)/math.Pow10(9), 'f', -1, 64)
			ReceiveLinkSpeedUnit = Gbps
		case G:
			TransmitLinkSpeed = strconv.FormatFloat(float64(network.TransmitLinkSpeed)/math.Pow10(9), 'f', -1, 64)
			TransmitLinkSpeedUnit = G
			ReceiveLinkSpeed = strconv.FormatFloat(float64(network.ReceiveLinkSpeed)/math.Pow10(9), 'f', -1, 64)
			ReceiveLinkSpeedUnit = G
		case Tbps:
			TransmitLinkSpeed = strconv.FormatFloat(float64(network.TransmitLinkSpeed)/math.Pow10(12), 'f', -1, 64)
			TransmitLinkSpeedUnit = Tbps
			ReceiveLinkSpeed = strconv.FormatFloat(float64(network.ReceiveLinkSpeed)/math.Pow10(12), 'f', -1, 64)
			ReceiveLinkSpeedUnit = Tbps
		case T:
			TransmitLinkSpeed = strconv.FormatFloat(float64(network.TransmitLinkSpeed)/math.Pow10(12), 'f', -1, 64)
			TransmitLinkSpeedUnit = T
			ReceiveLinkSpeed = strconv.FormatFloat(float64(network.ReceiveLinkSpeed)/math.Pow10(12), 'f', -1, 64)
			ReceiveLinkSpeedUnit = T
		case Auto:
			TransmitSpeedUnitIndex := (len(fmt.Sprintf("%d", network.TransmitLinkSpeed)) - 1) / 3
			if TransmitSpeedUnitIndex > 4 {
				TransmitSpeedUnitIndex = 4
			}
			switch TransmitSpeedUnitIndex {
			case 0:
				TransmitLinkSpeedUnit = bps
			case 1:
				TransmitLinkSpeedUnit = Kbps
			case 2:
				TransmitLinkSpeedUnit = Mbps
			case 3:
				TransmitLinkSpeedUnit = Gbps
			case 4:
				TransmitLinkSpeedUnit = Tbps
			}
			ReceiveSpeedUnitIndex := (len(fmt.Sprintf("%d", network.TransmitLinkSpeed)) - 1) / 3
			if ReceiveSpeedUnitIndex > 4 {
				ReceiveSpeedUnitIndex = 4
			}
			switch ReceiveSpeedUnitIndex {
			case 0:
				ReceiveLinkSpeedUnit = bps
			case 1:
				ReceiveLinkSpeedUnit = Kbps
			case 2:
				ReceiveLinkSpeedUnit = Mbps
			case 3:
				ReceiveLinkSpeedUnit = Gbps
			case 4:
				ReceiveLinkSpeedUnit = Tbps
			}
			if TransmitSpeedUnitIndex == 0 {
				TransmitLinkSpeed = fmt.Sprintf("%d", network.TransmitLinkSpeed)
			} else {
				TransmitLinkSpeed = fmt.Sprintf("%.3g", float64(network.TransmitLinkSpeed)/math.Pow10(3*TransmitSpeedUnitIndex))
			}
			if ReceiveSpeedUnitIndex == 0 {
				ReceiveLinkSpeed = fmt.Sprintf("%d", network.ReceiveLinkSpeed)
			} else {
				ReceiveLinkSpeed = fmt.Sprintf("%.3g", float64(network.ReceiveLinkSpeed)/math.Pow10(3*ReceiveSpeedUnitIndex))
			}
		case A:
			TransmitSpeedUnitIndex := (len(fmt.Sprintf("%d", network.TransmitLinkSpeed)) - 1) / 3
			if TransmitSpeedUnitIndex > 4 {
				TransmitSpeedUnitIndex = 4
			}
			switch TransmitSpeedUnitIndex {
			case 0:
				TransmitLinkSpeedUnit = b
			case 1:
				TransmitLinkSpeedUnit = K
			case 2:
				TransmitLinkSpeedUnit = M
			case 3:
				TransmitLinkSpeedUnit = G
			case 4:
				TransmitLinkSpeedUnit = T
			}
			ReceiveSpeedUnitIndex := (len(fmt.Sprintf("%d", network.TransmitLinkSpeed)) - 1) / 3
			if ReceiveSpeedUnitIndex > 4 {
				ReceiveSpeedUnitIndex = 4
			}
			switch ReceiveSpeedUnitIndex {
			case 0:
				ReceiveLinkSpeedUnit = b
			case 1:
				ReceiveLinkSpeedUnit = K
			case 2:
				ReceiveLinkSpeedUnit = M
			case 3:
				ReceiveLinkSpeedUnit = G
			case 4:
				ReceiveLinkSpeedUnit = T
			}
			if TransmitSpeedUnitIndex == 0 {
				TransmitLinkSpeed = fmt.Sprintf("%d", network.TransmitLinkSpeed)
			} else {
				TransmitLinkSpeed = fmt.Sprintf("%.3g", float64(network.TransmitLinkSpeed)/math.Pow10(3*TransmitSpeedUnitIndex))
			}
			if ReceiveSpeedUnitIndex == 0 {
				ReceiveLinkSpeed = fmt.Sprintf("%d", network.ReceiveLinkSpeed)
			} else {
				ReceiveLinkSpeed = fmt.Sprintf("%.3g", float64(network.ReceiveLinkSpeed)/math.Pow10(3*ReceiveSpeedUnitIndex))
			}
		}

		if LinkSpeedFull || TransmitLinkSpeedUnit != ReceiveLinkSpeedUnit {
			str += fmt.Sprintf("%s%s%s/%s%s", AT, TransmitLinkSpeed, TransmitLinkSpeedUnit, ReceiveLinkSpeed, ReceiveLinkSpeedUnit)
		} else if TransmitLinkSpeed == ReceiveLinkSpeed {
			str += fmt.Sprintf("%s%s%s", AT, TransmitLinkSpeed, TransmitLinkSpeedUnit)
		} else {
			str += fmt.Sprintf("%s%s/%s%s", AT, ReceiveLinkSpeed, TransmitLinkSpeed, TransmitLinkSpeedUnit)
		}
	}

	if len(str) == 0 {
		return icon
	}

	return str
}
