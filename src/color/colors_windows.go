package color

import (
	"errors"
	"syscall"
	"unsafe"

	"github.com/gookit/color"
	"golang.org/x/sys/windows"
)

var (
	dwmapi                      = syscall.NewLazyDLL("dwmapi.dll")
	procDwmGetColorizationColor = dwmapi.NewProc("DwmGetColorizationColor")
)

func GetAccentColor() (*RGB, error) {
	var accentColor uint32
	var pfOpaqueBlend bool
	_, _, e := procDwmGetColorizationColor.Call(
		uintptr(unsafe.Pointer(&accentColor)),
		uintptr(unsafe.Pointer(&pfOpaqueBlend)))
	if e != windows.ERROR_SUCCESS {
		return nil, errors.New("unable to get accent color")
	}
	return &RGB{
		R: byte(accentColor >> 16),
		G: byte(accentColor >> 8),
		B: byte(accentColor),
	}, nil
}

func (d *DefaultColors) SetAccentColor(defaultColor string) {
	rgb, err := GetAccentColor()
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
