package main

type osInfo struct {
	props          *properties
	env            environmentInfo
}

const (
	//Macos the string to use for macOS
	macOS Property = "macos"
	//LinuxIcon the string to use for linux
	Linux Property = "linux"
	//WindowsIcon the icon to use for windows
	Windows Property = "windows"
)

func (n *osInfo) enabled() bool {
	return true
}

func (n *osInfo) string() string {
	goos := n.env.getRuntimeGOOS()
	switch goos {
	case "windows":
		return n.props.getString(Windows, "\uE62A")
	case "darwin":
		return n.props.getString(macOS, "\uF179")
	case "linux":
		return n.props.getString(Linux, "\uF17C")
	default:
		return ""
	}
}

func (n *osInfo) init(props *properties, env environmentInfo) {
	n.props = props
	n.env = env
}