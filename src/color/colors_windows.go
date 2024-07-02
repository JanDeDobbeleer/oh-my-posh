package color

import (
	"errors"

	"github.com/gookit/color"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

func GetAccentColor(env runtime.Environment) (*RGB, error) {
	if env == nil {
		return nil, errors.New("unable to get color without environment")
	}

	// see https://stackoverflow.com/questions/3560890/vista-7-how-to-get-glass-color
	value, err := env.WindowsRegistryKeyValue(`HKEY_CURRENT_USER\Software\Microsoft\Windows\DWM\ColorizationColor`)
	if err != nil || value.ValueType != runtime.DWORD {
		return nil, err
	}

	return &RGB{
		R: byte(value.DWord >> 16),
		G: byte(value.DWord >> 8),
		B: byte(value.DWord),
	}, nil
}

func (d *Defaults) SetAccentColor(env runtime.Environment, defaultColor Ansi) {
	rgb, err := GetAccentColor(env)
	if err != nil {
		d.accent = &Set{
			Foreground: d.ToAnsi(defaultColor, false),
			Background: d.ToAnsi(defaultColor, true),
		}

		return
	}

	foreground := color.RGB(rgb.R, rgb.G, rgb.B, false)
	background := color.RGB(rgb.R, rgb.G, rgb.B, true)

	d.accent = &Set{
		Foreground: Ansi(foreground.String()),
		Background: Ansi(background.String()),
	}
}
