// Derived from https://github.com/Crosse/font-install
// Copyright 2020 Seth Wright <seth@crosse.org>
package font

import (
	"archive/zip"
	"bytes"
	"io"
	"path"
	"strings"
)

func contains[S ~[]E, E comparable](s S, e E) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func InstallZIP(data []byte, user bool) ([]string, error) {
	var families []string
	bytesReader := bytes.NewReader(data)

	zipReader, err := zip.NewReader(bytesReader, int64(bytesReader.Len()))
	if err != nil {
		return families, err
	}

	fonts := make(map[string]*Font)

	for _, zf := range zipReader.File {
		rc, err := zf.Open()
		if err != nil {
			return families, err
		}
		defer rc.Close()

		data, err := io.ReadAll(rc)
		if err != nil {
			return families, err
		}

		fontData, err := newFont(zf.Name, data)
		if err != nil {
			continue
		}

		if _, found := fonts[fontData.Name]; !found {
			fonts[fontData.Name] = fontData
		} else {
			// prefer OTF over TTF; otherwise prefer the first font we find
			first := strings.ToLower(path.Ext(fonts[fontData.Name].FileName))
			second := strings.ToLower(path.Ext(fontData.FileName))
			if first != second && second == ".otf" {
				fonts[fontData.Name] = fontData
			}
		}
	}

	for _, font := range fonts {
		if err = install(font, user); err != nil {
			return families, err
		}

		if !contains(families, font.Family) {
			families = append(families, font.Family)
		}
	}

	return families, nil
}
