package main

import (
	"testing"

	"github.com/gookit/color"
	"github.com/stretchr/testify/assert"
)

const (
	inputText = "This is white, <#ff5733>this is orange</>, white again"
)

func TestWriteAndRemoveText(t *testing.T) {
	formats := &ansiFormats{}
	formats.init("pwsh")
	renderer := &ANSIWriter{
		formats: formats,
	}
	text := renderer.writeAndRemoveText("#193549", "#fff", "This is white, ", "This is white, ", inputText)
	assert.Equal(t, "<#ff5733>this is orange</>, white again", text)
	assert.NotContains(t, renderer.render(), "<#ff5733>")
}

func TestWriteAndRemoveTextColored(t *testing.T) {
	formats := &ansiFormats{}
	formats.init("pwsh")
	renderer := &ANSIWriter{
		formats: formats,
	}
	text := renderer.writeAndRemoveText("#193549", "#ff5733", "this is orange", "<#ff5733>this is orange</>", inputText)
	assert.Equal(t, "This is white, , white again", text)
	assert.NotContains(t, renderer.render(), "<#ff5733>")
}

func TestWriteColorOverride(t *testing.T) {
	formats := &ansiFormats{}
	formats.init("pwsh")
	renderer := &ANSIWriter{
		formats: formats,
	}
	renderer.write("#193549", "#ff5733", inputText)
	assert.NotContains(t, renderer.render(), "<#ff5733>")
}

func TestWriteColorOverrideBackground(t *testing.T) {
	formats := &ansiFormats{}
	formats.init("pwsh")
	renderer := &ANSIWriter{
		formats: formats,
	}
	text := "This is white, <,#000000>this is black</>, white again"
	renderer.write("#193549", "#ff5733", text)
	assert.NotContains(t, renderer.render(), "000000")
}

func TestWriteColorOverrideBackground16(t *testing.T) {
	formats := &ansiFormats{}
	formats.init("pwsh")
	renderer := &ANSIWriter{
		formats: formats,
	}
	text := "This is default <,white> this background is changed</> default again"
	renderer.write("#193549", "#ff5733", text)
	assert.NotContains(t, renderer.render(), "white")
	assert.NotContains(t, renderer.render(), "</>")
	assert.NotContains(t, renderer.render(), "<,")
}

func TestWriteColorOverrideBoth(t *testing.T) {
	formats := &ansiFormats{}
	formats.init("pwsh")
	renderer := &ANSIWriter{
		formats: formats,
	}
	text := "This is white, <#000000,#ffffff>this is black</>, white again"
	renderer.write("#193549", "#ff5733", text)
	assert.NotContains(t, renderer.render(), "ffffff")
	assert.NotContains(t, renderer.render(), "000000")
}

func TestWriteColorOverrideBoth16(t *testing.T) {
	formats := &ansiFormats{}
	formats.init("pwsh")
	renderer := &ANSIWriter{
		formats: formats,
	}
	text := "This is white, <black,white>this is black</>, white again"
	renderer.write("#193549", "#ff5733", text)
	assert.NotContains(t, renderer.render(), "<black,white>")
	assert.NotContains(t, renderer.render(), "</>")
}

func TestWriteColorOverrideDouble(t *testing.T) {
	formats := &ansiFormats{}
	formats.init("pwsh")
	renderer := &ANSIWriter{
		formats: formats,
	}
	text := "<#ffffff>jan</>@<#ffffff>Jans-MBP</>"
	renderer.write("#193549", "#ff5733", text)
	assert.NotContains(t, renderer.render(), "<#ffffff>")
	assert.NotContains(t, renderer.render(), "</>")
}

func TestWriteColorTransparent(t *testing.T) {
	formats := &ansiFormats{}
	formats.init("pwsh")
	renderer := &ANSIWriter{
		formats: formats,
	}
	text := "This is white"
	renderer.writeColoredText("#193549", Transparent, text)
	t.Log(renderer.render())
}

func TestWriteColorName(t *testing.T) {
	formats := &ansiFormats{}
	formats.init("pwsh")
	renderer := &ANSIWriter{
		formats: formats,
	}
	text := "This is white, <red>this is red</>, white again"
	renderer.write("#193549", "red", text)
	assert.NotContains(t, renderer.render(), "<red>")
}

func TestWriteColorInvalid(t *testing.T) {
	formats := &ansiFormats{}
	formats.init("pwsh")
	renderer := &ANSIWriter{
		formats: formats,
	}
	text := "This is white, <invalid>this is orange</>, white again"
	renderer.write("#193549", "invalid", text)
	assert.Contains(t, renderer.render(), "<invalid>")
}

func TestGetAnsiFromColorStringBg(t *testing.T) {
	renderer := &ANSIWriter{}
	colorCode := renderer.getAnsiFromColorString("blue", true)
	assert.Equal(t, color.BgBlue.Code(), colorCode)
}

func TestGetAnsiFromColorStringFg(t *testing.T) {
	renderer := &ANSIWriter{}
	colorCode := renderer.getAnsiFromColorString("red", false)
	assert.Equal(t, color.FgRed.Code(), colorCode)
}

func TestGetAnsiFromColorStringHex(t *testing.T) {
	renderer := &ANSIWriter{}
	colorCode := renderer.getAnsiFromColorString("#AABBCC", false)
	assert.Equal(t, color.HEX("#AABBCC").Code(), colorCode)
}

func TestGetAnsiFromColorStringInvalidFg(t *testing.T) {
	renderer := &ANSIWriter{}
	colorCode := renderer.getAnsiFromColorString("invalid", false)
	assert.Equal(t, "", colorCode)
}

func TestGetAnsiFromColorStringInvalidBg(t *testing.T) {
	renderer := &ANSIWriter{}
	colorCode := renderer.getAnsiFromColorString("invalid", true)
	assert.Equal(t, "", colorCode)
}
