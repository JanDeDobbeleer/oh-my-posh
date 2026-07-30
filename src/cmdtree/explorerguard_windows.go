//go:build windows

package cmdtree

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// checkExplorerLaunch detects a double-click launch from Explorer: there is
// no console to read the output, so print an explanation, linger, and exit.
func checkExplorerLaunch() {
	if ExplorerLaunchHelpText == "" || !startedByExplorer() {
		return
	}

	fmt.Println(ExplorerLaunchHelpText)
	time.Sleep(5 * time.Second)
	os.Exit(1)
}

func startedByExplorer() bool {
	pid := windows.GetCurrentProcessId()

	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return false
	}

	defer func() { _ = windows.CloseHandle(snapshot) }()

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	var parentID uint32

	for err = windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		if entry.ProcessID == pid {
			parentID = entry.ParentProcessID
			break
		}
	}

	if parentID == 0 {
		return false
	}

	for err = windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		if entry.ProcessID != parentID {
			continue
		}

		name := windows.UTF16ToString(entry.ExeFile[:])
		return strings.EqualFold(name, "explorer.exe")
	}

	return false
}
