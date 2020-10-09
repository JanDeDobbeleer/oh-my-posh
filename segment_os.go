package main

type osInfo struct {
	props *properties
	env   environmentInfo
}

const (
	//MacOS the string/icon to use for MacOS
	MacOS Property = "macos"
	//Linux the string/icon to use for linux
	Linux Property = "linux"
	//Windows the string/icon to use for windows
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
		return n.props.getString(MacOS, "\uF179")
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
