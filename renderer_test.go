package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/gookit/color"
)

func TestWriteAndRemoveText(t *testing.T) {
	renderer := &Renderer{
		Buffer: new(bytes.Buffer),
	}
	renderer.init("pwsh")
	inputText := "This is white, <#ff5733>this is orange</>, white again"
	text := renderer.writeAndRemoveText("#193549", "#fff", "This is white, ", "This is white, ", inputText)
	assert.Equal(t, "<#ff5733>this is orange</>, white again", text)
	assert.NotContains(t, renderer.string(), "<#ff5733>")
}

func TestWriteAndRemoveTextColored(t *testing.T) {
	renderer := &Renderer{
		Buffer: new(bytes.Buffer),
	}
	renderer.init("pwsh")
	inputText := "This is white, <#ff5733>this is orange</>, white again"
	text := renderer.writeAndRemoveText("#193549", "#ff5733", "this is orange", "<#ff5733>this is orange</>", inputText)
	assert.Equal(t, "This is white, , white again", text)
	assert.NotContains(t, renderer.string(), "<#ff5733>")
}

func TestWriteColorOverride(t *testing.T) {
	renderer := &Renderer{
		Buffer: new(bytes.Buffer),
	}
	renderer.init("pwsh")
	text := "This is white, <#ff5733>this is orange</>, white again"
	renderer.write("#193549", "#ff5733", text)
	assert.NotContains(t, renderer.string(), "<#ff5733>")
}

func TestWriteColorTransparent(t *testing.T) {
	renderer := &Renderer{
		Buffer: new(bytes.Buffer),
	}
	renderer.init("pwsh")
	text := "This is white"
	renderer.writeColoredText("#193549", Transparent, text)
	t.Log(renderer.string())
}

func TestWriteColorName(t *testing.T) {
	// given
	renderer := &Renderer{
		Buffer: new(bytes.Buffer),
	}
	renderer.init("pwsh")
	text := "This is white, <red>this is red</>, white again"

	// when
	renderer.write("#193549", "red", text)

	// then
	assert.NotContains(t, renderer.string(), "<red>")
}

func TestWriteColorInvalid(t *testing.T) {
	// given
	renderer := &Renderer{
		Buffer: new(bytes.Buffer),
	}
	renderer.init("pwsh")
	text := "This is white, <invalid>this is orange</>, white again"

	// when
	renderer.write("#193549", "invalid", text)

	// then
	assert.Contains(t, renderer.string(), "<invalid>")
}

func TestLenWithoutANSI(t *testing.T) {
	text := "\x1b[44mhello\x1b[0m"
	renderer := &Renderer{
		Buffer: new(bytes.Buffer),
	}
	renderer.init("pwsh")
	strippedLength := renderer.lenWithoutANSI(text)
	assert.Equal(t, 5, strippedLength)
}

func TestLenWithoutANSIZsh(t *testing.T) {
	text := "%{\x1b[44m%}hello%{\x1b[0m%}"
	renderer := &Renderer{
		Buffer: new(bytes.Buffer),
	}
	renderer.init("zsh")
	strippedLength := renderer.lenWithoutANSI(text)
	assert.Equal(t, 5, strippedLength)
}

func TestGetAnsiFromColorStringBg(t *testing.T) {
	// given
	renderer := &Renderer{
		Buffer: new(bytes.Buffer),
	}

	// when
	colorCode := renderer.getAnsiFromColorString("blue", true)

	// then
	assert.Equal(t, color.BgBlue.Code(), colorCode)
}

func TestGetAnsiFromColorStringFg(t *testing.T) {
	// given
	renderer := &Renderer{
		Buffer: new(bytes.Buffer),
	}

	// when
	colorCode := renderer.getAnsiFromColorString("red", false)

	// then
	assert.Equal(t, color.FgRed.Code(), colorCode)
}

func TestGetAnsiFromColorStringHex(t *testing.T) {
	// given
	renderer := &Renderer{
		Buffer: new(bytes.Buffer),
	}

	// when
	colorCode := renderer.getAnsiFromColorString("#AABBCC", false)

	// then
	assert.Equal(t, color.HEX("#AABBCC").Code(), colorCode)
}

func TestGetAnsiFromColorStringInvalidFg(t *testing.T) {
	// given
	renderer := &Renderer{
		Buffer: new(bytes.Buffer),
	}

	// when
	colorCode := renderer.getAnsiFromColorString("invalid", false)

	// then
	assert.Equal(t, "", colorCode)
}

func TestGetAnsiFromColorStringInvalidBg(t *testing.T) {
	// given
	renderer := &Renderer{
		Buffer: new(bytes.Buffer),
	}

	// when
	colorCode := renderer.getAnsiFromColorString("invalid", true)

	// then
	assert.Equal(t, "", colorCode)
}
