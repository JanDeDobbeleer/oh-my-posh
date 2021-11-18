package main

import (
	"testing"

	"github.com/alecthomas/assert"
)

// FIXME: test is not completely deterministic, engineReder calls can have different
// outputs due to time - there can be 1 second difference on the clocks during
// rendering. The test is useful, and removing it completely will be incorrect - having
// a close-to-real test of no palette vs palette rendering parity is better than not
// having one. Not sure what to do with this one.
func TestEngineRendersSamePrompt(t *testing.T) {
	var (
		err                    error
		noPalette, withPalette string
	)

	noPalette, err = engineRender("jandedobbeleer.omp.json")
	if err != nil {
		t.Fatal(err)
	}

	withPalette, err = engineRender("jandedobbeleer-palette.omp.json")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, noPalette, withPalette)
}

func BenchmarkEngineRenderPalette(b *testing.B) {
	var err error
	for i := 0; i < b.N; i++ {
		_, err = engineRender("jandedobbeleer-palette.omp.json")
		if err != nil {
			b.Fatal(err)
		}
	}
}
