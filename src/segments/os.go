package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

type Os struct {
	Base

	Icon template.Markup
}

const (
	MacOS             options.Option = "macos"
	Linux             options.Option = "linux"
	Windows           options.Option = "windows"
	Android           options.Option = "android"
	DisplayDistroName options.Option = "display_distro_name"
)

func (oi *Os) Template() string {
	return " {{ if .WSL }}WSL at {{ end }}{{.Icon}} "
}

func (oi *Os) Enabled() bool {
	goos := oi.env.GOOS()
	switch goos {
	case runtime.WINDOWS:
		oi.Icon = oi.options.Markup(Windows, "\uE62A")
	case runtime.DARWIN:
		oi.Icon = oi.options.Markup(MacOS, "\uF179")
	case runtime.LINUX, runtime.FREEBSD:
		pf := oi.env.Platform()
		displayDistroName := oi.options.Bool(DisplayDistroName, false)
		if displayDistroName {
			icon := oi.options.String(options.Option(pf), "")
			if icon == "" {
				oi.Icon = template.EscapeMarkup(pf)
				break
			}
			oi.Icon = template.RawMarkup(icon)
			break
		}
		oi.Icon = oi.getDistroIcon(pf)
	case runtime.ANDROID:
		oi.Icon = oi.options.Markup(Android, "\ue70e")
	default:
		oi.Icon = template.EscapeMarkup(goos)
	}
	return true
}

func (oi *Os) getDistroIcon(distro string) template.Markup {
	iconMap := map[string]string{
		"alma":                "\uf31d",
		"almalinux":           "\uf31d",
		"almalinux9":          "\uf31d",
		"alpine":              "\uf300",
		"android":             "\ue70e",
		"aosc":                "\uf301",
		"arch":                "\uf303",
		"artix":               "\uf31f",
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
		return oi.options.Markup(options.Option(distro), icon)
	}

	icon := oi.options.String(options.Option(distro), "")
	if len(icon) > 0 {
		return template.RawMarkup(icon)
	}

	return oi.options.Markup(Linux, "\uF17C")
}
