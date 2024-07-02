//go:build !windows

package color

import "github.com/jandedobbeleer/oh-my-posh/src/platform"

func GetAccentColor(_ platform.Environment) (*RGB, error) {
	return nil, &platform.NotImplemented{}
}

func (d *Defaults) SetAccentColor(_ platform.Environment, defaultColor Ansi) {
	if len(defaultColor) == 0 {
		return
	}

	d.accent = &Set{
		Foreground: d.ToAnsi(defaultColor, false),
		Background: d.ToAnsi(defaultColor, true),
	}
}
