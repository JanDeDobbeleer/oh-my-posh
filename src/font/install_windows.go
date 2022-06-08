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
var FontsDir = filepath.Join(os.Getenv("WINDIR"), "Fonts")

const (
	WM_FONTCHANGE  = 0x001D // nolint:revive
	HWND_BROADCAST = 0xFFFF // nolint:revive
)

func install(font *Font) (err error) {
	// To install a font on Windows:
	//  - Copy the file to the fonts directory
	//  - Add registry entry
	//  - Call AddFontResourceW to set the font
	// -  Notify other applications that the fonts have changed
	fullPath := filepath.Join(FontsDir, font.FileName)
	err = os.WriteFile(fullPath, font.Data, 0644)
	if err != nil {
		return
	}

	// Add registry entry
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion\Fonts`, registry.WRITE)
	if err != nil {
		// If this fails, remove the font file as well.
		if nexterr := os.Remove(fullPath); nexterr != nil {
			return nexterr
		}

		return err
	}
	defer k.Close()

	name := fmt.Sprintf("%v (TrueType)", font.Name)
	if err = k.SetStringValue(name, font.FileName); err != nil {
		// If this fails, remove the font file as well.
		if nexterr := os.Remove(fullPath); nexterr != nil {
			return nexterr
		}

		return err
	}

	gdi32 := syscall.NewLazyDLL("gdi32.dll")
	proc := gdi32.NewProc("AddFontResourceW")

	fontPtr, err := syscall.UTF16PtrFromString(fullPath)
	if err != nil {
		return
	}

	ret, _, _ := proc.Call(uintptr(unsafe.Pointer(fontPtr)))
	if ret == 0 {
		return errors.New("unable to add font resource")
	}

	// Notify other applications that the fonts have changed
	user32 := syscall.NewLazyDLL("user32.dll")
	procSendMessageW := user32.NewProc("SendMessageW")
	_, _, e1 := syscall.SyscallN(procSendMessageW.Addr(), uintptr(HWND_BROADCAST), uintptr(WM_FONTCHANGE), 0, 0)
	if e1 != 0 {
		return errors.New("unable to broadcast font change")
	}

	return nil
}
