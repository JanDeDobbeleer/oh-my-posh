package segments

import (
	"oh-my-posh/platform"
	"oh-my-posh/properties"
)

type Os struct {
	props properties.Properties
	env   platform.Environment

	Icon string
}

const (
	// MacOS the string/icon to use for MacOS
	MacOS properties.Property = "macos"
	// Linux the string/icon to use for linux
	Linux properties.Property = "linux"
	// Windows the string/icon to use for windows
	Windows properties.Property = "windows"
	// Alma the string/icon to use for Alma
	Alma properties.Property = "alma"
	// Alpine the string/icon to use for Alpine
	Alpine properties.Property = "alpine"
	// Aosc the string/icon to use for Aosc
	Aosc properties.Property = "aosc"
	// Arch the string/icon to use for Arch
	Arch properties.Property = "arch"
	// Centos the string/icon to use for Centos
	Centos properties.Property = "centos"
	// Coreos the string/icon to use for Coreos
	Coreos properties.Property = "coreos"
	// Debian the string/icon to use for Debian
	Debian properties.Property = "debian"
	// Devuan the string/icon to use for Devuan
	Devuan properties.Property = "devuan"
	// Raspbian the string/icon to use for Raspbian
	Raspbian properties.Property = "raspbian"
	// Elementary the string/icon to use for Elementary
	Elementary properties.Property = "elementary"
	// Fedora the string/icon to use for Fedora
	Fedora properties.Property = "fedora"
	// Gentoo the string/icon to use for Gentoo
	Gentoo properties.Property = "gentoo"
	// Mageia the string/icon to use for Mageia
	Mageia properties.Property = "mageia"
	// Manjaro the string/icon to use for Manjaro
	Manjaro properties.Property = "manjaro"
	// Mint the string/icon to use for Mint
	Mint properties.Property = "mint"
	// Nixos the string/icon to use for Nixos
	Nixos properties.Property = "nixos"
	// Opensuse the string/icon to use for Opensuse
	Opensuse properties.Property = "opensuse"
	// Redhat the string/icon to use for Redhat
	Redhat properties.Property = "redhat"
	// Rocky the string/icon to use for Rocky linux
	Rocky properties.Property = "rocky"
	// Sabayon the string/icon to use for Sabayon
	Sabayon properties.Property = "sabayon"
	// Slackware the string/icon to use for Slackware
	Slackware properties.Property = "slackware"
	// Ubuntu the string/icon to use for Ubuntu
	Ubuntu properties.Property = "ubuntu"
	// DisplayDistroName display the distro name or not
	DisplayDistroName properties.Property = "display_distro_name"
)

func (oi *Os) Template() string {
	return " {{ if .WSL }}WSL at {{ end }}{{.Icon}} "
}

func (oi *Os) Enabled() bool {
	goos := oi.env.GOOS()
	switch goos {
	case platform.WINDOWS:
		oi.Icon = oi.props.GetString(Windows, "\uE62A")
	case platform.DARWIN:
		oi.Icon = oi.props.GetString(MacOS, "\uF179")
	case platform.LINUX:
		pf := oi.env.Platform()
		displayDistroName := oi.props.GetBool(DisplayDistroName, false)
		if displayDistroName {
			oi.Icon = pf
			break
		}
		oi.Icon = oi.getDistroIcon(pf)
	default:
		oi.Icon = goos
	}
	return true
}

func (oi *Os) getDistroIcon(distro string) string {
	switch distro {
	case "alma":
		return oi.props.GetString(Alma, "\uF31D")
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
	case "redhat":
		return oi.props.GetString(Redhat, "\uF316")
	case "rocky":
		return oi.props.GetString(Rocky, "\uF32B")
	case "sabayon":
		return oi.props.GetString(Sabayon, "\uF317")
	case "slackware":
		return oi.props.GetString(Slackware, "\uF319")
	case "ubuntu":
		return oi.props.GetString(Ubuntu, "\uF31b")
	}
	return oi.props.GetString(Linux, "\uF17C")
}

func (oi *Os) Init(props properties.Properties, env platform.Environment) {
	oi.props = props
	oi.env = env
}
