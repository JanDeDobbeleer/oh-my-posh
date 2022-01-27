package engine

import (
	"io/ioutil"
	"oh-my-posh/color"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func runImageTest(content string) error {
	poshImagePath := "jandedobbeleer.png"
	file, err := ioutil.TempFile("", poshImagePath)
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	ansi := &color.Ansi{}
	ansi.Init(plain)
	image := &ImageRenderer{
		AnsiString: content,
		Ansi:       ansi,
	}
	image.Init("~/jandedobbeleer.omp.json")
	err = image.SavePNG()
	return err
}

func TestStringImageFileWithText(t *testing.T) {
	err := runImageTest("foobar")
	assert.NoError(t, err)
}

func TestStringImageFileWithANSI(t *testing.T) {
	prompt := `[38;2;0;55;218;49m[7m\uE0B0[m[0m[48;2;0;55;218m[38;2;255;255;255m oh-my-posh
	 [0m[48;2;193;156;0m[38;2;0;55;218m\uE0B0[0m[48;2;193;156;0m[38;2;17;17;17m main ≡  ~4 -8 ?7 [0m[38;2;193;156;0m\uE0B0[0m
	[37m [0m[0m`
	err := runImageTest(prompt)
	assert.NoError(t, err)
}
