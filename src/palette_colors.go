package main

type PaletteColors struct {
	ansiColors AnsiColors
	palette    Palette
}

func (p *PaletteColors) AnsiColorFromString(colorString string, isBackground bool) AnsiColor {
	paletteColor, err := p.palette.ResolveColor(colorString)
	if err != nil {
		return emptyAnsiColor
	}
	ansiColor := p.ansiColors.AnsiColorFromString(paletteColor, isBackground)
	return ansiColor
}
