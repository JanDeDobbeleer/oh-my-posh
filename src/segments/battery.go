package segments

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"

	"github.com/distatus/battery"
)

type Battery struct {
	props properties.Properties
	env   environment.Environment

	*environment.BatteryInfo
	Error string
	Icon  string
}

const (
	// ChargingIcon to display when charging
	ChargingIcon properties.Property = "charging_icon"
	// DischargingIcon o display when discharging
	DischargingIcon properties.Property = "discharging_icon"
	// ChargedIcon to display when fully charged
	ChargedIcon properties.Property = "charged_icon"
	// NotChargingIcon to display when on AC power
	NotChargingIcon properties.Property = "not_charging_icon"
)

func (b *Battery) Template() string {
	return " {{ if not .Error }}{{ .Icon }}{{ .Percentage }}{{ end }}{{ .Error }} "
}

func (b *Battery) Enabled() bool {
	// disable in WSL1
	if b.env.IsWsl() && !b.env.IsWsl2() {
		return false
	}

	var err error
	b.BatteryInfo, err = b.env.BatteryState()

	if !b.enabledWhileError(err) {
		return false
	}

	// case on computer without batteries(no error, empty array)
	if err == nil && b.BatteryInfo == nil {
		return false
	}

	switch b.BatteryInfo.State {
	case battery.Discharging:
		b.Icon = b.props.GetString(DischargingIcon, "")
	case battery.NotCharging:
		b.Icon = b.props.GetString(NotChargingIcon, "")
	case battery.Charging:
		b.Icon = b.props.GetString(ChargingIcon, "")
	case battery.Full:
		b.Icon = b.props.GetString(ChargedIcon, "")
	case battery.Empty, battery.Unknown:
		return true
	}
	return true
}

func (b *Battery) enabledWhileError(err error) bool {
	if err == nil {
		return true
	}
	if _, ok := err.(*environment.NoBatteryError); ok {
		return false
	}
	displayError := b.props.GetBool(properties.DisplayError, false)
	if !displayError {
		return false
	}
	b.Error = err.Error()
	// On Windows, it sometimes errors when the battery is full.
	// This hack ensures we display a fully charged battery, even if
	// that state can be incorrect. It's better to "ignore" the error
	// than to not display the segment at all as that will confuse users.
	b.Percentage = 100
	b.State = battery.Full
	return true
}

func (b *Battery) Init(props properties.Properties, env environment.Environment) {
	b.props = props
	b.env = env
}
