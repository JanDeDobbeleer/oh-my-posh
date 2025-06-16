package image

import (
	"fmt"
	stdOS "os"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	regular = "regular"
)

type Fonts struct {
	Regular string `json:"regular"`
	Bold    string `json:"bold"`
	Italic  string `json:"italic"`
}

func (f *Fonts) IsValid() bool {
	if f == nil {
		return false
	}

	// Check that all required font paths are non-empty
	return f.Regular != "" && f.Bold != "" && f.Italic != ""
}

func (f *Fonts) Load() (map[string]font.Face, error) {
	defer log.Trace(time.Now(), "Load fonts")

	result := make(map[string]font.Face)

	fonts := map[string]Font{
		regular: Font(f.Regular),
		bold:    Font(f.Bold),
		italic:  Font(f.Italic),
	}

	for name, fontPath := range fonts {
		fontFace, err := fontPath.Load()
		if err != nil {
			return nil, fmt.Errorf("failed to load font %s: %w", fontPath, err)
		}

		result[name] = fontFace
	}

	return result, nil
}

type Font string

func (f Font) Load() (font.Face, error) {
	defer log.Trace(time.Now(), string(f))

	data, err := stdOS.ReadFile(string(f))
	if err != nil {
		return nil, fmt.Errorf("failed to read font file %s: %w", f, err)
	}

	fontObject, err := opentype.Parse(data)

	// handle collections
	if err != nil {
		collection, err := opentype.ParseCollection(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse font %s as single font or collection: %w", f, err)
		}

		if collection.NumFonts() == 0 {
			return nil, fmt.Errorf("font collection %s is empty", f)
		}

		fontObject, err = collection.Font(0)
		if err != nil {
			return nil, fmt.Errorf("failed to get first font from collection %s: %w", f, err)
		}
	}

	face, err := opentype.NewFace(fontObject, &opentype.FaceOptions{Size: 2.0 * 12, DPI: 144})
	if err != nil {
		return nil, fmt.Errorf("failed to create font face for %s: %w", f, err)
	}

	if face == nil {
		return nil, fmt.Errorf("failed to create font face for %s: face is nil", f)
	}

	return face, nil
}
