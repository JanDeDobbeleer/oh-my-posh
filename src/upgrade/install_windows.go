package upgrade

import (
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
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

func updateRegistry(version string) {
	key, err := getRegistryKey()
	if err != nil {
		return
	}

	version = strings.TrimLeft(version, "v")

	_ = key.SetStringValue("DisplayVersion", version)
	_ = key.SetStringValue("DisplayName", "Oh My Posh")

	splitted := strings.Split(version, ".")
	if len(splitted) < 3 {
		return
	}

	if u64, err := strconv.ParseUint(splitted[0], 10, 32); err == nil {
		major := uint32(u64)
		_ = key.SetDWordValue("MajorVersion", major)
		_ = key.SetDWordValue("VersionMajor", major)
	}

	if u64, err := strconv.ParseUint(splitted[1], 10, 32); err == nil {
		minor := uint32(u64)
		_ = key.SetDWordValue("MinorVersion", minor)
		_ = key.SetDWordValue("VersionMinor", minor)
	}
}

func getRegistryKey() (registry.Key, error) {
	path := `Software\Microsoft\Windows\CurrentVersion\Uninstall\Oh My Posh_is1`

	key, err := registry.OpenKey(registry.CURRENT_USER, path, registry.WRITE)
	if err == nil {
		return key, nil
	}

	return registry.OpenKey(registry.LOCAL_MACHINE, path, registry.WRITE)
}
