package font

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"golang.org/x/sys/windows/registry"
)

// https://docs.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-addfontresourcea

const (
	WM_FONTCHANGE  = 0x001D //nolint:revive
	HWND_BROADCAST = 0xFFFF //nolint:revive
)

func install(font *Font) error {
	// To install a font on Windows:
	//  - Copy the file to the fonts directory
	//  - Add registry entry
	//  - Call AddFontResourceW to set the font
	fontsDir := filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Microsoft", "Windows", "Fonts")

	log.Debugf("installing font %s to %s", font.FileName, fontsDir)

	// check if the Fonts folder exists, if not, create it
	if _, err := os.Stat(fontsDir); os.IsNotExist(err) {
		if err = os.MkdirAll(fontsDir, 0755); err != nil {
			return fmt.Errorf("unable to create fonts directory: %s", err.Error())
		}
	}

	log.Debug("fonts directory exists, proceeding with installation")

	fullPath := filepath.Join(fontsDir, font.FileName)
	// validate if the font is already installed, remove it in case it is
	if _, err := os.Stat(fullPath); err == nil {
		log.Debugf("font %s already exists, removing it", fullPath)
		if err = os.Remove(fullPath); err != nil {
			return fmt.Errorf("unable to remove existing font file: %s", err.Error())
		}
	}

	log.Debugf("writing font file to %s", fullPath)

	err := os.WriteFile(fullPath, font.Data, 0644)
	if err != nil {
		return fmt.Errorf("unable to write font file: %s", err.Error())
	}

	log.Debug("font file written successfully, proceeding with registry entry")

	// Add registry entry
	reg := registry.CURRENT_USER
	regValue := fullPath

	log.Debug("opening HKEY_CURRENT_USER for writing (SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\\Fonts)")

	k, _, err := registry.CreateKey(reg, `SOFTWARE\Microsoft\Windows NT\CurrentVersion\Fonts`, registry.WRITE)
	if err != nil {
		log.Error(err)
		// If this fails, remove the font file as well.
		if nexterr := os.Remove(fullPath); nexterr != nil {
			log.Error(nexterr)
			return errors.New("unable to delete font file after registry key open error")
		}

		return errors.New("unable to open HKEY_CURRENT_USER")
	}

	defer func() {
		err := k.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	fontName := fmt.Sprintf("%v (TrueType)", font.Name)
	var alreadyInstalled, newFontType bool

	log.Debugf("validating if font %s is already installed", fontName)

	// check if we already had this key set
	oldFullPath, _, err := k.GetStringValue(fontName)
	if err == nil {
		log.Debugf("font %s is already installed with path %s", fontName, oldFullPath)
		alreadyInstalled = true
		newFontType = oldFullPath != fullPath
	}

	if !alreadyInstalled {
		log.Debug("font is not registered, adding to registry")
		if err := k.SetStringValue(fontName, fullPath); err != nil {
			return err
		}

		log.Debug("font registry entry added successfully")
	}

	// do not call AddFontResourceW if the font was already installed
	if alreadyInstalled && !newFontType {
		log.Debugf("font %s is already installed, skipping AddFontResourceW", fontName)
		return nil
	}

	gdi32 := syscall.NewLazyDLL("gdi32.dll")
	addFontResourceW := gdi32.NewProc("AddFontResourceW")

	// remove the old font resource in case we have a new font type with the same name
	if newFontType {
		log.Debug("removing old font resource before adding new one")
		fontPtr, err := syscall.UTF16PtrFromString(oldFullPath)
		if err == nil {
			removeFontResourceW := gdi32.NewProc("RemoveFontResourceW")
			_, _, _ = removeFontResourceW.Call(uintptr(unsafe.Pointer(fontPtr)))
		}
	}

	if err = k.SetStringValue(fontName, regValue); err != nil {
		log.Error(err)
		// If this fails, remove the font file as well.
		if nexterr := os.Remove(fullPath); nexterr != nil {
			return errors.New("unable to delete font file after registry key set error")
		}

		return fmt.Errorf("unable to set registry value: %s", err.Error())
	}

	fontPtr, err := syscall.UTF16PtrFromString(fullPath)
	if err != nil {
		return err
	}

	ret, _, _ := addFontResourceW.Call(uintptr(unsafe.Pointer(fontPtr)))
	if ret == 0 {
		return errors.New("unable to add font resource using AddFontResourceW")
	}

	log.Debug("font resource added successfully")

	return nil
}
