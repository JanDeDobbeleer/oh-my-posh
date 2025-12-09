package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/battery"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

type Battery struct {
	Base
	Error string
	Icon  string
	battery.Info
}

const (
	// ChargingIcon to display when charging
	ChargingIcon options.Option = "charging_icon"
	// DischargingIcon o display when discharging
	DischargingIcon options.Option = "discharging_icon"
	// ChargedIcon to display when fully charged
	ChargedIcon options.Option = "charged_icon"
	// NotChargingIcon to display when on AC power
	NotChargingIcon options.Option = "not_charging_icon"
)

func (b *Battery) Template() string {
	return " {{ if not .Error }}{{ .Icon }}{{ .Percentage }}{{ end }}{{ .Error }} "
}

func (b *Battery) Enabled() bool {
	// disable in WSL1
	if b.env.IsWsl() && !b.env.IsWsl2() {
		return false
	}

	info, err := b.env.BatteryState()

	if !b.enabledWhileError(err) {
		return false
	}

	b.Info = *info

	// case on computer without batteries(no error, empty array)
	if err == nil && b.Percentage == 0 {
		return false
	}

	switch b.State {
	case battery.Discharging:
		b.Icon = b.options.String(DischargingIcon, "")
	case battery.NotCharging:
		b.Icon = b.options.String(NotChargingIcon, "")
	case battery.Charging:
		b.Icon = b.options.String(ChargingIcon, "")
	case battery.Full:
		b.Icon = b.options.String(ChargedIcon, "")
	case battery.Empty, battery.Unknown:
		return true
	}
	return true
}

func (b *Battery) enabledWhileError(err error) bool {
	if err == nil {
		return true
	}
	if _, ok := err.(*battery.NoBatteryError); ok {
		return false
	}
	displayError := b.options.Bool(options.DisplayError, false)
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
