package engine

import (
	"io/ioutil" //nolint:staticcheck,nolintlint
	"os"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/color"
	"github.com/jandedobbeleer/oh-my-posh/shell"

	"github.com/stretchr/testify/assert"
)

var cases = []struct {
	Case   string
	Config string
}{
	{Case: ".omp.json suffix", Config: "~/jandedobbeleer.omp.json"},
	{Case: ".omp.yaml suffix", Config: "~/jandedobbeleer.omp.yaml"},
	{Case: ".omp.yml suffix", Config: "~/jandedobbeleer.omp.yml"},
	{Case: ".omp.toml suffix", Config: "~/jandedobbeleer.omp.toml"},
	{Case: ".json suffix", Config: "~/jandedobbeleer.json"},
	{Case: ".yaml suffix", Config: "~/jandedobbeleer.yaml"},
	{Case: ".yml suffix", Config: "~/jandedobbeleer.yml"},
	{Case: ".toml suffix", Config: "~/jandedobbeleer.toml"},
}

func runImageTest(config, content string) (string, error) {
	poshImagePath := "jandedobbeleer.png"
	file, err := ioutil.TempFile("", poshImagePath)
	if err != nil {
		return "", err
	}
	defer os.Remove(file.Name())
	ansi := &color.AnsiWriter{}
	ansi.Init(shell.GENERIC)
	image := &ImageRenderer{
		AnsiString: content,
		Ansi:       ansi,
	}
	image.Init(config)
	err = image.SavePNG()
	if err == nil {
		os.Remove(image.Path)
	}
	return filepath.Base(image.Path), err
}

func TestStringImageFileWithText(t *testing.T) {
	for _, tc := range cases {
		filename, err := runImageTest(tc.Config, "foobar")
		assert.Equal(t, "jandedobbeleer.png", filename, tc.Case)
		assert.NoError(t, err)
	}
}

func TestStringImageFileWithANSI(t *testing.T) {
	prompt := `[38;2;40;105;131m[0m[48;2;40;105;131m[38;2;224;222;244m jan [0m[38;2;40;105;131m[0m[38;2;224;222;244m [0m`
	for _, tc := range cases {
		filename, err := runImageTest(tc.Config, prompt)
		assert.Equal(t, "jandedobbeleer.png", filename, tc.Case)
		assert.NoError(t, err)
	}
}
