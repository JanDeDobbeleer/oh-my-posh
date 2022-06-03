// Derived from https://github.com/Crosse/font-install
// Copyright 2020 Seth Wright <seth@crosse.org>
package font

import (
	"os"
	"path"
)

var FontsDir = path.Join(os.Getenv("HOME"), "Library", "Fonts")

func install(font *Font) (err error) {
	// On darwin/OSX, the user's fonts directory is ~/Library/Fonts,
	// and fonts should be installed directly into that path;
	// i.e., not in subfolders.
	fullPath := path.Join(FontsDir, path.Base(font.FileName))

	if err = os.MkdirAll(path.Dir(fullPath), 0700); err != nil {
		return
	}

	err = os.WriteFile(fullPath, font.Data, 0644)

	return
}
