// Copyright © 2020 The Homeport Team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// https://github.com/homeport/termshot

package main

import (
	_ "embed"
	"fmt"
	"math"
	"oh-my-posh/color"
	"oh-my-posh/regex"
	"strconv"
	"strings"

	"github.com/esimov/stackblur-go"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

const (
	red    = "#ED655A"
	yellow = "#E1C04C"
	green  = "#71BD47"

	// known ansi sequences

	fg                  = "FG"
	bg                  = "BG"
	str                 = "STR"
	url                 = "URL"
	invertedColor       = "inverted"
	invertedColorSingle = "invertedsingle"
	fullColor           = "full"
	foreground          = "foreground"
	reset               = "reset"
	bold                = "bold"
	boldReset           = "boldr"
	italic              = "italic"
	italicReset         = "italicr"
	underline           = "underline"
	underlineReset      = "underliner"
	strikethrough       = "strikethrough"
	strikethroughReset  = "strikethroughr"
	color16             = "color16"
	left                = "left"
	osc99               = "osc99"
	lineChange          = "linechange"
	consoleTitle        = "title"
	link                = "link"
)

//go:embed font/Hack-Nerd-Bold.ttf
var hackBold []byte

//go:embed font/Hack-Nerd-Regular.ttf
var hackRegular []byte

//go:embed font/Hack-Nerd-Italic.ttf
var hackItalic []byte

type RGB struct {
	r int
	g int
	b int
}

func NewRGBColor(ansiColor string) *RGB {
	colors := strings.Split(ansiColor, ";")
	b, _ := strconv.Atoi(colors[2])
	g, _ := strconv.Atoi(colors[1])
	r, _ := strconv.Atoi(colors[0])
	return &RGB{
		r: r,
		g: g,
		b: b,
	}
}

type ImageRenderer struct {
	ansiString string
	author     string
	ansi       *color.Ansi
	bgColor    string

	factor float64

	columns int
	rows    int

	defaultForegroundColor *RGB
	defaultBackgroundColor *RGB

	shadowBaseColor string
	shadowRadius    uint8
	shadowOffsetX   float64
	shadowOffsetY   float64

	padding float64
	margin  float64

	regular     font.Face
	bold        font.Face
	italic      font.Face
	lineSpacing float64

	// canvas switches
	style                string
	backgroundColor      *RGB
	foregroundColor      *RGB
	ansiSequenceRegexMap map[string]string
	rPromptOffset        int
	cursorPadding        int
}

func (ir *ImageRenderer) init() {
	f := 2.0

	ir.cleanContent()

	fontRegular, _ := truetype.Parse(hackRegular)
	fontBold, _ := truetype.Parse(hackBold)
	fontItalic, _ := truetype.Parse(hackItalic)
	fontFaceOptions := &truetype.Options{Size: f * 12, DPI: 144}

	ir.defaultForegroundColor = &RGB{255, 255, 255}
	ir.defaultBackgroundColor = &RGB{21, 21, 21}

	ir.factor = f

	ir.columns = 80
	ir.rows = 25

	ir.margin = f * 48
	ir.padding = f * 24

	ir.shadowBaseColor = "#10101066"
	ir.shadowRadius = uint8(math.Min(f*16, 255))
	ir.shadowOffsetX = f * 16
	ir.shadowOffsetY = f * 16

	ir.regular = truetype.NewFace(fontRegular, fontFaceOptions)
	ir.bold = truetype.NewFace(fontBold, fontFaceOptions)
	ir.italic = truetype.NewFace(fontItalic, fontFaceOptions)
	ir.lineSpacing = 1.2

	ir.ansiSequenceRegexMap = map[string]string{
		invertedColor:       `^(?P<STR>(\x1b\[38;2;(?P<BG>(\d+;?){3});49m){1}(\x1b\[7m))`,
		invertedColorSingle: `^(?P<STR>\x1b\[(?P<BG>\d{2,3});49m\x1b\[7m)`,
		fullColor:           `^(?P<STR>(\x1b\[48;2;(?P<BG>(\d+;?){3})m)(\x1b\[38;2;(?P<FG>(\d+;?){3})m))`,
		foreground:          `^(?P<STR>(\x1b\[38;2;(?P<FG>(\d+;?){3})m))`,
		reset:               `^(?P<STR>\x1b\[0m)`,
		bold:                `^(?P<STR>\x1b\[1m)`,
		boldReset:           `^(?P<STR>\x1b\[22m)`,
		italic:              `^(?P<STR>\x1b\[3m)`,
		italicReset:         `^(?P<STR>\x1b\[23m)`,
		underline:           `^(?P<STR>\x1b\[4m)`,
		underlineReset:      `^(?P<STR>\x1b\[24m)`,
		strikethrough:       `^(?P<STR>\x1b\[9m)`,
		strikethroughReset:  `^(?P<STR>\x1b\[29m)`,
		color16:             `^(?P<STR>\x1b\[(?P<FG>\d{2,3})m)`,
		left:                `^(?P<STR>\x1b\[(\d{1,3})D)`,
		osc99:               `^(?P<STR>\x1b\]9;9;(.+)\x1b\\)`,
		lineChange:          `^(?P<STR>\x1b\[(\d)[FB])`,
		consoleTitle:        `^(?P<STR>\x1b\]0;(.+)\007)`,
		link:                `^(?P<STR>\x1b]8;;file:\/\/(.+)\x1b\\(?P<URL>.+)\x1b]8;;\x1b\\)`,
	}
}

func (ir *ImageRenderer) fontHeight() float64 {
	return float64(ir.regular.Metrics().Height >> 6)
}

type RuneRange struct {
	Start rune
	End   rune
}

// If we're a Nerd Font code point, treat as double width
var doubleWidthRunes = []RuneRange{
	// Seti-UI + Custom range
	{Start: '\ue5fa', End: '\ue62b'},
	// Devicons
	{Start: '\ue700', End: '\ue7c5'},
	// Font Awesome
	{Start: '\uf000', End: '\uf2e0'},
	// Font Awesome Extension
	{Start: '\ue200', End: '\ue2a9'},
	// Material Design Icons
	{Start: '\uf500', End: '\ufd46'},
	// Weather
	{Start: '\ue300', End: '\ue3eb'},
	// Octicons
	{Start: '\uf400', End: '\uf4a8'},
	{Start: '\u2665', End: '\u2665'},
	{Start: '\u26A1', End: '\u26A1'},
	{Start: '\uf27c', End: '\uf27c'},
	// Powerline Extra Symbols (intentionally excluding single width bubbles (e0b4-e0b7) and pixelated (e0c4-e0c7))
	{Start: '\ue0a3', End: '\ue0a3'},
	{Start: '\ue0b8', End: '\ue0c3'},
	{Start: '\ue0c8', End: '\ue0c8'},
	{Start: '\ue0ca', End: '\ue0ca'},
	{Start: '\ue0cc', End: '\ue0d2'},
	{Start: '\ue0d4', End: '\ue0d4'},
	// IEC Power Symbols
	{Start: '\u23fb', End: '\u23fe'},
	{Start: '\u2b58', End: '\u2b58'},
	// Font Logos
	{Start: '\uf300', End: '\uf31c'},
	// Pomicons
	{Start: '\ue000', End: '\ue00d'},
}

// This is getting how many additional characters of width to allocate when drawing
// e.g. for characters that are 2 or more wide. A standard character will return 0
// Nerd Font glyphs will return 1, since most are double width
func (ir *ImageRenderer) runeAdditionalWidth(r rune) int {
	for _, runeRange := range doubleWidthRunes {
		if runeRange.Start <= r && r <= runeRange.End {
			return 1
		}
	}
	return 0
}

func (ir *ImageRenderer) calculateWidth() int {
	longest := 0
	for _, line := range strings.Split(ir.ansiString, "\n") {
		length := ir.ansi.LenWithoutANSI(line)
		for _, char := range line {
			length += ir.runeAdditionalWidth(char)
		}
		if length > longest {
			longest = length
		}
	}
	return longest
}

func (ir *ImageRenderer) cleanContent() {
	rPromptAnsi := "\x1b7\x1b[1000C"
	hasRPrompt := strings.Contains(ir.ansiString, rPromptAnsi)
	// clean abundance of empty lines
	ir.ansiString = strings.Trim(ir.ansiString, "\n")
	ir.ansiString = "\n" + ir.ansiString
	// clean string before render
	ir.ansiString = strings.ReplaceAll(ir.ansiString, "\x1b[m", "\x1b[0m")
	ir.ansiString = strings.ReplaceAll(ir.ansiString, "\x1b[K", "")
	ir.ansiString = strings.ReplaceAll(ir.ansiString, "\x1b[1F", "")
	ir.ansiString = strings.ReplaceAll(ir.ansiString, "\x1b8", "")
	ir.ansiString = strings.ReplaceAll(ir.ansiString, "\u2800", " ")
	// replace rprompt with adding and mark right aligned blocks with a pointer
	ir.ansiString = strings.ReplaceAll(ir.ansiString, rPromptAnsi, fmt.Sprintf("_%s", strings.Repeat(" ", ir.cursorPadding)))
	ir.ansiString = strings.ReplaceAll(ir.ansiString, "\x1b[1000C", strings.Repeat(" ", ir.rPromptOffset))
	if !hasRPrompt {
		ir.ansiString += fmt.Sprintf("_%s", strings.Repeat(" ", ir.cursorPadding))
	}
	// add watermarks
	ir.ansiString += "\n\n\x1b[1mhttps://ohmyposh.dev\x1b[22m"
	if len(ir.author) > 0 {
		createdBy := fmt.Sprintf(" by \x1b[1m%s\x1b[22m", ir.author)
		ir.ansiString += createdBy
	}
}

func (ir *ImageRenderer) measureContent() (width, height float64) {
	// get the longest line
	linewidth := ir.calculateWidth()
	// width, taken from the longest line
	tmpDrawer := &font.Drawer{Face: ir.regular}
	advance := tmpDrawer.MeasureString(strings.Repeat(" ", linewidth))
	width = float64(advance >> 6)
	// height, lines times font height and line spacing
	height = float64(len(strings.Split(ir.ansiString, "\n"))) * ir.fontHeight() * ir.lineSpacing
	return width, height
}

func (ir *ImageRenderer) SavePNG(path string) error {
	var f = func(value float64) float64 { return ir.factor * value }

	var (
		corner   = f(6)
		radius   = f(9)
		distance = f(25)
	)

	contentWidth, contentHeight := ir.measureContent()

	// Make sure the output window is big enough in case no content or very few
	// content will be rendered
	contentWidth = math.Max(contentWidth, 3*distance+3*radius)

	marginX, marginY := ir.margin, ir.margin
	paddingX, paddingY := ir.padding, ir.padding

	xOffset := marginX
	yOffset := marginY
	titleOffset := f(40)

	width := contentWidth + 2*marginX + 2*paddingX
	height := contentHeight + 2*marginY + 2*paddingY + titleOffset

	dc := gg.NewContext(int(width), int(height))

	xOffset -= ir.shadowOffsetX / 2
	yOffset -= ir.shadowOffsetY / 2

	bc := gg.NewContext(int(width), int(height))
	bc.DrawRoundedRectangle(xOffset+ir.shadowOffsetX, yOffset+ir.shadowOffsetY, width-2*marginX, height-2*marginY, corner)
	bc.SetHexColor(ir.shadowBaseColor)
	bc.Fill()

	// var done = make(chan struct{}, ir.shadowRadius)
	shadow, err := stackblur.Run(
		bc.Image(),
		uint32(ir.shadowRadius),
	)

	if err != nil {
		return err
	}

	// <-done
	dc.DrawImage(shadow, 0, 0)

	// Draw rounded rectangle with outline and three button to produce the
	// impression of a window with controls and a content area
	dc.DrawRoundedRectangle(xOffset, yOffset, width-2*marginX, height-2*marginY, corner)
	dc.SetHexColor(ir.bgColor)
	dc.Fill()

	dc.DrawRoundedRectangle(xOffset, yOffset, width-2*marginX, height-2*marginY, corner)
	dc.SetHexColor("#404040")
	dc.SetLineWidth(f(1))
	dc.Stroke()

	for i, color := range []string{red, yellow, green} {
		dc.DrawCircle(xOffset+paddingX+float64(i)*distance+f(4), yOffset+paddingY+f(4), radius)
		dc.SetHexColor(color)
		dc.Fill()
	}

	// Apply the actual text into the prepared content area of the window
	var x, y float64 = xOffset + paddingX, yOffset + paddingY + titleOffset + ir.fontHeight()

	for len(ir.ansiString) != 0 {
		if !ir.shouldPrint() {
			continue
		}
		runes := []rune(ir.ansiString)
		if len(runes) == 0 {
			continue
		}
		str := string(runes[0:1])
		ir.ansiString = string(runes[1:])
		switch ir.style {
		case bold:
			dc.SetFontFace(ir.bold)
		case italic:
			dc.SetFontFace(ir.italic)
		default:
			dc.SetFontFace(ir.regular)
		}

		w, h := dc.MeasureString(str)
		// The gg library unfortunately returns a single character width for *all* glyphs in a font.
		// So if we know the glyph to occupy n additional characters in width, allocate that area
		// e.g. this will double the space for Nerd Fonts, but some could even be 3 or 4 wide
		// If there's 0 additional characters of width (the common case), this won't add anything
		w += (w * float64(ir.runeAdditionalWidth(runes[0])))

		if ir.backgroundColor != nil {
			dc.SetRGB255(ir.backgroundColor.r, ir.backgroundColor.g, ir.backgroundColor.b)
			// The background for a character needs love to align to the font we're using
			// Not all fonts are rendered the same height or starting position,
			// so we're shifting the background rectangles vertically to correct
			dc.DrawRectangle(x, y-h+3, w, h+9)
			dc.Fill()
		}
		if ir.foregroundColor != nil {
			dc.SetRGB255(ir.foregroundColor.r, ir.foregroundColor.g, ir.foregroundColor.b)
		} else {
			dc.SetRGB255(ir.defaultForegroundColor.r, ir.defaultForegroundColor.g, ir.defaultForegroundColor.b)
		}

		if str == "\n" {
			x = xOffset + paddingX
			y += h * ir.lineSpacing
			continue
		}

		dc.DrawString(str, x, y)

		if ir.style == underline {
			dc.DrawLine(x, y+f(4), x+w, y+f(4))
			dc.SetLineWidth(f(1))
			dc.Stroke()
		}

		x += w
	}

	return dc.SavePNG(path)
}

func (ir *ImageRenderer) shouldPrint() bool {
	for sequence, re := range ir.ansiSequenceRegexMap {
		match := regex.FindNamedRegexMatch(re, ir.ansiString)
		if len(match) == 0 {
			continue
		}
		ir.ansiString = strings.TrimPrefix(ir.ansiString, match[str])
		switch sequence {
		case invertedColor:
			ir.foregroundColor = ir.defaultBackgroundColor
			ir.backgroundColor = NewRGBColor(match[bg])
			return false
		case invertedColorSingle:
			ir.foregroundColor = ir.defaultBackgroundColor
			bgColor, _ := strconv.Atoi(match[bg])
			bgColor += 10
			ir.setBase16Color(fmt.Sprint(bgColor))
			return false
		case fullColor:
			ir.foregroundColor = NewRGBColor(match[fg])
			ir.backgroundColor = NewRGBColor(match[bg])
			return false
		case foreground:
			ir.foregroundColor = NewRGBColor(match[fg])
			return false
		case reset:
			ir.foregroundColor = ir.defaultForegroundColor
			ir.backgroundColor = nil
			return false
		case bold, italic, underline:
			ir.style = sequence
			return false
		case boldReset, italicReset, underlineReset:
			ir.style = ""
			return false
		case strikethrough, strikethroughReset, left, osc99, lineChange, consoleTitle:
			return false
		case color16:
			ir.setBase16Color(match[fg])
			return false
		case link:
			ir.ansiString = match[url] + ir.ansiString
		}
	}
	return true
}

func (ir *ImageRenderer) setBase16Color(colorStr string) {
	tempColor := ir.defaultForegroundColor
	colorInt, err := strconv.Atoi(colorStr)
	if err != nil {
		ir.foregroundColor = tempColor
	}
	switch colorInt {
	case 30, 40: // Black
		tempColor = &RGB{1, 1, 1}
	case 31, 41: // Red
		tempColor = &RGB{222, 56, 43}
	case 32, 42: // Green
		tempColor = &RGB{57, 181, 74}
	case 33, 43: // Yellow
		tempColor = &RGB{255, 199, 6}
	case 34, 44: // Blue
		tempColor = &RGB{0, 111, 184}
	case 35, 45: // Magenta
		tempColor = &RGB{118, 38, 113}
	case 36, 46: // Cyan
		tempColor = &RGB{44, 181, 233}
	case 37, 47: // White
		tempColor = &RGB{204, 204, 204}
	case 90, 100: // Bright Black (Gray)
		tempColor = &RGB{128, 128, 128}
	case 91, 101: // Bright Red
		tempColor = &RGB{255, 0, 0}
	case 92, 102: // Bright Green
		tempColor = &RGB{0, 255, 0}
	case 93, 103: // Bright Yellow
		tempColor = &RGB{255, 255, 0}
	case 94, 104: // Bright Blue
		tempColor = &RGB{0, 0, 255}
	case 95, 105: // Bright Magenta
		tempColor = &RGB{255, 0, 255}
	case 96, 106: // Bright Cyan
		tempColor = &RGB{101, 194, 205}
	case 97, 107: // Bright White
		tempColor = &RGB{255, 255, 255}
	}
	if colorInt < 40 || (colorInt >= 90 && colorInt < 100) {
		ir.foregroundColor = tempColor
		return
	}
	ir.backgroundColor = tempColor
}
