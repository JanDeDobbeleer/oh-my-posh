package platform

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/regex"

	"github.com/jandedobbeleer/oh-my-posh/src/platform/battery"
)

func mapMostLogicalState(state string) battery.State {
	switch state {
	case "charging":
		return battery.Charging
	case "discharging":
		return battery.Discharging
	case "AC attached":
		return battery.NotCharging
	case "full":
		return battery.Full
	case "empty":
		return battery.Empty
	case "charged":
		return battery.Full
	default:
		return battery.Unknown
	}
}

func (env *Shell) parseBatteryOutput(output string) (*battery.Info, error) {
	matches := regex.FindNamedRegexMatch(`(?P<PERCENTAGE>[0-9]{1,3})%; (?P<STATE>[a-zA-Z\s]+);`, output)
	if len(matches) != 2 {
		err := errors.New("Unable to find battery state based on output")
		env.Error("BatteryState", err)
		return nil, err
	}
	var percentage int
	var err error
	if percentage, err = strconv.Atoi(matches["PERCENTAGE"]); err != nil {
		env.Error("BatteryState", err)
		return nil, errors.New("Unable to parse battery percentage")
	}
	return &battery.Info{
		Percentage: percentage,
		State:      mapMostLogicalState(matches["STATE"]),
	}, nil
}

func (env *Shell) BatteryState() (*battery.Info, error) {
	defer env.Trace(time.Now(), "BatteryState")
	output, err := env.RunCommand("pmset", "-g", "batt")
	if err != nil {
		env.Error("BatteryState", err)
		return nil, err
	}
	if !strings.Contains(output, "Battery") {
		return nil, errors.New("No battery found")
	}
	return env.parseBatteryOutput(output)
}
