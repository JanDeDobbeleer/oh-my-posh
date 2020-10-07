package main

type osicon struct {
	props          *properties
	env            environmentInfo
}

const (
	//MacosIcon the icon to use for macOS
	macOSIcon Property = "macos_icon"
	//LinuxIcon the icon to use for linux
	LinuxIcon Property = "linux_icon"
	//WindowsIcon the icon to use for windows
	WindowsIcon Property = "windows_icon"
)

func (n *osicon) enabled() bool {
	return true
}

func (n *osicon) string() string {
	goos := n.env.getRuntimeGOOS()
	switch goos {
	case "windows":
		return n.props.getString(WindowsIcon, "\uF17A")
	case "darwin":
		return n.props.getString(macOSIcon, "\uF179")
	case "linux":
		return n.props.getString(LinuxIcon, "\uF17C")
	default:
		return ""
	}
}

func (n *osicon) init(props *properties, env environmentInfo) {
	n.props = props
	n.env = env
}