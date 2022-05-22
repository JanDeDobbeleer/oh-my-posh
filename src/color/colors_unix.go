//go:build !windows

package color

import "errors"

func GetAccentColor() (*RGB, error) {
	return nil, errors.New("not implemented")
}

func (d *DefaultColors) SetAccentColor(defaultColor string) {
	if len(defaultColor) == 0 {
		return
	}
	d.accent = &Color{
		Foreground: string(d.AnsiColorFromString(defaultColor, false)),
		Background: string(d.AnsiColorFromString(defaultColor, true)),
	}
}
