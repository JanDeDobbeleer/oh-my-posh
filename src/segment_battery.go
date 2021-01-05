package main

import (
	"fmt"
	"math"

	"github.com/distatus/battery"
)

type batt struct {
	props          *properties
	env            environmentInfo
	percentageText string
}

const (
	// BatteryIcon to display in front of the battery
	BatteryIcon Property = "battery_icon"
	// DisplayError to display when an error occurs or not
	DisplayError Property = "display_error"
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
	bt, err := b.env.getBatteryInfo()

	displayError := b.props.getBool(DisplayError, false)
	if err != nil && displayError {
		b.percentageText = "BATT ERR"
		return true
	}
	if err != nil {
		// On Windows, it sometimes errors when the battery is full.
		// This hack ensures we display a fully charged battery, even if
		// that state can be incorrect. It's better to "ignore" the error
		// than to not display the segment at all as that will confuse users.
		bt = &battery.Battery{
			Current: 100,
			Full:    100,
			State:   battery.Full,
		}
	}

	display := b.props.getBool(DisplayCharging, true)
	if !display && (bt.State == battery.Charging || bt.State == battery.Full) {
		return false
	}

	batteryPercentage := bt.Current / bt.Full * 100
	batteryPercentage = math.Min(100, batteryPercentage)
	percentageText := fmt.Sprintf("%.0f", batteryPercentage)
	var icon string
	var colorPorperty Property
	switch bt.State {
	case battery.Discharging:
		colorPorperty = DischargingColor
		icon = b.props.getString(DischargingIcon, "")
	case battery.Charging:
		colorPorperty = ChargingColor
		icon = b.props.getString(ChargingIcon, "")
	case battery.Full:
		colorPorperty = ChargedColor
		icon = b.props.getString(ChargedIcon, "")
	case battery.Empty, battery.Unknown:
		b.percentageText = percentageText
		return true
	}
	colorBackground := b.props.getBool(ColorBackground, false)
	if colorBackground {
		b.props.background = b.props.getColor(colorPorperty, b.props.background)
	} else {
		b.props.foreground = b.props.getColor(colorPorperty, b.props.foreground)
	}
	batteryIcon := b.props.getString(BatteryIcon, "")
	b.percentageText = fmt.Sprintf("%s%s%s", icon, batteryIcon, percentageText)
	return true
}

func (b *batt) string() string {
	return b.percentageText
}

func (b *batt) init(props *properties, env environmentInfo) {
	b.props = props
	b.env = env
}
