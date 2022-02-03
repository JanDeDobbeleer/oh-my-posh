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
	prompt := `[38;2;40;105;131m[0m[48;2;40;105;131m[38;2;224;222;244m jan [0m[38;2;40;105;131m[0m[38;2;224;222;244m [0m`
	err := runImageTest(prompt)
	assert.NoError(t, err)
}
