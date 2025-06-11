package image

import (
	"os"
	"path/filepath"
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

func Test_loadFonts(t *testing.T) {
	pkgDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	fileName := "OcodoMonoNerdFont-Light.ttf"
	fontFilePath := filepath.Join(pkgDir, "testdata", fileName)

	testCases := []struct {
		name       string
		setEnvVars func()
		wantErr    bool
		errMessage string
	}{
		{
			name: "with valid POSH_FONT environment variables expect valid font(s)",
			setEnvVars: func() {
				os.Setenv("POSH_FONT_REGULAR", fontFilePath)
				os.Setenv("POSH_FONT_BOLD", fontFilePath)
				os.Setenv("POSH_FONT_ITALIC", fontFilePath)
			},
			wantErr: false,
		},
		{
			name: "without POSH_FONT environment variables expect default behavior",
			setEnvVars: func() {
				os.Unsetenv("POSH_FONT_REGULAR")
				os.Unsetenv("POSH_FONT_BOLD")
				os.Unsetenv("POSH_FONT_ITALIC")
			},
			wantErr: false,
		},
		{
			name: "with invalid/empty font file expect error",
			setEnvVars: func() {
				os.Setenv("POSH_FONT_REGULAR", "./testdata/not-a-NerdFont.ttf")
				os.Setenv("POSH_FONT_BOLD", "./testdata/not-a-NerdFont.ttf")
				os.Setenv("POSH_FONT_ITALIC", "./testdata/not-a-NerdFont.ttf")
			},
			wantErr:    true,
			errMessage: "failed to load regular font: font [./testdata/not-a-NerdFont.ttf] could not be parsed",
		},
		{
			name: "with invalid font name expect error",
			setEnvVars: func() {
				os.Setenv("POSH_FONT_REGULAR", "./testdata/not-a-font.ttf")
				os.Setenv("POSH_FONT_BOLD", "./testdata/not-a-font.ttf")
				os.Setenv("POSH_FONT_ITALIC", "./testdata/not-a-font.ttf")
			},
			wantErr:    true,
			errMessage: "failed to load regular font: filename [./testdata/not-a-font.ttf] should contain NerdFont",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setEnvVars()

			r := &Renderer{}
			err := r.loadFonts()

			if tc.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.errMessage)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
