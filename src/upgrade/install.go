package upgrade

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func install(tag string) error {
	setState(validating)

	executable, err := os.Executable()
	if err != nil {
		return err
	}

	setState(downloading)

	data, err := downloadAndVerify(tag)
	if err != nil {
		return err
	}

	targetDir := filepath.Dir(executable)
	fileName := filepath.Base(executable)

	newPath := filepath.Join(targetDir, fmt.Sprintf(".%s.new", fileName))
	fp, err := os.OpenFile(newPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0775)
	if err != nil {
		return err
	}

	defer fp.Close()

	_, err = io.Copy(fp, bytes.NewReader(data))
	if err != nil {
		return err
	}

	// windows will have a lock when we do not close the file
	fp.Close()

	oldPath := filepath.Join(targetDir, fmt.Sprintf(".%s.old", fileName))

	_ = os.Remove(oldPath)

	err = os.Rename(executable, oldPath)
	if err != nil {
		return err
	}

	err = os.Rename(newPath, executable)

	if err != nil {
		// rollback
		rerr := os.Rename(oldPath, executable)
		if rerr != nil {
			return rerr
		}

		return err
	}

	removeErr := os.Remove(oldPath)

	// hide the old executable if we can't remove it
	if removeErr != nil {
		_ = hideFile(oldPath)
	}

	return nil
}
