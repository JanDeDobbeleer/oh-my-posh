package main

import (
	"testing"
)

func BenchmarkEngineRenderPalette(b *testing.B) {
	var err error
	for i := 0; i < b.N; i++ {
		_, err = engineRender("jandedobbeleer-palette.omp.json")
		if err != nil {
			b.Fatal(err)
		}
	}
}
