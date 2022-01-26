package main

import "oh-my-posh/environment"

type osInfo struct {
	props Properties
	env   environment.Environment

	Icon string
}

const (
	// MacOS the string/icon to use for MacOS
	MacOS Property = "macos"
	// Linux the string/icon to use for linux
	Linux Property = "linux"
	// Windows the string/icon to use for windows
	Windows Property = "windows"
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

func (oi *osInfo) template() string {
	return "{{ if .WSL }}WSL at {{ end }}{{.Icon}}"
}

func (oi *osInfo) enabled() bool {
	goos := oi.env.GOOS()
	switch goos {
	case environment.WindowsPlatform:
		oi.Icon = oi.props.GetString(Windows, "\uE62A")
	case environment.DarwinPlatform:
		oi.Icon = oi.props.GetString(MacOS, "\uF179")
	case environment.LinuxPlatform:
		platform := oi.env.Platform()
		displayDistroName := oi.props.GetBool(DisplayDistroName, false)
		if displayDistroName {
			oi.Icon = platform
			break
		}
		oi.Icon = oi.getDistroIcon(platform)
	default:
		oi.Icon = goos
	}
	return true
}

func (oi *osInfo) getDistroIcon(distro string) string {
	switch distro {
	case "alpine":
		return oi.props.GetString(Alpine, "\uF300")
	case "aosc":
		return oi.props.GetString(Aosc, "\uF301")
	case "arch":
		return oi.props.GetString(Arch, "\uF303")
	case "centos":
		return oi.props.GetString(Centos, "\uF304")
	case "coreos":
		return oi.props.GetString(Coreos, "\uF305")
	case "debian":
		return oi.props.GetString(Debian, "\uF306")
	case "devuan":
		return oi.props.GetString(Devuan, "\uF307")
	case "raspbian":
		return oi.props.GetString(Raspbian, "\uF315")
	case "elementary":
		return oi.props.GetString(Elementary, "\uF309")
	case "fedora":
		return oi.props.GetString(Fedora, "\uF30a")
	case "gentoo":
		return oi.props.GetString(Gentoo, "\uF30d")
	case "mageia":
		return oi.props.GetString(Mageia, "\uF310")
	case "manjaro":
		return oi.props.GetString(Manjaro, "\uF312")
	case "mint":
		return oi.props.GetString(Mint, "\uF30e")
	case "nixos":
		return oi.props.GetString(Nixos, "\uF313")
	case "opensuse":
		return oi.props.GetString(Opensuse, "\uF314")
	case "sabayon":
		return oi.props.GetString(Sabayon, "\uF317")
	case "slackware":
		return oi.props.GetString(Slackware, "\uF319")
	case "ubuntu":
		return oi.props.GetString(Ubuntu, "\uF31b")
	}
	return oi.props.GetString(Linux, "\uF17C")
}

func (oi *osInfo) init(props Properties, env environment.Environment) {
	oi.props = props
	oi.env = env
}
