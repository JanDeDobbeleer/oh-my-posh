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

func install(font *Font, admin bool) (err error) {
	// To install a font on Windows:
	//  - Copy the file to the fonts directory
	//  - Add registry entry
	//  - Call AddFontResourceW to set the font
	// -  Notify other applications that the fonts have changed
	fontsDir := filepath.Join(os.Getenv("WINDIR"), "Fonts")
	if !admin {
		fontsDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Microsoft", "Windows", "Fonts")
	}

	fullPath := filepath.Join(fontsDir, font.FileName)
	// validate if font is already installed, remove it in case it is
	if _, err := os.Stat(fullPath); err == nil {
		if err = os.Remove(fullPath); err != nil {
			return fmt.Errorf("Unable to remove existing font file: %s", err.Error())
		}
	}
	err = os.WriteFile(fullPath, font.Data, 0644)
	if err != nil {
		return fmt.Errorf("Unable to write font file: %s", err.Error())
	}

	// Add registry entry
	reg := registry.LOCAL_MACHINE
	if !admin {
		reg = registry.CURRENT_USER
	}

	k, err := registry.OpenKey(reg, `SOFTWARE\Microsoft\Windows NT\CurrentVersion\Fonts`, registry.WRITE)
	if err != nil {
		// If this fails, remove the font file as well.
		if nexterr := os.Remove(fullPath); nexterr != nil {
			return errors.New("Unable to delete font file after registry key open error")
		}

		return fmt.Errorf("Unable to open registry key: %s", err.Error())
	}
	defer k.Close()

	name := fmt.Sprintf("%v (TrueType)", font.Name)
	if err = k.SetStringValue(name, font.FileName); err != nil {
		// If this fails, remove the font file as well.
		if nexterr := os.Remove(fullPath); nexterr != nil {
			return errors.New("Unable to delete font file after registry key set error")
		}

		return fmt.Errorf("Unable to set registry value: %s", err.Error())
	}

	gdi32 := syscall.NewLazyDLL("gdi32.dll")
	proc := gdi32.NewProc("AddFontResourceW")

	fontPtr, err := syscall.UTF16PtrFromString(fullPath)
	if err != nil {
		return
	}

	ret, _, _ := proc.Call(uintptr(unsafe.Pointer(fontPtr)))
	if ret == 0 {
		return errors.New("Unable to add font resource using AddFontResourceW")
	}

	return nil
}
