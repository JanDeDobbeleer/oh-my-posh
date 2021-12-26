package main

type wifi struct {
	props properties
	env   environmentInfo

	wifiInfo
}

const (
	defaultTemplate = "{{ if .Error }}{{ .Error }}{{ else }}\uFAA8 {{ .SSID }} {{ .Signal }}% {{ .ReceiveRate }}Mbps{{ end }}"
)

func (w *wifi) enabled() bool {
	// This segment only supports Windows/WSL for now
	if w.env.getPlatform() != windowsPlatform && !w.env.isWsl() {
		return false
	}
	wifiInfo, err := w.env.getWifiNetwork()
	displayError := w.props.getBool(DisplayError, false)
	if err != nil && displayError {
		w.Error = err.Error()
		return true
	}
	if err != nil || wifiInfo == nil {
		return false
	}
	w.wifiInfo = *wifiInfo
	return true
}

func (w *wifi) string() string {
	segmentTemplate := w.props.getString(SegmentTemplate, defaultTemplate)
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  w,
		Env:      w.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}

	return text
}

func (w *wifi) init(props properties, env environmentInfo) {
	w.props = props
	w.env = env
}
