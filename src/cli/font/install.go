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

func InstallZIP(data []byte, m *main) ([]string, error) {
	var families []string
	bytesReader := bytes.NewReader(data)

	zipReader, err := zip.NewReader(bytesReader, int64(bytesReader.Len()))
	if err != nil {
		return families, err
	}

	fonts := make(map[string]*Font)

	for _, file := range zipReader.File {
		// prevent zipslip attacks
		// https://security.snyk.io/research/zip-slip-vulnerability
		// skip folders
		if strings.Contains(file.Name, "..") || strings.HasSuffix(file.Name, "/") {
			continue
		}

		fontFileName := path.Base(file.Name)
		fontRelativeFileName := strings.TrimPrefix(file.Name, m.Folder)

		// do not install fonts that are not in the specified installation folder
		if fontFileName != fontRelativeFileName {
			continue
		}

		fontReader, err := file.Open()
		if err != nil {
			continue
		}

		defer fontReader.Close()

		fontBytes, err := io.ReadAll(fontReader)
		if err != nil {
			continue
		}

		font, err := newFont(fontFileName, fontBytes)
		if err != nil {
			continue
		}

		if _, found := fonts[font.Name]; !found {
			fonts[font.Name] = font
			continue
		}

		// prefer .ttf files over other file types when we have a duplicate
		first := strings.ToLower(path.Ext(fonts[font.Name].FileName))
		second := strings.ToLower(path.Ext(font.FileName))
		if first != second && second == ".ttf" {
			fonts[font.Name] = font
		}
	}

	for _, font := range fonts {
		if err = install(font, m.system); err != nil {
			continue
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
