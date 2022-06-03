// Derived from https://github.com/Crosse/font-install
// Copyright 2020 Seth Wright <seth@crosse.org>
package font

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"golang.org/x/sys/windows/registry"
)

// FontsDir denotes the path to the user's fonts directory on Linux.
// Windows doesn't have the concept of a permanent, per-user collection
// of fonts, meaning that all fonts are stored in the system-level fonts
// directory, which is %WINDIR%\Fonts by default.
var FontsDir = path.Join(os.Getenv("WINDIR"), "Fonts")

func install(font *Font) (err error) {
	// To install a font on Windows:
	//  - Copy the file to the fonts directory
	//  - Create a registry entry for the font
	fullPath := path.Join(FontsDir, font.FileName)

	err = ioutil.WriteFile(fullPath, font.Data, 0644) //nolint:gosec
	if err != nil {
		return err
	}

	// Second, write metadata about the font to the registry.
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion\Fonts`, registry.WRITE)
	if err != nil {
		// If this fails, remove the font file as well.
		if nexterr := os.Remove(fullPath); nexterr != nil {
			return nexterr
		}

		return err
	}
	defer k.Close()

	// Apparently it's "ok" to mark an OpenType font as "TrueType",
	// and since this tool only supports True- and OpenType fonts,
	// this should be Okay(tm).
	// Besides, Windows does it, so why can't I?
	valueName := fmt.Sprintf("%v (TrueType)", font.FileName)
	if err = k.SetStringValue(font.Name, valueName); err != nil {
		// If this fails, remove the font file as well.
		if nexterr := os.Remove(fullPath); nexterr != nil {
			return nexterr
		}

		return err
	}

	return nil
}
