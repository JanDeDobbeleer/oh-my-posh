package color

import (
	"errors"
	"strconv"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

func GetAccentColor(env runtime.Environment) (*RGB, error) {
	output, err := env.RunCommand("defaults", "read", "-g", "AppleAccentColor")
	if err != nil {
		log.Error(err)
		return nil, errors.New("unable to read accent color")
	}

	index, err := strconv.Atoi(output)
	if err != nil {
		log.Error(err)
		return nil, errors.New("unable to parse accent color index")
	}

	var accentColors = map[int]RGB{
		-1: {152, 152, 152}, // Graphite
		0:  {224, 55, 62},   // Red
		1:  {247, 130, 25},  // Orange
		2:  {255, 199, 38},  // Yellow
		3:  {96, 186, 70},   // Green
		4:  {0, 122, 255},   // Blue
		5:  {149, 61, 150},  // Purple
		6:  {247, 79, 159},  // Pink
	}

	color, exists := accentColors[index]
	if !exists {
		color = accentColors[6] // Default to graphite (white)
	}

	return &color, nil
}
