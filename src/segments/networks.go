package segments

import (
	"fmt"
	"math"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"oh-my-posh/regex"
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
	SSIDAbbr := n.props.GetInt(properties.Property("SSIDAbbr"), 0)
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
		switch LinkSpeedUnit {
		case bps:
			str += fmt.Sprintf("%s%d/%d%s", AT, network.TransmitLinkSpeed, network.ReceiveLinkSpeed, bps)
		case b:
			str += fmt.Sprintf("%s%d/%d%s", AT, network.TransmitLinkSpeed, network.ReceiveLinkSpeed, b)
		case Kbps:
			str += fmt.Sprintf("%s%.3g/%.3g%s", AT, float64(network.TransmitLinkSpeed)/math.Pow10(3), float64(network.ReceiveLinkSpeed)/math.Pow10(3), Kbps)
		case K:
			str += fmt.Sprintf("%s%.3g/%.3g%s", AT, float64(network.TransmitLinkSpeed)/math.Pow10(3), float64(network.ReceiveLinkSpeed)/math.Pow10(3), K)
		case Mbps:
			str += fmt.Sprintf("%s%.3g/%.3g%s", AT, float64(network.TransmitLinkSpeed)/math.Pow10(6), float64(network.ReceiveLinkSpeed)/math.Pow10(6), Mbps)
		case M:
			str += fmt.Sprintf("%s%.3g/%.3g%s", AT, float64(network.TransmitLinkSpeed)/math.Pow10(6), float64(network.ReceiveLinkSpeed)/math.Pow10(6), M)
		case Gbps:
			str += fmt.Sprintf("%s%.3g/%.3g%s", AT, float64(network.TransmitLinkSpeed)/math.Pow10(9), float64(network.ReceiveLinkSpeed)/math.Pow10(9), Gbps)
		case G:
			str += fmt.Sprintf("%s%.3g/%.3g%s", AT, float64(network.TransmitLinkSpeed)/math.Pow10(9), float64(network.ReceiveLinkSpeed)/math.Pow10(9), G)
		case Tbps:
			str += fmt.Sprintf("%s%.3g/%.3g%s", AT, float64(network.TransmitLinkSpeed)/math.Pow10(12), float64(network.ReceiveLinkSpeed)/math.Pow10(12), Tbps)
		case T:
			str += fmt.Sprintf("%s%.3g/%.3g%s", AT, float64(network.TransmitLinkSpeed)/math.Pow10(12), float64(network.ReceiveLinkSpeed)/math.Pow10(12), T)
		case Auto:
			TransmitUnit := (len(fmt.Sprintf("%d", network.TransmitLinkSpeed)) - 1) / 3
			if TransmitUnit > 4 {
				TransmitUnit = 4
			}
			ReceiveUnit := (len(fmt.Sprintf("%d", network.TransmitLinkSpeed)) - 1) / 3
			if ReceiveUnit > 4 {
				ReceiveUnit = 4
			}
			str += fmt.Sprintf("%s%.3g", AT, float64(network.TransmitLinkSpeed)/math.Pow10(3*TransmitUnit))
			if TransmitUnit != ReceiveUnit {
				switch TransmitUnit {
				case 0:
					str += string(bps)
				case 1:
					str += string(Kbps)
				case 2:
					str += string(Mbps)
				case 3:
					str += string(Gbps)
				case 4:
					str += string(Tbps)
				}
			}
			str += fmt.Sprintf("/%.3g", float64(network.ReceiveLinkSpeed)/math.Pow10(3*ReceiveUnit))
			switch ReceiveUnit {
			case 0:
				str += string(bps)
			case 1:
				str += string(Kbps)
			case 2:
				str += string(Mbps)
			case 3:
				str += string(Gbps)
			case 4:
				str += string(Tbps)
			}
		case A:
			TransmitUnit := (len(fmt.Sprintf("%d", network.TransmitLinkSpeed)) - 1) / 3
			if TransmitUnit > 4 {
				TransmitUnit = 4
			}
			ReceiveUnit := (len(fmt.Sprintf("%d", network.TransmitLinkSpeed)) - 1) / 3
			if ReceiveUnit > 4 {
				ReceiveUnit = 4
			}
			str += fmt.Sprintf("%s%.3g", AT, float64(network.TransmitLinkSpeed)/math.Pow10(3*TransmitUnit))
			if TransmitUnit != ReceiveUnit {
				switch TransmitUnit {
				case 0:
					str += string(b)
				case 1:
					str += string(K)
				case 2:
					str += string(M)
				case 3:
					str += string(G)
				case 4:
					str += string(T)
				}
			}
			str += fmt.Sprintf("/%.3g", float64(network.ReceiveLinkSpeed)/math.Pow10(3*ReceiveUnit))
			switch ReceiveUnit {
			case 0:
				str += string(b)
			case 1:
				str += string(K)
			case 2:
				str += string(M)
			case 3:
				str += string(G)
			case 4:
				str += string(T)
			}
		}
	}
	return str
}
