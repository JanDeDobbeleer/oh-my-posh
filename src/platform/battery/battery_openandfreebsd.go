//go:build openbsd || freebsd

package battery

import (
	"errors"
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/platform/cmd"
)

// See https://man.openbsd.org/man8/apm.8
func mapMostLogicalState(state string) State {
	switch state {
	case "3":
		return Charging
	case "0", "1":
		return Discharging
	case "2":
		return Empty
	default:
		return Unknown
	}
}

func parseBatteryOutput(apm_percentage string, apm_status string) (*Info, error) {
	percentage, err := strconv.Atoi(strings.TrimSpace(apm_percentage))
	if err != nil {
		return nil, errors.New("Unable to parse battery percentage")
	}

	if percentage == 100 {
		return &Info{
			Percentage: percentage,
			State:      Full,
		}, nil
	}

	return &Info{
		Percentage: percentage,
		State:      mapMostLogicalState(apm_status),
	}, nil
}

func Get() (*Info, error) {
	apm_percentage, err := cmd.Run("apm", "-l")
	if err != nil {
		return nil, err
	}

	apm_status, err := cmd.Run("apm", "-b")
	if err != nil {
		return nil, err
	}

	return parseBatteryOutput(apm_percentage, apm_status)

}
