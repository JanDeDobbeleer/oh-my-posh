package main

type wifi struct {
	props *properties
	env   environmentInfo
}

const (
	//NewProp switches something
	NewProp Property = "WiFi"
)

func (w *wifi) enabled() bool {
	return true
}

func (w *wifi) string() string {
	// newText := w.props.getString(NewProp, "\uEFF1")
	// return newText

	// If in Linux, who knows. Gonna need a physical linux machine to develop this since the wlan interface isn't available in wsl
	// But possible solution is to cat /proc/net/wireless, and if that file doesn't exist then don't show the segment

	// In Windows and in wsl you can use the command `netsh.exe wlan show interfaces`. Note the .exe is important for this to work in wsl.

	// if w.env.getPlatform() == windowsPlatform || w.env.isWsl() {
	// 	output := w.env.runShellCommand(w.env.getShellName(), "netsh.exe wlan show interfaces")
	// 	return s
	// }
	// return "WiFi"

	s := w.env.getPlatform()
	if w.env.isWsl() {
		s += " isWsl"
	}
	return s
}

func (w *wifi) init(props *properties, env environmentInfo) {
	w.props = props
	w.env = env
}
