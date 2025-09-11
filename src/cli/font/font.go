// Derived from https://github.com/Crosse/font-install
// Copyright 2020 Seth Wright <seth@crosse.org>
package font

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"path"
	"strings"

	"github.com/ConradIrwin/font/sfnt"
)

func init() {
	gob.Register([]*Font{})
	gob.Register([]*Asset{})
}

// Font describes a font file and the various metadata associated with it.
type Font struct {
	Name     string                 `json:"name,omitempty" jsonschema:"title=Font name,description=The name of the font"`
	Family   string                 `json:"-"`
	FileName string                 `json:"-"`
	Metadata map[sfnt.NameID]string `json:"-"`
	Data     []byte                 `json:"-"`
}

func (f *Font) Apply() error {
	asset, err := ResolveFontAsset(f.Name)
	if err != nil {
		return err
	}

	zipFile, err := Download(asset.URL)
	if err != nil {
		return err
	}

	_, err = InstallZIP(zipFile, asset.Folder)
	return err
}

func (f *Font) Equal(font *Font) bool {
	if font == nil {
		return false
	}

	return f.Name == font.Name
}

func (f *Font) Resolve() (*Font, bool) {
	return nil, false
}

// fontExtensions is a list of file extensions that denote fonts.
// Only files ending with these extensions will be installed.
var fontExtensions = map[string]bool{
	".otf": true,
	".ttf": true,
}

// newFont creates a newFont Font struct.
// fileName is the font's file name, and data is a byte slice containing the font file data.
// It returns a FontData struct describing the font, or an error.
func newFont(fileName string, data []byte) (*Font, error) {
	if _, ok := fontExtensions[strings.ToLower(path.Ext(fileName))]; !ok {
		return nil, fmt.Errorf("not a font: %v", fileName)
	}

	font := &Font{
		FileName: fileName,
		Metadata: make(map[sfnt.NameID]string),
		Data:     data,
	}

	fontData, err := sfnt.Parse(bytes.NewReader(font.Data))
	if err != nil {
		return nil, err
	}

	if !fontData.HasTable(sfnt.TagName) {
		return nil, fmt.Errorf("font %v has no name table", fileName)
	}

	nameTable, err := fontData.NameTable()
	if err != nil {
		return nil, err
	}

	for _, nameEntry := range nameTable.List() {
		font.Metadata[nameEntry.NameID] = nameEntry.String()
	}

	font.Name = font.Metadata[sfnt.NameFull]
	font.Family = font.Metadata[sfnt.NamePreferredFamily]

	if font.Family == "" {
		if v, ok := font.Metadata[sfnt.NameFontFamily]; ok {
			font.Family = v
		}
	}

	if font.Name == "" {
		font.Name = fileName
	}

	return font, nil
}
