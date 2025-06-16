//go:build !windows && !darwin

// Derived from https://github.com/Crosse/font-install
// Copyright 2020 Seth Wright <seth@crosse.org>
package font

import (
	"os"
	"path"
	"strings"
)

var (
	fontsDir       = path.Join(os.Getenv("HOME"), "/.local/share/fonts")
	systemFontsDir = "/usr/share/fonts"
)

func install(font *Font, _ bool) error {
	// If we're running as root, install the font system-wide.
	targetDir := fontsDir
	if os.Geteuid() == 0 {
		targetDir = systemFontsDir
	}

	// On Linux, fontconfig can understand subdirectories. So, to keep the
	// font directory clean, install all font files for a particular font
	// family into a subdirectory named after the family (with hyphens instead
	// of spaces).
	fullPath := path.Join(targetDir,
		strings.ToLower(strings.ReplaceAll(font.Family, " ", "-")),
		path.Base(font.FileName))

	if err := os.MkdirAll(path.Dir(fullPath), 0700); err != nil {
		return err
	}

	return os.WriteFile(fullPath, font.Data, 0644)
}
