//go:build !windows

package color

import "github.com/jandedobbeleer/oh-my-posh/src/runtime"

func GetAccentColor(_ runtime.Environment) (*RGB, error) {
	return nil, &runtime.NotImplemented{}
}

func (d *Defaults) SetAccentColor(_ runtime.Environment, defaultColor Ansi) {
	if len(defaultColor) == 0 {
		return
	}

	d.accent = &Set{
		Foreground: d.ToAnsi(defaultColor, false),
		Background: d.ToAnsi(defaultColor, true),
	}
}
