//go:build linux

// Derived from https://github.com/Crosse/font-install
// Copyright 2020 Seth Wright <seth@crosse.org>
package font

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// FontsDir denotes the path to the user's fonts directory on Unix-like systems.
var FontsDir = path.Join(os.Getenv("HOME"), "/.local/share/fonts")

func install(font *Font) (err error) {
	// On Linux, fontconfig can understand subdirectories. So, to keep the
	// font directory clean, install all font files for a particular font
	// family into a subdirectory named after the family (with hyphens instead
	// of spaces).
	fullPath := path.Join(FontsDir,
		strings.ToLower(strings.ReplaceAll(font.Family, " ", "-")),
		path.Base(font.FileName))

	if err = os.MkdirAll(path.Dir(fullPath), 0700); err != nil {
		return err
	}

	return ioutil.WriteFile(fullPath, font.Data, 0644)
}
