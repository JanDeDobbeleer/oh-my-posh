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
		"alma":                "\uf31d",
		"almalinux":           "\uf31d",
		"almalinux9":          "\uf31d",
		"alpine":              "\uf300",
		"android":             "\ue70e",
		"aosc":                "\uf301",
		"arch":                "\uf303",
		"centos":              "\uf304",
		"coreos":              "\uf305",
		"debian":              "\uf306",
		"deepin":              "\uf321",
		"devuan":              "\uf307",
		"elementary":          "\uf309",
		"endeavouros":         "\uf322",
		"fedora":              "\uf30a",
		"freebsd":             "\U000f08e0",
		"gentoo":              "\uf30d",
		"kali":                "\uf327",
		"mageia":              "\uf310",
		"manjaro":             "\uf312",
		"mint":                "\U000f08ed",
		"neon":                "\uf331",
		"nixos":               "\uf313",
		"opensuse":            "\uf314",
		"opensuse-tumbleweed": "\uf314",
		"raspbian":            "\uf315",
		"redhat":              "\uf316",
		"rocky":               "\uf32b",
		"sabayon":             "\uf317",
		"slackware":           "\uf319",
		"ubuntu":              "\uf31b",
		"void":                "\uf32e",
		"zorin":               "\uf32f",
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
