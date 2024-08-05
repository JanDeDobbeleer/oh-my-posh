package upgrade

import (
	"errors"
	"path/filepath"
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

func updateRegistry(version, executable string) {
	// needs to be the parent directory of the executable's bin directory
	// with a trailing backslash to match the registry key
	// in case this wasn't installed with the installer, nothing will match
	// and we don't set the registry keys
	installLocation := filepath.Dir(executable)
	installLocation = filepath.Dir(installLocation)
	installLocation += `\`

	key, err := getRegistryKey(installLocation)
	if err != nil {
		key.Close()
		return
	}

	version = strings.TrimLeft(version, "v")

	_ = key.SetStringValue("DisplayVersion", version)
	_ = key.SetStringValue("DisplayName", "Oh My Posh")

	splitted := strings.Split(version, ".")
	if len(splitted) < 3 {
		key.Close()
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

	key.Close()
}

// getRegistryKey tries all known registry paths to find the one we need to adjust (if any)
func getRegistryKey(installLocation string) (registry.Key, error) {
	knownRegistryPaths := []struct {
		Path string
		Key  registry.Key
	}{
		{`Software\Microsoft\Windows\CurrentVersion\Uninstall`, registry.CURRENT_USER},
		{`Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`, registry.CURRENT_USER},
		{`Software\Microsoft\Windows\CurrentVersion\Uninstall`, registry.LOCAL_MACHINE},
		{`Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`, registry.LOCAL_MACHINE},
	}

	for _, path := range knownRegistryPaths {
		key, ok := tryRegistryKey(path.Key, path.Path, installLocation)
		if ok {
			return key, nil
		}
	}

	return registry.CURRENT_USER, errors.New("could not find registry key")
}

// tryRegistryKey tries to open the registry key for the given path
// and checks if the install location matches with the current executable's location
func tryRegistryKey(key registry.Key, path, installLocation string) (registry.Key, bool) {
	path += `\Oh My Posh_is1`

	readKey, err := registry.OpenKey(key, path, registry.READ)
	if err != nil {
		return key, false
	}

	location, _, err := readKey.GetStringValue("InstallLocation")
	if err != nil {
		return key, false
	}

	if location != installLocation {
		return key, false
	}

	key, err = registry.OpenKey(key, path, registry.WRITE)
	return key, err == nil
}
