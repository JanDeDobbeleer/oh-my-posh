package main

import (
	"fmt"

	"github.com/distatus/battery"
)

type batt struct {
	props *properties
	env   environmentInfo
}

const (
	//BatteryIcon to display in front of the battery
	BatteryIcon Property = "battery_icon"
	//ChargingIcon to display when charging
	ChargingIcon Property = "charging_icon"
	//DischargingIcon o display when discharging
	DischargingIcon Property = "discharging_icon"
	//ChargedIcon to display when fully charged
	ChargedIcon Property = "charged_icon"
	//ChargedColor to display when fully charged
	ChargedColor Property = "charged_color"
	//ChargingColor to display when charging
	ChargingColor Property = "charging_color"
	//DischargingColor to display when discharging
	DischargingColor Property = "discharging_color"
)

func (b *batt) enabled() bool {
	return true
}

func (b *batt) string() string {
	bt, err := b.env.getBatteryInfo()
	if err != nil {
		return "BATT ERR"
	}
	batteryPercentage := bt.Current / bt.Full * 100
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
	default:
		return percentageText
	}
	colorBackground := b.props.getBool(ColorBackground, false)
	if colorBackground {
		b.props.background = b.props.getColor(colorPorperty, b.props.background)
	} else {
		b.props.foreground = b.props.getColor(colorPorperty, b.props.foreground)
	}
	batteryIcon := b.props.getString(BatteryIcon, "")
	return fmt.Sprintf("%s%s%s", icon, batteryIcon, percentageText)
}

func (b *batt) init(props *properties, env environmentInfo) {
	b.props = props
	b.env = env
}
