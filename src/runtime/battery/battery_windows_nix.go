//go:build !darwin && !netbsd && !openbsd && !freebsd && !js && !illumos && !solaris

package battery

import (
	"math"
)

type battery struct {
	// Current battery state.
	State State
	// Current (momentary) capacity (in mWh).
	Current float64
	// Last known full capacity (in mWh).
	Full float64
	// Current voltage (in V).
	Voltage float64
}

func mapMostLogicalState(currentState, newState State) State {
	switch currentState {
	case Discharging, NotCharging:
		return Discharging
	case Empty:
		return newState
	case Charging:
		if newState == Discharging {
			return Discharging
		}
		return Charging
	case Unknown:
		return newState
	case Full:
		return newState
	}
	return newState
}

func Get() (*Info, error) {
	parseBatteryInfo := func(batteries []*battery) *Info {
		var info Info
		var current, total float64
		var state State
		for _, bt := range batteries {
			current += bt.Current
			total += bt.Full
			state = mapMostLogicalState(state, bt.State)
		}
		batteryPercentage := current / total * 100
		info.Percentage = int(math.Min(100, batteryPercentage))
		info.State = state
		return &info
	}

	batteries, err := systemGetAll()
	if err != nil {
		return nil, err
	}
	return parseBatteryInfo(batteries), nil
}
