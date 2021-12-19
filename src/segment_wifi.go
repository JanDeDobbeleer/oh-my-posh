package main

import (
	"fmt"
)

type wifi struct {
	props          properties
	env            environmentInfo
	Connected      bool
	State          string
	SSID           string
	RadioType      string
	Authentication string
	Channel        int
	ReceiveRate    int
	TransmitRate   int
	Signal         int
}

const (
	defaultTemplate = "{{ if .Connected }}\uFAA8{{ else }}\uFAA9{{ end }}{{ if .Connected }}{{ .SSID }} {{ .Signal }}% {{ .ReceiveRate }}Mbps{{ else }}{{ .State }}{{ end }}"
)

func (w *wifi) enabled() bool {
	// This segment only supports Windows/WSL for now
	if w.env.getPlatform() != windowsPlatform && !w.env.isWsl() {
		return false
	}
	wifiInfo, err := w.env.getWifiNetworks()
	displayError := w.props.getBool(DisplayError, false)
	if err != nil && displayError {
		w.State = fmt.Sprintf("WIFI ERR: %s", err)
		return true
	}
	if err != nil {
		return false
	}

	if wifiInfo == nil {
		return false
	}

	w.Connected = wifiInfo.Connected
	w.State = wifiInfo.State
	w.SSID = wifiInfo.SSID
	w.RadioType = wifiInfo.BSSType
	w.Authentication = wifiInfo.Auth
	w.Channel = wifiInfo.Channel
	w.ReceiveRate = wifiInfo.ReceiveRate / 1024
	w.TransmitRate = wifiInfo.TransmitRate / 1024
	w.Signal = wifiInfo.Signal

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
