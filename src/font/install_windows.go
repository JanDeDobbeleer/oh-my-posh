package font

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

// https://docs.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-addfontresourcea

const (
	WM_FONTCHANGE  = 0x001D //nolint:revive
	HWND_BROADCAST = 0xFFFF //nolint:revive
)

func install(font *Font, admin bool) error {
	// To install a font on Windows:
	//  - Copy the file to the fonts directory
	//  - Add registry entry
	//  - Call AddFontResourceW to set the font
	fontsDir := filepath.Join(os.Getenv("WINDIR"), "Fonts")
	if !admin {
		fontsDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Microsoft", "Windows", "Fonts")
	}

	// check if the Fonts folder exists, if not, create it
	if _, err := os.Stat(fontsDir); os.IsNotExist(err) {
		if err = os.MkdirAll(fontsDir, 0755); err != nil {
			return fmt.Errorf("unable to create fonts directory: %s", err.Error())
		}
	}

	fullPath := filepath.Join(fontsDir, font.FileName)
	// validate if the font is already installed, remove it in case it is
	if _, err := os.Stat(fullPath); err == nil {
		if err = os.Remove(fullPath); err != nil {
			return fmt.Errorf("unable to remove existing font file: %s", err.Error())
		}
	}

	err := os.WriteFile(fullPath, font.Data, 0644)
	if err != nil {
		return fmt.Errorf("unable to write font file: %s", err.Error())
	}

	// Add registry entry
	var reg = registry.LOCAL_MACHINE
	var regValue = font.FileName
	if !admin {
		reg = registry.CURRENT_USER
		regValue = fullPath
	}

	k, err := registry.OpenKey(reg, `SOFTWARE\Microsoft\Windows NT\CurrentVersion\Fonts`, registry.WRITE)
	if err != nil {
		// If this fails, remove the font file as well.
		if nexterr := os.Remove(fullPath); nexterr != nil {
			return errors.New("unable to delete font file after registry key open error")
		}

		return fmt.Errorf("unable to open registry key: %s", err.Error())
	}

	defer k.Close()

	fontName := fmt.Sprintf("%v (TrueType)", font.Name)
	var alreadyInstalled, newFontType bool

	// check if we already had this key set
	oldFullPath, _, err := k.GetStringValue(fontName)
	if err == nil {
		alreadyInstalled = true
		newFontType = oldFullPath != fullPath
	}

	// do not call AddFontResourceW if the font was already installed
	if alreadyInstalled && !newFontType {
		return nil
	}

	gdi32 := syscall.NewLazyDLL("gdi32.dll")
	addFontResourceW := gdi32.NewProc("AddFontResourceW")

	// remove the old font resource in case we have a new font type with the same name
	if newFontType {
		fontPtr, err := syscall.UTF16PtrFromString(oldFullPath)
		if err == nil {
			removeFontResourceW := gdi32.NewProc("RemoveFontResourceW")
			_, _, _ = removeFontResourceW.Call(uintptr(unsafe.Pointer(fontPtr)))
		}
	}

	if err = k.SetStringValue(fontName, regValue); err != nil {
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

	return nil
}
