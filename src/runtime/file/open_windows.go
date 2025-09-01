package file

import (
	"os"

	"golang.org/x/sys/windows"
)

func Open(path string) (*os.File, error) {
	// GENERIC_READ, FILE_SHARE_READ
	handle, err := windows.CreateFile(
		windows.StringToUTF16Ptr(path),
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)

	if err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(handle), path), nil
}
