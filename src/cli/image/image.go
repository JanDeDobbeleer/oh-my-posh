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

package image

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"io"
	"math"
	stdOS "os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	font_ "github.com/jandedobbeleer/oh-my-posh/src/cli/font"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"github.com/esimov/stackblur-go"
	"github.com/fogleman/gg"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type ConnectionError struct {
	reason string
}

func (f *ConnectionError) Error() string {
	return f.reason
}

const (
	red    = "#ED655A"
	yellow = "#E1C04C"
	green  = "#71BD47"

	// known ansi sequences

	fg                  = "FG"
	bg                  = "BG"
	bc                  = "BC" // for base 16 colors
	str                 = "STR"
	text                = "TEXT"
	invertedColor       = "inverted"
	invertedColorSingle = "invertedsingle"
	fullColor           = "full"
	foreground          = "foreground"
	background          = "background"
	reset               = "reset"
	bold                = "bold"
	boldReset           = "boldr"
	italic              = "italic"
	italicReset         = "italicr"
	underline           = "underline"
	underlineReset      = "underliner"
	overline            = "overline"
	overlineReset       = "overliner"
	strikethrough       = "strikethrough"
	strikethroughReset  = "strikethroughr"
	backgroundReset     = "backgroundr"
	color16             = "color16"
	left                = "left"
	lineChange          = "linechange"
	consoleTitle        = "title"
	link                = "link"
)

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

type Renderer struct {
	italic                 font.Face
	bold                   font.Face
	regular                font.Face
	backgroundColor        *RGB
	ansiSequenceRegexMap   map[string]string
	foregroundColor        *RGB
	defaultBackgroundColor *RGB
	defaultForegroundColor *RGB
	Settings
	Path            string
	AnsiString      string
	shadowBaseColor string
	style           string
	shadowOffsetX   float64
	margin          float64
	factor          float64
	shadowOffsetY   float64
	rows            int
	lineSpacing     float64
	columns         int
	padding         float64
	shadowRadius    uint8
}

func (ir *Renderer) Init(env runtime.Environment) error {
	ir.setOutputPath(env.Flags().ConfigPath)

	ir.cleanContent()

	if err := ir.loadFonts(); err != nil {
		return err
	}

	ir.initDefaults()

	return nil
}

func (ir *Renderer) loadFonts() error {
	if !ir.Fonts.IsValid() {
		return ir.loadDefaultFonts()
	}

	fonts, err := ir.Fonts.Load()
	if err != nil {
		return err
	}

	ir.regular = fonts[regular]
	ir.bold = fonts[bold]
	ir.italic = fonts[italic]

	return nil
}

func (ir *Renderer) initDefaults() {
	ir.defaultForegroundColor = &RGB{255, 255, 255}
	ir.defaultBackgroundColor = &RGB{21, 21, 21}

	ir.factor = 2.0
	ir.columns = 80
	ir.rows = 25

	ir.margin = ir.factor * 48
	ir.padding = ir.factor * 24

	ir.shadowBaseColor = "#10101066"
	ir.shadowRadius = uint8(math.Min(ir.factor*16, 255))
	ir.shadowOffsetX = ir.factor * 16
	ir.shadowOffsetY = ir.factor * 16

	ir.lineSpacing = 1.2

	// Set background color from settings if provided, otherwise use default
	if ir.BackgroundColor == "" {
		ir.BackgroundColor = "#151515" // Default dark background
	}

	ir.ansiSequenceRegexMap = map[string]string{
		invertedColor:       `^(?P<STR>(\x1b\[38;2;(?P<BG>(\d+;?){3});49m){1}(\x1b\[7m))`,
		invertedColorSingle: `^(?P<STR>\x1b\[(?P<BG>\d{2,3});49m\x1b\[7m)`,
		fullColor:           `^(?P<STR>(\x1b\[48;2;(?P<BG>(\d+;?){3})m)(\x1b\[38;2;(?P<FG>(\d+;?){3})m))`,
		foreground:          `^(?P<STR>(\x1b\[38;2;(?P<FG>(\d+;?){3})m))`,
		background:          `^(?P<STR>(\x1b\[48;2;(?P<BG>(\d+;?){3})m))`,
		reset:               `^(?P<STR>\x1b\[0m)`,
		bold:                `^(?P<STR>\x1b\[1m)`,
		boldReset:           `^(?P<STR>\x1b\[22m)`,
		italic:              `^(?P<STR>\x1b\[3m)`,
		italicReset:         `^(?P<STR>\x1b\[23m)`,
		underline:           `^(?P<STR>\x1b\[4m)`,
		underlineReset:      `^(?P<STR>\x1b\[24m)`,
		overline:            `^(?P<STR>\x1b\[53m)`,
		overlineReset:       `^(?P<STR>\x1b\[55m)`,
		strikethrough:       `^(?P<STR>\x1b\[9m)`,
		strikethroughReset:  `^(?P<STR>\x1b\[29m)`,
		backgroundReset:     `^(?P<STR>\x1b\[49m)`,
		color16:             `^(?P<STR>\x1b\[(?P<BC>[349][0-7]|10[0-7]|39)m)`,
		left:                `^(?P<STR>\x1b\[(\d{1,3})D)`,
		lineChange:          `^(?P<STR>\x1b\[(\d)[FB])`,
		consoleTitle:        `^(?P<STR>\x1b\]0;(.+)\007)`,
		link:                fmt.Sprintf(`^%s`, regex.LINK),
	}
}

func (ir *Renderer) setOutputPath(config string) {
	if len(ir.Path) != 0 {
		return
	}

	if config == "" {
		ir.Path = "prompt.png"
		return
	}

	config = filepath.Base(config)

	match := regex.FindNamedRegexMatch(`(\.?)(?P<STR>.*)\.(json|yaml|yml|toml|jsonc)`, config)
	path := strings.TrimRight(match[str], ".omp")

	if path == "" {
		path = "prompt"
	}

	ir.Path = fmt.Sprintf("%s.png", path)
}

func (ir *Renderer) loadDefaultFonts() error {
	var data []byte

	fontCachePath := filepath.Join(cache.Path(), "Hack.zip")
	if _, err := stdOS.Stat(fontCachePath); err == nil {
		data, _ = stdOS.ReadFile(fontCachePath)
	}

	// Download font if not cached
	if data == nil {
		url := "https://github.com/ryanoasis/nerd-fonts/releases/download/v3.2.1/Hack.zip"
		var err error

		data, err = font_.Download(url)
		if err != nil {
			return &ConnectionError{reason: err.Error()}
		}

		err = stdOS.WriteFile(fontCachePath, data, 0644)
		if err != nil {
			return err
		}
	}

	bytesReader := bytes.NewReader(data)
	zipReader, err := zip.NewReader(bytesReader, int64(bytesReader.Len()))
	if err != nil {
		return err
	}

	fontFaceOptions := &opentype.FaceOptions{Size: 2.0 * 12, DPI: 144}

	parseFont := func(file *zip.File) (font.Face, error) {
		rc, err := file.Open()
		if err != nil {
			return nil, err
		}

		defer rc.Close()

		data, err := io.ReadAll(rc)
		if err != nil {
			return nil, err
		}

		font, err := opentype.Parse(data)
		if err != nil {
			return nil, err
		}

		fontFace, err := opentype.NewFace(font, fontFaceOptions)
		if err != nil {
			return nil, err
		}
		return fontFace, nil
	}

	for _, file := range zipReader.File {
		switch file.Name {
		case "HackNerdFont-Regular.ttf":
			if regular, err := parseFont(file); err == nil {
				ir.regular = regular
			}
		case "HackNerdFont-Bold.ttf":
			if bold, err := parseFont(file); err == nil {
				ir.bold = bold
			}
		case "HackNerdFont-Italic.ttf":
			if italic, err := parseFont(file); err == nil {
				ir.italic = italic
			}
		}
	}

	return nil
}

func (ir *Renderer) fontHeight() float64 {
	return float64(ir.regular.Metrics().Height >> 6)
}

type RuneRange struct {
	Start rune
	End   rune
}

// If we're a Nerd Font code point, treat as double width
var doubleWidthRunes = []RuneRange{
	// Seti-UI + Custom range
	{Start: '\ue5fa', End: '\ue6b1'},
	// Devicons
	{Start: '\ue700', End: '\ue7c5'},
	// Font Awesome
	{Start: '\uf000', End: '\uf2e0'},
	// Font Awesome Extension
	{Start: '\ue200', End: '\ue2a9'},
	// Material Design Icons
	{Start: '\U000f0001', End: '\U000f1af0'},
	// Weather
	{Start: '\ue300', End: '\ue3e3'},
	// Octicons
	{Start: '\uf400', End: '\uf532'},
	{Start: '\u2665', End: '\u2665'},
	{Start: '\u26A1', End: '\u26A1'},
	// Powerline Extra Symbols (intentionally excluding single width bubbles (e0b4-e0b7) and pixelated (e0c4-e0c7))
	{Start: '\ue0a3', End: '\ue0a3'},
	{Start: '\ue0b4', End: '\ue0c8'},
	{Start: '\ue0ca', End: '\ue0ca'},
	{Start: '\ue0cc', End: '\ue0d4'},
	// IEC Power Symbols
	{Start: '\u23fb', End: '\u23fe'},
	{Start: '\u2b58', End: '\u2b58'},
	// Font Logos
	{Start: '\uf300', End: '\uf372'},
	// Pomicons
	{Start: '\ue000', End: '\ue00a'},
	// Codicons
	{Start: '\uea60', End: '\uebeb'},
}

// This is getting how many additional characters of width to allocate when drawing
// e.g. for characters that are 2 or more wide. A standard character will return 0
// Nerd Font glyphs will return 1, since most are double width
func (ir *Renderer) runeAdditionalWidth(r rune) int {
	// exclude the round leading diamond
	singles := []rune{'\ue0b6', '\ue0ba', '\ue0bc'}
	if slices.Contains(singles, r) {
		return 0
	}

	for _, runeRange := range doubleWidthRunes {
		if runeRange.Start <= r && r <= runeRange.End {
			return 1
		}
	}
	return 0
}

func (ir *Renderer) cleanContent() {
	// clean abundance of empty lines
	ir.AnsiString = strings.Trim(ir.AnsiString, "\n")
	ir.AnsiString = "\n" + ir.AnsiString

	// clean string before render
	ir.AnsiString = strings.ReplaceAll(ir.AnsiString, "\x1b[m", "\x1b[0m")
	ir.AnsiString = strings.ReplaceAll(ir.AnsiString, "\x1b[K", "")
	ir.AnsiString = strings.ReplaceAll(ir.AnsiString, "\x1b[0J", "")
	ir.AnsiString = strings.ReplaceAll(ir.AnsiString, "\x1b[27m", "")
	ir.AnsiString = strings.ReplaceAll(ir.AnsiString, "\x1b8", "")
	ir.AnsiString = strings.ReplaceAll(ir.AnsiString, "\u2800", " ")

	// cursor indication
	saveCursorAnsi := "\x1b7"
	if !strings.Contains(ir.AnsiString, saveCursorAnsi) {
		ir.AnsiString += ir.Cursor
	}

	ir.AnsiString = strings.ReplaceAll(ir.AnsiString, saveCursorAnsi, ir.Cursor)

	// add watermarks
	ir.AnsiString += "\n\n\x1b[1mohmyposh.dev\x1b[22m"
	if len(ir.Author) > 0 {
		createdBy := fmt.Sprintf(" by \x1b[1m%s\x1b[22m", ir.Author)
		ir.AnsiString += createdBy
	}
}

func (ir *Renderer) measureContent() (width, height float64) {
	// Use actual rendering logic for accurate width measurement
	// This simulates the exact same process as the actual drawing to ensure
	// the canvas width perfectly matches the rendered content width
	var maxX float64
	var x float64

	// Save original ansi string and style state
	originalAnsi := ir.AnsiString
	originalStyle := ir.style
	ir.style = ""

	tmpDrawer := &font.Drawer{Face: ir.regular}

	for ir.AnsiString != "" {
		if !ir.processAnsiSequence() {
			continue
		}

		runes := []rune(ir.AnsiString)
		if len(runes) == 0 {
			continue
		}

		str := string(runes[0:1])
		ir.AnsiString = string(runes[1:])

		// Use appropriate font face for measurement
		var face font.Face
		switch ir.style {
		case bold:
			face = ir.bold
		case italic:
			face = ir.italic
		default:
			face = ir.regular
		}

		tmpDrawer.Face = face
		advance := tmpDrawer.MeasureString(str)
		w := float64(advance >> 6)

		// Add additional width for Nerd Font glyphs
		w += (w * float64(ir.runeAdditionalWidth(runes[0])))

		if str == "\n" {
			x = 0
			continue
		}

		x += w
		if x > maxX {
			maxX = x
		}
	}

	// Restore original state
	ir.AnsiString = originalAnsi
	ir.style = originalStyle

	// Ensure we have a minimum width for very short content
	minWidth := tmpDrawer.MeasureString(strings.Repeat(" ", 80))
	width = math.Max(maxX, float64(minWidth>>6))

	// height, lines times font height and line spacing
	lines := strings.Split(originalAnsi, "\n")
	height = float64(len(lines)) * ir.fontHeight() * ir.lineSpacing
	return width, height
}

func (ir *Renderer) SavePNG() error {
	var scale = func(value float64) float64 { return ir.factor * value }

	var (
		corner   = scale(6)
		radius   = scale(9)
		distance = scale(25)
	)

	contentWidth, contentHeight := ir.measureContent()

	// Make sure the output window is big enough in case no content or very few
	// content will be rendered. Also account for potential font variations.
	minRequiredWidth := 3*distance + 3*radius
	// Add extra buffer for wider fonts (20% more than minimum)
	minRequiredWidth *= 1.2
	contentWidth = math.Max(contentWidth, minRequiredWidth)

	marginX, marginY := ir.margin, ir.margin
	paddingX, paddingY := ir.padding, ir.padding

	xOffset := marginX
	yOffset := marginY
	titleOffset := scale(40)

	width := contentWidth + 2*marginX + 2*paddingX
	height := contentHeight + 2*marginY + 2*paddingY + titleOffset

	dc := gg.NewContext(int(width), int(height))

	xOffset -= ir.shadowOffsetX / 2
	yOffset -= ir.shadowOffsetY / 2

	bc := gg.NewContext(int(width), int(height))
	bc.DrawRoundedRectangle(xOffset+ir.shadowOffsetX, yOffset+ir.shadowOffsetY, width-2*marginX, height-2*marginY, corner)
	bc.SetHexColor(ir.shadowBaseColor)
	bc.Fill()

	dst := image.NewNRGBA(bc.Image().Bounds())

	// var done = make(chan struct{}, ir.shadowRadius)
	err := stackblur.Process(
		dst,
		bc.Image(),
		uint32(ir.shadowRadius),
	)

	if err != nil {
		return err
	}

	// <-done
	dc.DrawImage(dst, 0, 0)

	// Draw rounded rectangle with outline and three button to produce the
	// impression of a window with controls and a content area
	dc.DrawRoundedRectangle(xOffset, yOffset, width-2*marginX, height-2*marginY, corner)
	dc.SetHexColor(ir.BackgroundColor)
	dc.Fill()

	dc.DrawRoundedRectangle(xOffset, yOffset, width-2*marginX, height-2*marginY, corner)
	dc.SetHexColor("#404040")
	dc.SetLineWidth(scale(1))
	dc.Stroke()

	for i, color := range []string{red, yellow, green} {
		dc.DrawCircle(xOffset+paddingX+float64(i)*distance+scale(4), yOffset+paddingY+scale(4), radius)
		dc.SetHexColor(color)
		dc.Fill()
	}

	// Apply the actual text into the prepared content area of the window
	var x, y = xOffset + paddingX, yOffset + paddingY + titleOffset + ir.fontHeight()

	for ir.AnsiString != "" {
		if !ir.processAnsiSequence() {
			continue
		}

		runes := []rune(ir.AnsiString)
		if len(runes) == 0 {
			continue
		}

		str := string(runes[0:1])
		ir.AnsiString = string(runes[1:])
		switch ir.style {
		case bold:
			dc.SetFontFace(ir.bold)
		case italic:
			dc.SetFontFace(ir.italic)
		default:
			dc.SetFontFace(ir.regular)
		}

		w, _ := dc.MeasureString(str)
		// The gg library unfortunately returns a single character width for *all* glyphs in a font.
		// So if we know the glyph to occupy n additional characters in width, allocate that area
		// e.g. this will double the space for Nerd Fonts, but some could even be 3 or 4 wide
		// If there's 0 additional characters of width (the common case), this won't add anything
		w += (w * float64(ir.runeAdditionalWidth(runes[0])))

		if ir.backgroundColor != nil {
			dc.SetRGB255(ir.backgroundColor.r, ir.backgroundColor.g, ir.backgroundColor.b)
			// Use consistent line height for all background rectangles
			fontLineHeight := ir.fontHeight() * ir.lineSpacing

			// Center all characters (including powerline glyphs) within the line height
			// Position background to align properly with text baseline and ensure consistent height
			bgY := y - fontLineHeight*0.75 // Adjusted for better centering with text
			bgHeight := fontLineHeight

			dc.DrawRectangle(x, bgY, w, bgHeight)
			dc.Fill()
		}

		if ir.foregroundColor != nil {
			dc.SetRGB255(ir.foregroundColor.r, ir.foregroundColor.g, ir.foregroundColor.b)
		} else {
			dc.SetRGB255(ir.defaultForegroundColor.r, ir.defaultForegroundColor.g, ir.defaultForegroundColor.b)
		}

		if str == "\n" {
			x = xOffset + paddingX
			y += ir.fontHeight() * ir.lineSpacing // Use consistent line height instead of character height
			continue
		}

		dc.DrawString(str, x, y)

		if ir.style == underline {
			dc.DrawLine(x, y+scale(4), x+w, y+scale(4))
			dc.SetLineWidth(scale(1))
			dc.Stroke()
		}

		if ir.style == overline {
			dc.DrawLine(x, y-scale(22), x+w, y-scale(22))
			dc.SetLineWidth(scale(1))
			dc.Stroke()
		}

		x += w
	}

	return dc.SavePNG(ir.Path)
}

func (ir *Renderer) processAnsiSequence() bool {
	for sequence, re := range ir.ansiSequenceRegexMap {
		match := regex.FindNamedRegexMatch(re, ir.AnsiString)
		if len(match) == 0 {
			continue
		}

		ir.AnsiString = strings.TrimPrefix(ir.AnsiString, match[str])
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
		case background:
			ir.backgroundColor = NewRGBColor(match[bg])
			return false
		case reset:
			ir.foregroundColor = ir.defaultForegroundColor
			ir.backgroundColor = nil
			return false
		case backgroundReset:
			ir.backgroundColor = nil
			return false
		case bold, italic, underline, overline:
			ir.style = sequence
			return false
		case boldReset, italicReset, underlineReset, overlineReset:
			ir.style = ""
			return false
		case strikethrough, strikethroughReset, left, lineChange, consoleTitle:
			return false
		case color16:
			ir.setBase16Color(match[bc])
			return false
		case link:
			ir.AnsiString = match[text] + ir.AnsiString
		}
	}

	return true
}

func (ir *Renderer) setBase16Color(colorStr string) {
	tempColor := ir.defaultForegroundColor

	colorInt, err := strconv.Atoi(colorStr)
	if err != nil {
		ir.foregroundColor = tempColor
		return
	}

	// Check for color override first
	colorName := colorNameFromCode(colorInt)
	if rgb, err := ir.Colors.RGBFromColorName(colorName); err == nil {
		tempColor = rgb
	}

	// If no override found, use default colors
	if tempColor == ir.defaultForegroundColor {
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
	}

	if colorInt < 40 || (colorInt >= 90 && colorInt < 100) {
		ir.foregroundColor = tempColor
		return
	}

	ir.backgroundColor = tempColor
}

// colorNameFromCode maps ANSI color codes to color names
func colorNameFromCode(colorInt int) string {
	switch colorInt {
	case 30, 40:
		return "black"
	case 31, 41:
		return "red"
	case 32, 42:
		return "green"
	case 33, 43:
		return "yellow"
	case 34, 44:
		return "blue"
	case 35, 45:
		return "magenta"
	case 36, 46:
		return "cyan"
	case 37, 47:
		return "white"
	case 90, 100:
		return "darkGray"
	case 91, 101:
		return "lightRed"
	case 92, 102:
		return "lightGreen"
	case 93, 103:
		return "lightYellow"
	case 94, 104:
		return "lightBlue"
	case 95, 105:
		return "lightMagenta"
	case 96, 106:
		return "lightCyan"
	case 97, 107:
		return "lightWhite"
	default:
		return ""
	}
}
