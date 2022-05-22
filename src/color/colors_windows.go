package color

import (
	"syscall"
	"unsafe"

	"github.com/gookit/color"
	"golang.org/x/sys/windows"
)

var (
	dwmapi                      = syscall.NewLazyDLL("dwmapi.dll")
	procDwmGetColorizationColor = dwmapi.NewProc("DwmGetColorizationColor")
)

func (d *DefaultColors) SetAccentColor(defaultColor string) {
	var accentColor uint32
	var pfOpaqueBlend bool
	_, _, e := procDwmGetColorizationColor.Call(
		uintptr(unsafe.Pointer(&accentColor)),
		uintptr(unsafe.Pointer(&pfOpaqueBlend)))
	if e != windows.ERROR_SUCCESS {
		d.accent = &Color{
			Foreground: string(d.AnsiColorFromString(defaultColor, false)),
			Background: string(d.AnsiColorFromString(defaultColor, true)),
		}
		return
	}
	r := byte(accentColor >> 16)
	g := byte(accentColor >> 8)
	b := byte(accentColor)
	foreground := color.RGB(r, g, b, false)
	background := color.RGB(r, g, b, true)
	d.accent = &Color{
		Foreground: foreground.String(),
		Background: background.String(),
	}
}
