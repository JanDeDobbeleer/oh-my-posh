package battery

import (
	"errors"
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/platform/cmd"
	"github.com/jandedobbeleer/oh-my-posh/regex"
)

func mapMostLogicalState(state string) State {
	switch state {
	case "charging":
		return Charging
	case "discharging":
		return Discharging
	case "AC attached":
		return NotCharging
	case "full":
		return Full
	case "empty":
		return Empty
	case "charged":
		return Full
	default:
		return Unknown
	}
}

func parseBatteryOutput(output string) (*Info, error) {
	matches := regex.FindNamedRegexMatch(`(?P<PERCENTAGE>[0-9]{1,3})%; (?P<STATE>[a-zA-Z\s]+);`, output)
	if len(matches) != 2 {
		msg := "Unable to find battery state based on output"
		return nil, errors.New(msg)
	}
	var percentage int
	var err error
	if percentage, err = strconv.Atoi(matches["PERCENTAGE"]); err != nil {
		return nil, errors.New("Unable to parse battery percentage")
	}
	return &Info{
		Percentage: percentage,
		State:      mapMostLogicalState(matches["STATE"]),
	}, nil
}

func Get() (*Info, error) {
	output, err := cmd.Run("pmset", "-g", "batt")
	if err != nil {
		return nil, err
	}
	if !strings.Contains(output, "Battery") {
		return nil, ErrNotFound
	}
	return parseBatteryOutput(output)
}
