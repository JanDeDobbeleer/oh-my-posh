package ansi

import (
	"errors"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"

	"github.com/gookit/color"
)

func GetAccentColor(env platform.Environment) (*RGB, error) {
	if env == nil {
		return nil, errors.New("unable to get color without environment")
	}
	// see https://stackoverflow.com/questions/3560890/vista-7-how-to-get-glass-color
	value, err := env.WindowsRegistryKeyValue(`HKEY_CURRENT_USER\Software\Microsoft\Windows\DWM\ColorizationColor`)
	if err != nil || value.ValueType != platform.DWORD {
		return nil, err
	}
	return &RGB{
		R: byte(value.DWord >> 16),
		G: byte(value.DWord >> 8),
		B: byte(value.DWord),
	}, nil
}

func (d *DefaultColors) SetAccentColor(env platform.Environment, defaultColor string) {
	rgb, err := GetAccentColor(env)
	if err != nil {
		d.accent = &Colors{
			Foreground: string(d.ToColor(defaultColor, false, env.Flags().TrueColor)),
			Background: string(d.ToColor(defaultColor, true, env.Flags().TrueColor)),
		}
		return
	}
	foreground := color.RGB(rgb.R, rgb.G, rgb.B, false)
	background := color.RGB(rgb.R, rgb.G, rgb.B, true)
	d.accent = &Colors{
		Foreground: foreground.String(),
		Background: background.String(),
	}
}
