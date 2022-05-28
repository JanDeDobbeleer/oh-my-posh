//go:build !windows

package color

import (
	"errors"
	"oh-my-posh/environment"
)

func GetAccentColor(env environment.Environment) (*RGB, error) {
	return nil, errors.New("not implemented")
}

func (d *DefaultColors) SetAccentColor(env environment.Environment, defaultColor string) {
	if len(defaultColor) == 0 {
		return
	}
	d.accent = &Color{
		Foreground: string(d.AnsiColorFromString(defaultColor, false)),
		Background: string(d.AnsiColorFromString(defaultColor, true)),
	}
}
