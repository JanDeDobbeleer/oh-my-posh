package color

import (
	"errors"
	"oh-my-posh/platform"

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
		d.accent = &Color{
			Foreground: string(d.AnsiColorFromString(defaultColor, false)),
			Background: string(d.AnsiColorFromString(defaultColor, true)),
		}
		return
	}
	foreground := color.RGB(rgb.R, rgb.G, rgb.B, false)
	background := color.RGB(rgb.R, rgb.G, rgb.B, true)
	d.accent = &Color{
		Foreground: foreground.String(),
		Background: background.String(),
	}
}
