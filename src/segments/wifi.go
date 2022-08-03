package segments

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type Wifi struct {
	props properties.Properties
	env   environment.Environment

	Error string

	environment.WifiInfo
}

const (
	defaultTemplate = " {{ if .Error }}{{ .Error }}{{ else }}\uFAA8 {{ .SSID }} {{ .Signal }}% {{ .ReceiveRate }}Mbps{{ end }} "
)

func (w *Wifi) Template() string {
	return defaultTemplate
}

func (w *Wifi) Enabled() bool {
	// This segment only supports Windows/WSL for now
	if w.env.Platform() != environment.WINDOWS && !w.env.IsWsl() {
		return false
	}
	wifiInfo, err := w.env.WifiNetwork()
	displayError := w.props.GetBool(properties.DisplayError, false)
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

func (w *Wifi) Init(props properties.Properties, env environment.Environment) {
	w.props = props
	w.env = env
}
