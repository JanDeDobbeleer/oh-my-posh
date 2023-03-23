//go:build !windows

package ansi

import "github.com/jandedobbeleer/oh-my-posh/src/platform"

func GetAccentColor(_ platform.Environment) (*RGB, error) {
	return nil, &platform.NotImplemented{}
}

func (d *DefaultColors) SetAccentColor(env platform.Environment, defaultColor string) {
	if len(defaultColor) == 0 {
		return
	}
	d.accent = &Colors{
		Foreground: string(d.ToColor(defaultColor, false, env.Flags().TrueColor)),
		Background: string(d.ToColor(defaultColor, true, env.Flags().TrueColor)),
	}
}
