package terminal

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

// BenchmarkWriteAnchors benchmarks the writer with realistic anchor-heavy input
// containing color overrides, bold, and a hyperlink — 6+ anchor tokens total.
func BenchmarkWriteAnchors(b *testing.B) {
	Init(shell.PWSH)
	Colors = &color.Defaults{}
	colors := &color.Set{Foreground: "white", Background: "blue"}
	input := `<#CAF25B>src</> / <red,blue>main</> <b>oh-my-posh</b> <u>feat</u> <LINK>https://example.com<TEXT>link</TEXT></LINK>`
	b.ReportAllocs()
	for b.Loop() {
		CurrentColors = colors
		ParentColors = nil
		Write(colors.Background, colors.Foreground, input)
		_, _ = String()
	}
}

// BenchmarkWritePlainASCII benchmarks the writer with 80 chars of plain ASCII — no anchors.
func BenchmarkWritePlainASCII(b *testing.B) {
	Init(shell.PWSH)
	Colors = &color.Defaults{}
	colors := &color.Set{Foreground: "white", Background: "blue"}
	input := "this is a plain eighty character ascii string with no color anchors whatsoever!!"
	b.ReportAllocs()
	for b.Loop() {
		CurrentColors = colors
		ParentColors = nil
		Write(colors.Background, colors.Foreground, input)
		_, _ = String()
	}
}
