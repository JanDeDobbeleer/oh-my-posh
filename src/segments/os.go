package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
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
	iconMap := map[string]string{
		"alma":                "\uF31D",
		"almalinux9":          "\uF31D",
		"alpine":              "\uF300",
		"aosc":                "\uF301",
		"arch":                "\uF303",
		"centos":              "\uF304",
		"coreos":              "\uF305",
		"debian":              "\uF306",
		"devuan":              "\uF307",
		"raspbian":            "\uF315",
		"elementary":          "\uF309",
		"fedora":              "\uF30a",
		"gentoo":              "\uF30d",
		"mageia":              "\uF310",
		"manjaro":             "\uF312",
		"mint":                "\uF30e",
		"nixos":               "\uF313",
		"opensuse":            "\uF314",
		"opensuse-tumbleweed": "\uF314",
		"redhat":              "\uF316",
		"rocky":               "\uF32B",
		"sabayon":             "\uF317",
		"slackware":           "\uF319",
		"ubuntu":              "\uF31b",
		"android":             "\uf17b",
	}

	if icon, ok := iconMap[distro]; ok {
		return oi.props.GetString(properties.Property(distro), icon)
	}

	icon := oi.props.GetString(properties.Property(distro), "")
	if len(icon) > 0 {
		return icon
	}

	return oi.props.GetString(Linux, "\uF17C")
}

func (oi *Os) Init(props properties.Properties, env platform.Environment) {
	oi.props = props
	oi.env = env
}
