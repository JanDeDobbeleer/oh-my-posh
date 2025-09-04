//go:build !windows

package cache

import (
	"io"
	"os"
)

func openFile(filePath string) (io.ReadWriteCloser, error) {
	return os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0o644)
}
