package main

import (
	"math"

	"github.com/distatus/battery"
)

type batt struct {
	props *properties
	env   environmentInfo

	battery.Battery
	Percentage int
	Error      string
	Icon       string
}

const (
	// ChargingIcon to display when charging
	ChargingIcon Property = "charging_icon"
	// DischargingIcon o display when discharging
	DischargingIcon Property = "discharging_icon"
	// ChargedIcon to display when fully charged
	ChargedIcon Property = "charged_icon"
)

func (b *batt) enabled() bool {
	batteries, err := b.env.getBatteryInfo()

	if !b.enabledWhileError(err) {
		return false
	}

	// case on computer without batteries(no error, empty array)
	if err == nil && len(batteries) == 0 {
		return false
	}

	for _, bt := range batteries {
		b.Battery.Current += bt.Current
		b.Battery.Full += bt.Full
		b.Battery.State = b.mapMostLogicalState(b.Battery.State, bt.State)
	}
	batteryPercentage := b.Battery.Current / b.Battery.Full * 100
	b.Percentage = int(math.Min(100, batteryPercentage))

	if !b.shouldDisplay() {
		return false
	}

	switch b.Battery.State {
	case battery.Discharging, battery.NotCharging:
		b.Icon = b.props.getString(DischargingIcon, "")
	case battery.Charging:
		b.Icon = b.props.getString(ChargingIcon, "")
	case battery.Full:
		b.Icon = b.props.getString(ChargedIcon, "")
	case battery.Empty, battery.Unknown:
		return true
	}
	b.colorSegment()
	return true
}

func (b *batt) enabledWhileError(err error) bool {
	if err == nil {
		return true
	}
	if _, ok := err.(*noBatteryError); ok {
		return false
	}
	displayError := b.props.getBool(DisplayError, false)
	if !displayError {
		return false
	}
	b.Error = err.Error()
	// On Windows, it sometimes errors when the battery is full.
	// This hack ensures we display a fully charged battery, even if
	// that state can be incorrect. It's better to "ignore" the error
	// than to not display the segment at all as that will confuse users.
	b.Battery.Current = 100
	b.Battery.Full = 10
	b.Battery.State = battery.Full
	return true
}

func (b *batt) mapMostLogicalState(currentState, newState battery.State) battery.State {
	switch currentState {
	case battery.Discharging, battery.NotCharging:
		return battery.Discharging
	case battery.Empty:
		return newState
	case battery.Charging:
		if newState == battery.Discharging {
			return battery.Discharging
		}
		return battery.Charging
	case battery.Unknown:
		return newState
	case battery.Full:
		return newState
	}
	return newState
}

func (b *batt) string() string {
	segmentTemplate := b.props.getString(SegmentTemplate, "{{ if not .Error }}{{.Icon}}{{.Percentage}}{{ end }}{{.Error}}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  b,
		Env:      b.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (b *batt) init(props *properties, env environmentInfo) {
	b.props = props
	b.env = env
}
