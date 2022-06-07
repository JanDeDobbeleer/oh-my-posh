// Derived from https://github.com/Crosse/font-install
// Copyright 2020 Seth Wright <seth@crosse.org>
package font

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path"
	"runtime"
	"strings"
)

func Install(font string) error {
	location := fmt.Sprintf("https://github.com/ryanoasis/nerd-fonts/releases/latest/download/%s.zip", font)
	zipFile, err := Download(location)
	if err != nil {
		return err
	}
	return InstallZIP(zipFile)
}

func InstallZIP(data []byte) (err error) {
	bytesReader := bytes.NewReader(data)

	zipReader, err := zip.NewReader(bytesReader, int64(bytesReader.Len()))
	if err != nil {
		return
	}

	fonts := make(map[string]*Font)

	for _, zf := range zipReader.File {
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		data, err := io.ReadAll(rc)
		if err != nil {
			return err
		}

		fontData, err := new(zf.Name, data)
		if err != nil {
			continue
		}

		if _, ok := fonts[fontData.Name]; !ok {
			fonts[fontData.Name] = fontData
		} else {
			// Prefer OTF over TTF; otherwise prefer the first font we found.
			first := strings.ToLower(path.Ext(fonts[fontData.Name].FileName))
			second := strings.ToLower(path.Ext(fontData.FileName))
			if first != second && second == ".otf" {
				fonts[fontData.Name] = fontData
			}
		}
	}

	for _, font := range fonts {
		if !shouldInstall(font.Name) {
			continue
		}

		// print("Installing %s", font.Name)
		if err = install(font); err != nil {
			return err
		}
	}

	return nil
}

func shouldInstall(name string) bool {
	name = strings.ToLower(name)
	switch runtime.GOOS {
	case "windows":
		return strings.Contains(name, "windows compatible")
	default:
		return !strings.Contains(name, "windows compatible")
	}
}
