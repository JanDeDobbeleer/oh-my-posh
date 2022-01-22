package main

type osInfo struct {
	props Properties
	env   Environment

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

func (n *osInfo) enabled() bool {
	goos := n.env.getRuntimeGOOS()
	switch goos {
	case windowsPlatform:
		n.Icon = n.props.getString(Windows, "\uE62A")
	case darwinPlatform:
		n.Icon = n.props.getString(MacOS, "\uF179")
	case linuxPlatform:
		platform := n.env.getPlatform()
		displayDistroName := n.props.getBool(DisplayDistroName, false)
		if displayDistroName {
			n.Icon = platform
			break
		}
		n.Icon = n.getDistroIcon(platform)
	default:
		n.Icon = goos
	}
	return true
}

func (n *osInfo) string() string {
	segmentTemplate := n.props.getString(SegmentTemplate, "{{ if .WSL }}WSL at {{ end }}{{.Icon}}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  n,
		Env:      n.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (n *osInfo) getDistroIcon(distro string) string {
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
