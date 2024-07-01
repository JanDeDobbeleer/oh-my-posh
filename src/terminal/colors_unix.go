//go:build !windows

package terminal

import "github.com/jandedobbeleer/oh-my-posh/src/platform"

func GetAccentColor(_ platform.Environment) (*RGB, error) {
	return nil, &platform.NotImplemented{}
}

func (d *DefaultColors) SetAccentColor(_ platform.Environment, defaultColor string) {
	if len(defaultColor) == 0 {
		return
	}

	d.accent = &Colors{
		Foreground: string(d.ToColor(defaultColor, false)),
		Background: string(d.ToColor(defaultColor, true)),
	}
}
