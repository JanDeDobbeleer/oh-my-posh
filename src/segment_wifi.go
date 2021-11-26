package main

import (
	"fmt"
	"strconv"
	"strings"
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

	// Bail out of no netsh command found
	cmd := "netsh.exe"
	if !w.env.hasCommand(cmd) {
		return false
	}

	// Attempt to retrieve output from netsh command
	cmdResult, err := w.env.runCommand(cmd, "wlan", "show", "interfaces")
	displayError := w.props.getBool(DisplayError, false)
	if err != nil && displayError {
		w.State = fmt.Sprintf("WIFI ERR: %s", err)
		return true
	}
	if err != nil {
		return false
	}

	// Extract data from netsh cmdResult
	parseNetshCmdResult(cmdResult, w)

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

func parseNetshCmdResult(netshCmdResult string, w *wifi) {
	lines := strings.Split(netshCmdResult, "\n")
	for _, line := range lines {
		matches := strings.Split(line, " : ")
		if len(matches) != 2 {
			continue
		}
		name := strings.TrimSpace(matches[0])
		value := strings.TrimSpace(matches[1])

		switch name {
		case "State":
			w.State = value
			w.Connected = value == "connected"
		case "SSID":
			w.SSID = value
		case "Radio type":
			w.RadioType = value
		case "Authentication":
			w.Authentication = value
		case "Channel":
			if intValue, err := strconv.Atoi(value); err == nil {
				w.Channel = intValue
			}
		case "Receive rate (Mbps)":
			if intValue, err := strconv.Atoi(strings.Split(value, ".")[0]); err == nil {
				w.ReceiveRate = intValue
			}
		case "Transmit rate (Mbps)":
			if intValue, err := strconv.Atoi(strings.Split(value, ".")[0]); err == nil {
				w.TransmitRate = intValue
			}
		case "Signal":
			if intValue, err := strconv.Atoi(strings.TrimRight(value, "%")); err == nil {
				w.Signal = intValue
			}
		}
	}
}
