package terminal

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

func benchmarkWrite(b *testing.B, background color.Ansi) {
	Init(shell.GENERIC)
	CurrentColors = &color.Set{Foreground: "#eff1f5", Background: background}
	ParentColors = []*color.Set{{Background: "#DD7878", Foreground: "#4c4f69"}}
	Colors = &color.Defaults{}

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		Write(background, "#eff1f5", "  oh-my-posh  src ")
		String()
	}
}

func BenchmarkWriteSolid(b *testing.B) {
	benchmarkWrite(b, "#DC8A78")
}

func BenchmarkWriteGradient(b *testing.B) {
	benchmarkWrite(b, "linear-gradient(#DC8A78, #DD7878)")
}

func BenchmarkWriteGradientChained(b *testing.B) {
	benchmarkWrite(b, "linear-gradient(parentBackground, #EA76CB)")
}
