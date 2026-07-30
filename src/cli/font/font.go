// Derived from https://github.com/Crosse/font-install
// Copyright 2020 Seth Wright <seth@crosse.org>
package font

import (
	"encoding/gob"
	"fmt"
	"path"
	"strings"
)

func init() {
	gob.Register([]*Font{})
	gob.Register([]*Asset{})
}

type Font struct {
	Name     string                 `json:"name,omitempty" jsonschema:"title=Font name,description=The name of the font"`
	Family   string                 `json:"-"`
	FileName string                 `json:"-"`
	Metadata map[nameID]string `json:"-"`
	Data     []byte                 `json:"-"`
}

func (f *Font) Apply() error {
	_, err := downloadAndInstall(f.Name, "")
	return err
}

func downloadAndInstall(font, zipFolder string) (string, error) {
	asset, err := ResolveFontAsset(font)
	if err != nil {
		return "", err
	}

	if asset.Folder != "" && zipFolder == "" {
		zipFolder = asset.Folder
	}

	zipFile, err := Download(asset.URL)
	if err != nil {
		return "", err
	}

	_, err = InstallZIP(zipFile, zipFolder)
	return asset.Name, err
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

var fontExtensions = map[string]bool{
	".otf": true,
	".ttf": true,
}

func newFont(fileName string, data []byte) (*Font, error) {
	if _, ok := fontExtensions[strings.ToLower(path.Ext(fileName))]; !ok {
		return nil, fmt.Errorf("not a font: %v", fileName)
	}

	font := &Font{
		FileName: fileName,
		Metadata: make(map[nameID]string),
		Data:     data,
	}

	table, ok, err := readNameTable(font.Data)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("font %v has no name table", fileName)
	}

	entries, err := parseNameTable(table)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		font.Metadata[entry.nameID] = entry.String()
	}

	font.Name = font.Metadata[nameFull]
	font.Family = font.Metadata[namePreferredFamily]

	if font.Family == "" {
		if v, ok := font.Metadata[nameFontFamily]; ok {
			font.Family = v
		}
	}

	if font.Name == "" {
		font.Name = fileName
	}

	return font, nil
}
