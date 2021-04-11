package main

import (
	"math"

	"github.com/distatus/battery"
)

type batt struct {
	props      *properties
	env        environmentInfo
	Battery    *battery.Battery
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
	// ChargedColor to display when fully charged
	ChargedColor Property = "charged_color"
	// ChargingColor to display when charging
	ChargingColor Property = "charging_color"
	// DischargingColor to display when discharging
	DischargingColor Property = "discharging_color"
	// DisplayCharging Hide the battery icon while it's charging
	DisplayCharging Property = "display_charging"
)

func (b *batt) enabled() bool {
	var err error
	b.Battery, err = b.env.getBatteryInfo()

	displayError := b.props.getBool(DisplayError, false)
	if err != nil && displayError {
		b.Error = err.Error()
		return true
	}
	if err != nil {
		// On Windows, it sometimes errors when the battery is full.
		// This hack ensures we display a fully charged battery, even if
		// that state can be incorrect. It's better to "ignore" the error
		// than to not display the segment at all as that will confuse users.
		b.Battery = &battery.Battery{
			Current: 100,
			Full:    100,
			State:   battery.Full,
		}
	}

	display := b.props.getBool(DisplayCharging, true)
	if !display && (b.Battery.State == battery.Charging || b.Battery.State == battery.Full) {
		return false
	}

	batteryPercentage := b.Battery.Current / b.Battery.Full * 100
	b.Percentage = int(math.Min(100, batteryPercentage))
	var colorPorperty Property
	switch b.Battery.State {
	case battery.Discharging:
		colorPorperty = DischargingColor
		b.Icon = b.props.getString(DischargingIcon, "")
	case battery.Charging:
		colorPorperty = ChargingColor
		b.Icon = b.props.getString(ChargingIcon, "")
	case battery.Full:
		colorPorperty = ChargedColor
		b.Icon = b.props.getString(ChargedIcon, "")
	case battery.Empty, battery.Unknown:
		return true
	}
	colorBackground := b.props.getBool(ColorBackground, false)
	if colorBackground {
		b.props.background = b.props.getColor(colorPorperty, b.props.background)
	} else {
		b.props.foreground = b.props.getColor(colorPorperty, b.props.foreground)
	}
	return true
}

func (b *batt) string() string {
	segmentTemplate := b.props.getString(SegmentTemplate, "{{.Icon}}{{ if not .Error }}{{.Percentage}}{{ end }}{{.Error}}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  b,
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
