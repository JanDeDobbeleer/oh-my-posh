package image

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetOutputPath(t *testing.T) {
	cases := []struct {
		Case     string
		Config   string
		Path     string
		Expected string
	}{
		{Case: "default config", Expected: "prompt.png"},
		{Case: "hidden file", Config: ".posh.omp.json", Expected: "posh.png"},
		{Case: "hidden file toml", Config: ".posh.omp.toml", Expected: "posh.png"},
		{Case: "hidden file yaml", Config: ".posh.omp.yaml", Expected: "posh.png"},
		{Case: "hidden file yml", Config: ".posh.omp.yml", Expected: "posh.png"},
		{Case: "path provided", Path: "mytheme.png", Expected: "mytheme.png"},
		{Case: "relative, no omp", Config: "~/jandedobbeleer.json", Expected: "jandedobbeleer.png"},
		{Case: "relative path", Config: "~/jandedobbeleer.omp.json", Expected: "jandedobbeleer.png"},
		{Case: "invalid config name", Config: "~/jandedobbeleer.omp.foo", Expected: "prompt.png"},
	}

	for _, tc := range cases {
		image := &Renderer{
			Path: tc.Path,
		}

		image.setOutputPath(tc.Config)

		assert.Equal(t, tc.Expected, image.Path, tc.Case)
	}
}
