package environment

import (
	"errors"
	"oh-my-posh/regex"
	"strconv"
	"strings"
	"time"

	"oh-my-posh/environment/battery"
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

func (env *ShellEnvironment) parseBatteryOutput(output string) (*battery.Info, error) {
	matches := regex.FindNamedRegexMatch(`(?P<PERCENTAGE>[0-9]{1,3})%; (?P<STATE>[a-zA-Z\s]+);`, output)
	if len(matches) != 2 {
		msg := "Unable to find battery state based on output"
		env.Log(Error, "BatteryState", msg)
		return nil, errors.New(msg)
	}
	var percentage int
	var err error
	if percentage, err = strconv.Atoi(matches["PERCENTAGE"]); err != nil {
		env.Log(Error, "BatteryState", err.Error())
		return nil, errors.New("Unable to parse battery percentage")
	}
	return &battery.Info{
		Percentage: percentage,
		State:      mapMostLogicalState(matches["STATE"]),
	}, nil
}

func (env *ShellEnvironment) BatteryState() (*battery.Info, error) {
	defer env.Trace(time.Now(), "BatteryState")
	output, err := env.RunCommand("pmset", "-g", "batt")
	if err != nil {
		env.Log(Error, "BatteryState", err.Error())
		return nil, err
	}
	if !strings.Contains(output, "Battery") {
		return nil, errors.New("No battery found")
	}
	return env.parseBatteryOutput(output)
}
