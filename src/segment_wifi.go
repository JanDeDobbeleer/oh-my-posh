package main

import "net"

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
	interfaces, err := net.Interfaces()
	if err != nil {
		return "WiFi failed"
	}
	// https://github.com/skycoin/examples/blob/af669cb42b6cc73e7fef5198fc9b5c4b6afce357/aether/wifi/wifi.go
	return "WiFi"
}

func (w *wifi) init(props *properties, env environmentInfo) {
	w.props = props
	w.env = env
}
