package main

import (
	"fmt"
)

type osInfo struct {
	props Properties
	env   Environment
	OS    string
}

const (
	// MacOS the string/icon to use for MacOS
	MacOS Property = "macos"
	// Linux the string/icon to use for linux
	Linux Property = "linux"
	// Windows the string/icon to use for windows
	Windows Property = "windows"
	// WSL the string/icon to use for WSL
	WSL Property = "wsl"
	// WSLSeparator shows between WSL and Linux properties when WSL is detected
	WSLSeparator Property = "wsl_separator"
	// Alpine the string/icon to use for Alpine
	Alpine Property = "alpine"
	// Aosc the string/icon to use for Aosc
	Aosc Property = "aosc"
	// Arch the string/icon to use for Arch
	Arch Property = "arch"
	// Centos the string/icon to use for Centos
	Centos Property = "centos"
	// Coreos the string/icon to use for Coreos
	Coreos Property = "coreos"
	// Debian the string/icon to use for Debian
	Debian Property = "debian"
	// Devuan the string/icon to use for Devuan
	Devuan Property = "devuan"
	// Raspbian the string/icon to use for Raspbian
	Raspbian Property = "raspbian"
	// Elementary the string/icon to use for Elementary
	Elementary Property = "elementary"
	// Fedora the string/icon to use for Fedora
	Fedora Property = "fedora"
	// Gentoo the string/icon to use for Gentoo
	Gentoo Property = "gentoo"
	// Mageia the string/icon to use for Mageia
	Mageia Property = "mageia"
	// Manjaro the string/icon to use for Manjaro
	Manjaro Property = "manjaro"
	// Mint the string/icon to use for Mint
	Mint Property = "mint"
	// Nixos the string/icon to use for Nixos
	Nixos Property = "nixos"
	// Opensuse the string/icon to use for Opensuse
	Opensuse Property = "opensuse"
	// Sabayon the string/icon to use for Sabayon
	Sabayon Property = "sabayon"
	// Slackware the string/icon to use for Slackware
	Slackware Property = "slackware"
	// Ubuntu the string/icon to use for Ubuntu
	Ubuntu Property = "ubuntu"
	// DisplayDistroName display the distro name or not
	DisplayDistroName Property = "display_distro_name"
)

func (n *osInfo) enabled() bool {
	return true
}

func (n *osInfo) string() string {
	goos := n.env.getRuntimeGOOS()
	switch goos {
	case windowsPlatform:
		n.OS = windowsPlatform
		return n.props.getString(Windows, "\uE62A")
	case darwinPlatform:
		n.OS = darwinPlatform
		return n.props.getString(MacOS, "\uF179")
	case linuxPlatform:
		wsl := n.env.getenv("WSL_DISTRO_NAME")
		p := n.env.getPlatform()
		if len(wsl) == 0 {
			n.OS = p
			return n.getDistroName(p, "")
		}
		n.OS = wsl
		return fmt.Sprintf("%s%s%s",
			n.props.getString(WSL, "WSL"),
			n.props.getString(WSLSeparator, " - "),
			n.getDistroName(p, wsl))
	default:
		n.OS = goos
		return goos
	}
}

func (n *osInfo) getDistroName(distro, defaultName string) string {
	displayDistroName := n.props.getBool(DisplayDistroName, false)
	if displayDistroName && len(defaultName) > 0 {
		return defaultName
	}
	if displayDistroName {
		return distro
	}
	switch distro {
	case "alpine":
		return n.props.getString(Alpine, "\uF300")
	case "aosc":
		return n.props.getString(Aosc, "\uF301")
	case "arch":
		return n.props.getString(Arch, "\uF303")
	case "centos":
		return n.props.getString(Centos, "\uF304")
	case "coreos":
		return n.props.getString(Coreos, "\uF305")
	case "debian":
		return n.props.getString(Debian, "\uF306")
	case "devuan":
		return n.props.getString(Devuan, "\uF307")
	case "raspbian":
		return n.props.getString(Raspbian, "\uF315")
	case "elementary":
		return n.props.getString(Elementary, "\uF309")
	case "fedora":
		return n.props.getString(Fedora, "\uF30a")
	case "gentoo":
		return n.props.getString(Gentoo, "\uF30d")
	case "mageia":
		return n.props.getString(Mageia, "\uF310")
	case "manjaro":
		return n.props.getString(Manjaro, "\uF312")
	case "mint":
		return n.props.getString(Mint, "\uF30e")
	case "nixos":
		return n.props.getString(Nixos, "\uF313")
	case "opensuse":
		return n.props.getString(Opensuse, "\uF314")
	case "sabayon":
		return n.props.getString(Sabayon, "\uF317")
	case "slackware":
		return n.props.getString(Slackware, "\uF319")
	case "ubuntu":
		return n.props.getString(Ubuntu, "\uF31b")
	}
	return n.props.getString(Linux, "\uF17C")
}

func (n *osInfo) init(props Properties, env Environment) {
	n.props = props
	n.env = env
}
