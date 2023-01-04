//go:build !windows

package ansi

import "github.com/jandedobbeleer/oh-my-posh/platform"

func GetAccentColor(env platform.Environment) (*RGB, error) {
	return nil, &platform.NotImplemented{}
}

func (d *DefaultColors) SetAccentColor(env platform.Environment, defaultColor string) {
	if len(defaultColor) == 0 {
		return
	}
	d.accent = &cachedColor{
		Foreground: string(d.AnsiColorFromString(defaultColor, false)),
		Background: string(d.AnsiColorFromString(defaultColor, true)),
	}
}
