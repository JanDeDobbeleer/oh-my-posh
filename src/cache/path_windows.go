package cache

import (
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

func platformCachePath() (string, bool) {
	if pfn, OK := PackageFamilyName(); OK {
		// WINDOWS MSIX cache folder, will only be present when oh-my-posh is installed via MSIX
		msixLocalAppData := filepath.Join(os.Getenv("LOCALAPPDATA"), "Packages", pfn, "LocalCache", "Local")
		if cachePath, OK := returnOrBuildCachePath(msixLocalAppData); OK {
			return cachePath, true
		}
	}

	// WINDOWS cache folder, should not exist elsewhere
	if cachePath, OK := returnOrBuildCachePath(os.Getenv("LOCALAPPDATA")); OK {
		return cachePath, true
	}

	return "", false
}

func PackageFamilyName() (string, bool) {
	log.Trace(time.Now(), "PackageFamilyName")

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procGetCurrentPackageFamilyName := kernel32.NewProc("GetCurrentPackageFamilyName")

	var length uint32 = 256
	buf := make([]uint16, length)
	ret, _, _ := procGetCurrentPackageFamilyName.Call(
		uintptr(unsafe.Pointer(&length)),
		uintptr(unsafe.Pointer(&buf[0])),
	)

	if ret != 0 {
		log.Debug("failed to get PackageFamilyName")
		return "", false
	}

	pfn := syscall.UTF16ToString(buf)
	log.Debug("PackageFamilyName:", pfn)

	return pfn, true
}
