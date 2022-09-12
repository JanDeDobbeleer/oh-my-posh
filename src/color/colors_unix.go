//go:build !windows

package color

import (
	"oh-my-posh/environment"
)

func GetAccentColor(env environment.Environment) (*RGB, error) {
	return nil, &environment.NotImplemented{}
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
