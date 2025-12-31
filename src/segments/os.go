package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

type Os struct {
	Base

	Icon string
}

const (
	// MacOS the string/icon to use for MacOS
	MacOS options.Option = "macos"
	// Linux the string/icon to use for linux
	Linux options.Option = "linux"
	// Windows the string/icon to use for windows
	Windows options.Option = "windows"
	// Android the string/icon to use for android
	Android options.Option = "android"
	// DisplayDistroName display the distro name or not
	DisplayDistroName options.Option = "display_distro_name"
)

func (oi *Os) Template() string {
	return " {{ if .WSL }}WSL at {{ end }}{{.Icon}} "
}

func (oi *Os) Enabled() bool {
	goos := oi.env.GOOS()
	switch goos {
	case runtime.WINDOWS:
		oi.Icon = oi.options.String(Windows, "\uE62A")
	case runtime.DARWIN:
		oi.Icon = oi.options.String(MacOS, "\uF179")
	case runtime.LINUX, runtime.FREEBSD:
		pf := oi.env.Platform()
		displayDistroName := oi.options.Bool(DisplayDistroName, false)
		if displayDistroName {
			oi.Icon = oi.options.String(options.Option(pf), pf)
			break
		}
		oi.Icon = oi.getDistroIcon(pf)
	case runtime.ANDROID:
		oi.Icon = oi.options.String(Android, "\ue70e")
	default:
		oi.Icon = goos
	}
	return true
}

func (oi *Os) getDistroIcon(distro string) string {
	iconMap := map[string]string{
		"alma":                "\uF31D",
		"almalinux":           "\uF31D",
		"almalinux9":          "\uF31D",
		"alpine":              "\uF300",
		"android":             "\ue70e",
		"aosc":                "\uF301",
		"arch":                "\uF303",
		"centos":              "\uF304",
		"coreos":              "\uF305",
		"debian":              "\uF306",
		"deepin":              "\uF321",
		"devuan":              "\uF307",
		"elementary":          "\uF309",
		"endeavouros":         "\uF322",
		"fedora":              "\uF30a",
		"freebsd":             "\U000f08e0",
		"gentoo":              "\uF30d",
		"kali":                "\uf327",
		"mageia":              "\uF310",
		"manjaro":             "\uF312",
		"mint":                "\U000f08ed",
		"nixos":               "\uF313",
		"opensuse":            "\uF314",
		"opensuse-tumbleweed": "\uF314",
		"raspbian":            "\uF315",
		"redhat":              "\uF316",
		"rocky":               "\uF32B",
		"sabayon":             "\uF317",
		"slackware":           "\uF319",
		"ubuntu":              "\uF31b",
		"neon":                "\uf331",
	}

	if icon, ok := iconMap[distro]; ok {
		return oi.options.String(options.Option(distro), icon)
	}

	icon := oi.options.String(options.Option(distro), "")
	if len(icon) > 0 {
		return icon
	}

	return oi.options.String(Linux, "\uF17C")
}
