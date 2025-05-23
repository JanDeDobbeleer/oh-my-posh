package upgrade

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

func install(cfg *Config) error {
	setState(validating)

	executable, err := os.Executable()
	if err != nil {
		log.Debug("failed to get executable path")
		return err
	}

	targetDir := filepath.Dir(executable)
	fileName := filepath.Base(executable)

	newPath := filepath.Join(targetDir, fmt.Sprintf(".%s.new", fileName))
	fp, err := os.OpenFile(newPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0775)
	if err != nil {
		log.Error(err)
		return errors.New("we do not have permissions to update")
	}

	setState(downloading)

	data, err := downloadAndVerify(cfg)
	if err != nil {
		log.Debug("failed to download and verify")
		return err
	}

	setState(installing)

	_, err = io.Copy(fp, bytes.NewReader(data))
	// windows will have a lock when we do not close the file
	fp.Close()

	if err != nil {
		log.Debug("failed to copy data to new file")
		return err
	}

	oldPath := filepath.Join(targetDir, fmt.Sprintf(".%s.old", fileName))

	_ = os.Remove(oldPath)

	err = os.Rename(executable, oldPath)
	if err != nil {
		log.Debug("failed to rename old file")
		return err
	}

	err = os.Rename(newPath, executable)

	if err != nil {
		log.Debug("failed to rename new file, rolling back")
		// rollback
		rerr := os.Rename(oldPath, executable)
		if rerr != nil {
			log.Debug("failed to rollback old file")
			return rerr
		}

		return err
	}

	removeErr := os.Remove(oldPath)

	// hide the old executable if we can't remove it
	if removeErr != nil {
		log.Error(removeErr)
		// hide the old executable
		_ = hideFile(oldPath)
	}

	return nil
}
