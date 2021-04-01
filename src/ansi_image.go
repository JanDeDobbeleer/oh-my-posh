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

package main

import (
	_ "embed"
	colour "image/color"
	"math"
	"strings"

	"github.com/esimov/stackblur-go"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/gonvenience/term"
	"github.com/gookit/color"
	"golang.org/x/image/font"
)

const (
	red    = "#ED655A"
	yellow = "#E1C04C"
	green  = "#71BD47"
)

type ImageWriter struct {
	engine     *engine
	ansiWriter *ANSIWriter
	canvas     *gg.Context

	factor float64

	columns int
	rows    int

	defaultForegroundColor colour.Color

	drawShadow      bool
	shadowBaseColor string
	shadowRadius    uint8
	shadowOffsetX   float64
	shadowOffsetY   float64

	padding float64
	margin  float64

	regular     font.Face
	bold        font.Face
	italic      font.Face
	boldItalic  font.Face
	lineSpacing float64
	tabSpaces   int

	textPositionX float64
	textPositionY float64
	xOffset       float64
	yOffset       float64
	paddingX      float64
}

//go:embed font/VictorMono-Bold.ttf
var victorMonoBold []byte

//go:embed font/VictorMono-Regular.ttf
var victorMonoRegular []byte

//go:embed font/VictorMono-BoldItalic.ttf
var victorMonoBoldItalic []byte

//go:embed font/VictorMono-Italic.ttf
var victorMonoItalic []byte

func NewImageCreator(ansiWriter *ANSIWriter, engine *engine) *ImageWriter {
	f := 2.0

	fontRegular, _ := truetype.Parse(victorMonoRegular)
	fontBold, _ := truetype.Parse(victorMonoBold)
	fontItalic, _ := truetype.Parse(victorMonoItalic)
	fontBoldItalic, _ := truetype.Parse(victorMonoBoldItalic)
	fontFaceOptions := &truetype.Options{Size: f * 12, DPI: 144}

	cols, rows := term.GetTerminalSize()
	return &ImageWriter{
		ansiWriter:             ansiWriter,
		engine:                 engine,
		defaultForegroundColor: colour.Black,

		factor: f,

		columns: cols,
		rows:    rows,

		margin:  f * 48,
		padding: f * 24,

		drawShadow:      true,
		shadowBaseColor: "#10101066",
		shadowRadius:    uint8(math.Min(f*16, 255)),
		shadowOffsetX:   f * 16,
		shadowOffsetY:   f * 16,

		regular:     truetype.NewFace(fontRegular, fontFaceOptions),
		bold:        truetype.NewFace(fontBold, fontFaceOptions),
		italic:      truetype.NewFace(fontItalic, fontFaceOptions),
		boldItalic:  truetype.NewFace(fontBoldItalic, fontFaceOptions),
		lineSpacing: 1.2,
		tabSpaces:   2,
	}
}

func (iw *ImageWriter) renderImage() {
	//todo: fetch the string
	// iw.engine.print()
	iw.SetupPNG("[38;2;195;134;241m[0m[48;2;195;134;241m[38;2;195;134;241m█[0m[48;2;195;134;241m[38;2;255;255;255mjan[0m[48;2;195;134;241m[38;2;255;255;255m@[0m[48;2;195;134;241m[38;2;255;255;255mJans-MBP[0m[48;2;195;134;241m[38;2;195;134;241m█[0m[38;2;195;134;241m[0m[38;2;255;71;156;49m[7m[m[0m[48;2;255;71;156m[38;2;255;255;255m  src[0m[48;2;255;71;156m[38;2;255;71;156m█[0m[48;2;255;251;56m[38;2;255;71;156m[0m[48;2;255;251;56m[38;2;255;251;56m█[0m[48;2;255;251;56m[38;2;25;53;73m screenshot-themes ≡  ~10 -2 ?4  8[0m[48;2;255;251;56m[38;2;255;251;56m█[0m[48;2;255;87;34m[38;2;255;251;56m[0m[48;2;255;87;34m[38;2;255;87;34m█[0m[48;2;255;87;34m[38;2;25;53;73m72 [0m[48;2;0;119;194m[38;2;255;87;34m[0m[48;2;0;119;194m[38;2;255;255;255m ﲵ uni[0m[48;2;0;119;194m[38;2;0;119;194m█[0m[48;2;255;255;255m[38;2;0;119;194m[0m[48;2;255;255;255m[38;2;255;255;255m█[0m[48;2;255;255;255m[38;2;17;17;17mINVALID CONFIG PATH[0m[48;2;255;255;255m[38;2;255;255;255m█[0m[38;2;255;255;255m[0m[38;2;46;149;153;49m[7m[m[0m[48;2;46;149;153m[38;2;255;255;255m [0m[48;2;46;149;153m[38;2;46;149;153m█[0m[38;2;46;149;153m[0m]0;src[0m ")
	iw.engine.writer = iw
	iw.engine.render()
}

func (iw *ImageWriter) fontHeight() float64 {
	return float64(iw.regular.Metrics().Height >> 6)
}

func (iw *ImageWriter) measureContent(content string) (width, height float64) {
	lines := strings.Split(
		strings.TrimSuffix(
			content,
			"\n",
		),
		"\n",
	)

	// width, max width of all lines
	for _, line := range lines {
		advance := iw.ansiWriter.formats.lenWithoutANSI(line)
		if lineWidth := float64(advance >> 6); lineWidth > width {
			width = lineWidth
		}
	}

	// height, lines times font height and line spacing
	height = float64(len(lines)) * iw.fontHeight() * iw.lineSpacing

	return width, height
}

func (iw *ImageWriter) SetupPNG(prompt string) {
	var f = func(value float64) float64 { return iw.factor * value }

	var (
		corner   = f(6)
		radius   = f(9)
		distance = f(25)
	)

	contentWidth, contentHeight := iw.measureContent(prompt)

	// Make sure the output window is big enough in case no content or very few
	// content will be rendered
	contentWidth = math.Max(contentWidth, 3*distance+3*radius)

	marginX, marginY := iw.margin, iw.margin
	iw.paddingX = iw.padding
	paddingY := iw.padding

	iw.xOffset = marginX
	iw.yOffset = marginY
	titleOffset := f(40)

	width := contentWidth + 2*marginX + 2*iw.paddingX
	height := contentHeight + 2*marginY + 2*paddingY + titleOffset

	iw.canvas = gg.NewContext(int(width), int(height))

	// Optional: Apply blurred rounded rectangle to mimic the window shadow
	//
	if iw.drawShadow {
		iw.xOffset -= iw.shadowOffsetX / 2
		iw.yOffset -= iw.shadowOffsetY / 2

		bc := gg.NewContext(int(width), int(height))
		bc.DrawRoundedRectangle(iw.xOffset+iw.shadowOffsetX, iw.yOffset+iw.shadowOffsetY, width-2*marginX, height-2*marginY, corner)
		bc.SetHexColor(iw.shadowBaseColor)
		bc.Fill()

		var done = make(chan struct{}, iw.shadowRadius)
		shadow := stackblur.Process(
			bc.Image(),
			uint32(width),
			uint32(height),
			uint32(iw.shadowRadius),
			done,
		)

		<-done
		iw.canvas.DrawImage(shadow, 0, 0)
	}

	// Draw rounded rectangle with outline and three button to produce the
	// impression of a window with controls and a content area
	iw.canvas.DrawRoundedRectangle(iw.xOffset, iw.yOffset, width-2*marginX, height-2*marginY, corner)
	iw.canvas.SetHexColor("#151515")
	iw.canvas.Fill()

	iw.canvas.DrawRoundedRectangle(iw.xOffset, iw.yOffset, width-2*marginX, height-2*marginY, corner)
	iw.canvas.SetHexColor("#404040")
	iw.canvas.SetLineWidth(f(1))
	iw.canvas.Stroke()

	for i, color := range []string{red, yellow, green} {
		iw.canvas.DrawCircle(iw.xOffset+iw.paddingX+float64(i)*distance+f(4), iw.yOffset+paddingY+f(4), radius)
		iw.canvas.SetHexColor(color)
		iw.canvas.Fill()
	}

	iw.textPositionX, iw.textPositionY = iw.xOffset+iw.paddingX, iw.yOffset+paddingY+titleOffset+iw.fontHeight()
}

func (iw *ImageWriter) setCanvasRGBFromColorString(text, colorString string, isBackground bool) {
	getRGBFromColorName := func(colorName string, isisBackground bool) color.RGBColor {
		var colors map[string]color.Color
		var exColors map[string]color.Color
		if isisBackground {
			colors = color.BgColors
			exColors = color.ExBgColors
		} else {
			colors = color.FgColors
			exColors = color.ExFgColors
		}
		var c color.Color
		if strings.HasPrefix(colorName, "light") {
			c = exColors[colorName]
		} else {
			c = colors[colorName]
		}
		rgb := color.C256ToRgb(uint8(c))
		return color.RGBFromSlice(rgb, isBackground)
	}
	colorFromName, err := getColorFromName(colorString, isBackground)
	var rgbColor color.RGBColor
	if err == nil {
		rgbColor = getRGBFromColorName(colorFromName, isBackground)
	} else {
		rgbColor = color.HEX(colorString, isBackground)
	}
	iw.canvas.SetRGB255(rgbColor.Values()[0], rgbColor.Values()[1], rgbColor.Values()[2])
	if isBackground {
		w, h := iw.canvas.MeasureString(text)
		iw.canvas.DrawRectangle(iw.textPositionX, iw.textPositionY-h+12, w, h)
		iw.canvas.Fill()
		return
	}
}

func (iw *ImageWriter) writeColoredText(background, foreground, text string) {
	// Avoid emitting empty strings with color codes
	if text == "" {
		return
	}
	if foreground == Transparent && background != "" && iw.ansiWriter.terminalBackground != "" {
		iw.setCanvasRGBFromColorString(text, background, true)
		iw.setCanvasRGBFromColorString(text, iw.ansiWriter.terminalBackground, false)
	}
	if foreground == Transparent && background != "" {
		iw.setCanvasRGBFromColorString(text, background, false)
		iw.canvas.SetColor(iw.defaultForegroundColor)
	} else if background == "" || background == Transparent {
		iw.setCanvasRGBFromColorString(text, foreground, false)
	} else {
		iw.setCanvasRGBFromColorString(text, background, true)
		iw.setCanvasRGBFromColorString(text, foreground, false)
	}
	w, h := iw.canvas.MeasureString(text)
	switch text {
	case "\n":
		iw.textPositionX = iw.xOffset + iw.paddingX
		iw.textPositionY += h * iw.lineSpacing
		return
	case "\t":
		iw.textPositionX += w * float64(iw.tabSpaces)
		return
	case "✗": // mitigate issue #1 by replacing it with a similar character
		text = "×"
	}

	iw.canvas.DrawString(text, iw.textPositionX, iw.textPositionY)
	iw.textPositionX += w
}

func (iw *ImageWriter) writeAndRemoveText(background, foreground, text, textToRemove, parentText string) string {
	iw.writeColoredText(background, foreground, text)
	return strings.Replace(parentText, textToRemove, "", 1)
}

func (iw *ImageWriter) write(background, foreground, text string) {
	text = iw.ansiWriter.formats.formatText(text)
	// first we match for any potentially valid colors enclosed in <>
	match := findAllNamedRegexMatch(`<(?P<foreground>[^,>]+)?,?(?P<background>[^>]+)?>(?P<content>[^<]*)<\/>`, text)
	for i := range match {
		extractedForegroundColor := match[i]["foreground"]
		extractedBackgroundColor := match[i]["background"]
		if col := iw.ansiWriter.getAnsiFromColorString(extractedForegroundColor, false); col == "" && extractedForegroundColor != Transparent && len(extractedBackgroundColor) == 0 {
			continue // we skip invalid colors
		}
		if col := iw.ansiWriter.getAnsiFromColorString(extractedBackgroundColor, true); col == "" && extractedBackgroundColor != Transparent && len(extractedForegroundColor) == 0 {
			continue // we skip invalid colors
		}
		// reuse function colors if only one was specified
		if len(extractedBackgroundColor) == 0 {
			extractedBackgroundColor = background
		}
		if len(extractedForegroundColor) == 0 {
			extractedForegroundColor = foreground
		}
		escapedTextSegment := match[i]["text"]
		innerText := match[i]["content"]
		textBeforeColorOverride := strings.Split(text, escapedTextSegment)[0]
		text = iw.writeAndRemoveText(background, foreground, textBeforeColorOverride, textBeforeColorOverride, text)
		text = iw.writeAndRemoveText(extractedBackgroundColor, extractedForegroundColor, innerText, escapedTextSegment, text)
	}
	// color the remaining part of text with background and foreground
	iw.writeColoredText(background, foreground, text)
}

func (iw *ImageWriter) render() string {
	err := iw.canvas.SavePNG("prompt.png")
	if err != nil {
		return err.Error()
	}
	return "prompt.png"
}

func (iw *ImageWriter) reset() {}
