//go:build !windows

package color

func (d *DefaultColors) SetAccentColor(defaultColor string) {
	if len(defaultColor) == 0 {
		return
	}
	d.accent = &Color{
		Foreground: string(d.AnsiColorFromString(defaultColor, false)),
		Background: string(d.AnsiColorFromString(defaultColor, true)),
	}
}
