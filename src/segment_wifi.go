package main

import (
	"regexp"
	"strings"
)

type wifi struct {
	props              *properties
	env                environmentInfo
	State              string
	SSID               string
	RadioType          string
	AuthenticationType string
	Channel            string
	ReceiveRate        string
	TransmitRate       string
	SignalStrength     string
}

const (
	State              = "State"
	SSID               = "SSID"
	RadioType          = "Radio type"
	AuthenticationType = "Authentication"
	Channel            = "Channel"
	ReceiveRate        = "Receive rate (Mbps)"
	TransmitRate       = "Transmit rate (Mbps)"
	SignalStrength     = "Signal"
)

func (w *wifi) enabled() bool {
	// If in Linux, who wnows. Gonna need a physical linux machine to develop this since the wlan interface isn't available in wsl
	// But possible solution is to cat /proc/net/wireless, and if that file doesn't exist then don't show the segment
	// In Windows and in wsl you can use the command `netsh.exe wlan show interfaces`. Note the .exe is important for this to worw in wsl.
	if w.env.getPlatform() == windowsPlatform || w.env.isWsl() {
		cmd := "netsh"
		if !w.env.hasCommand(cmd) {
			return false
		}
		cmdResult, err := w.env.runCommand(cmd, "wlan", "show", "interfaces")
		displayError := w.props.getBool(DisplayError, false)
		if err != nil && displayError {
			w.State = "WIFI ERR"
			// w.Namespace = w.Context
			return true
		}
		if err != nil {
			return false
		}

		regex := regexp.MustCompile(`(.+) : (.+)`)
		lines := strings.Split(cmdResult, "\n")
		for _, line := range lines[3 : len(lines)-3] {
			matches := regex.FindStringSubmatch(line)
			if len(matches) != 3 {
				continue
			}
			name := strings.TrimSpace(matches[1])
			value := strings.TrimSpace(matches[2])

			switch name {
			case State:
				w.State = value
			case SSID:
				w.SSID = value
			case RadioType:
				w.RadioType = value
			case AuthenticationType:
				w.AuthenticationType = value
			case Channel:
				w.Channel = value
			case ReceiveRate:
				w.ReceiveRate = strings.Split(value, ".")[0]
			case TransmitRate:
				w.TransmitRate = strings.Split(value, ".")[0]
			case SignalStrength:
				w.SignalStrength = value
			}
		}
	}

	return true
}

func (w *wifi) string() string {
	segmentTemplate := w.props.getString(SegmentTemplate, "{{.SSID}} {{.SignalStrength}} {{.ReceiveRate}}Mbps")
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

func (w *wifi) init(props *properties, env environmentInfo) {
	w.props = props
	w.env = env
}
