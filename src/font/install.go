// Derived from https://github.com/Crosse/font-install
// Copyright 2020 Seth Wright <seth@crosse.org>
package font

import (
	"archive/zip"
	"bytes"
	"io"
	"path"
	stdruntime "runtime"
	"slices"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/cmd"
)

func contains[S ~[]E, E comparable](s S, e E) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

func InstallZIP(data []byte, user, ttf bool) ([]string, error) {
	// prefer OTF over TTF; otherwise prefer the first font we find
	extension := ".otf"
	if ttf {
		extension = ".ttf"
	}

	var families []string
	bytesReader := bytes.NewReader(data)

	zipReader, err := zip.NewReader(bytesReader, int64(bytesReader.Len()))
	if err != nil {
		return families, err
	}

	fonts := make(map[string]*Font)

	for _, zf := range zipReader.File {
		// prevent zipslip attacks
		// https://security.snyk.io/research/zip-slip-vulnerability
		if strings.Contains(zf.Name, "..") {
			continue
		}

		rc, err := zf.Open()
		if err != nil {
			return families, err
		}

		defer rc.Close()

		data, err := io.ReadAll(rc)
		if err != nil {
			return families, err
		}

		fontData, err := newFont(path.Base(zf.Name), data)
		if err != nil {
			continue
		}

		if _, found := fonts[fontData.Name]; !found {
			fonts[fontData.Name] = fontData
			continue
		}

		// respect the user's preference for TTF or OTF
		first := strings.ToLower(path.Ext(fonts[fontData.Name].FileName))
		second := strings.ToLower(path.Ext(fontData.FileName))
		if first != second && second == extension {
			fonts[fontData.Name] = fontData
		}
	}

	for _, font := range fonts {
		if err = install(font, user); err != nil {
			return families, err
		}

		if found := contains(families, font.Family); !found {
			families = append(families, font.Family)
		}
	}

	// Update the font cache when installing fonts on Linux
	if stdruntime.GOOS == runtime.LINUX || stdruntime.GOOS == runtime.DARWIN {
		_, _ = cmd.Run("fc-cache", "-f")
	}

	slices.Sort(families)

	return families, nil
}
