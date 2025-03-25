package battery

import (
	"errors"
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/cmd"
)

func Get() (*Info, error) {
	output, err := cmd.Run("envstat", "-s", "acpibat0:charge", "-n")
	if err != nil {
		return nil, err
	}
	percentage, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil {
		return nil, errors.New("unable to parse battery percentage")
	}
	return &Info{
		Percentage: percentage,
		State:      Unknown,
	}, nil
}
