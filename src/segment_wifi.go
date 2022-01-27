package main

type wifi struct {
	props Properties
	env   Environment

	WifiInfo
}

const (
	defaultTemplate = "{{ if .Error }}{{ .Error }}{{ else }}\uFAA8 {{ .SSID }} {{ .Signal }}% {{ .ReceiveRate }}Mbps{{ end }}"
)

func (w *wifi) template() string {
	return defaultTemplate
}

func (w *wifi) enabled() bool {
	// This segment only supports Windows/WSL for now
	if w.env.Platform() != windowsPlatform && !w.env.IsWsl() {
		return false
	}
	wifiInfo, err := w.env.WifiNetwork()
	displayError := w.props.getBool(DisplayError, false)
	if err != nil && displayError {
		w.Error = err.Error()
		return true
	}
	if err != nil || wifiInfo == nil {
		return false
	}
	w.WifiInfo = *wifiInfo
	return true
}

func (w *wifi) init(props Properties, env Environment) {
	w.props = props
	w.env = env
}
