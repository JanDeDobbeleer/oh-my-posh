package upgrade

import (
	"syscall"
	"unsafe"
)

func hideFile(path string) error {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setFileAttributes := kernel32.NewProc("SetFileAttributesW")

	ptr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	r1, _, err := setFileAttributes.Call(uintptr(unsafe.Pointer(ptr)), 2)

	if r1 == 0 {
		return err
	}

	return nil
}
