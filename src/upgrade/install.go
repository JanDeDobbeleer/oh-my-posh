package upgrade

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/platform/net"
)

func install() error {
	setState(validating)

	executable, err := os.Executable()
	if err != nil {
		return err
	}

	extension := ""
	if runtime.GOOS == platform.WINDOWS {
		extension = ".exe"
	}

	asset := fmt.Sprintf("posh-%s-%s%s", runtime.GOOS, runtime.GOARCH, extension)

	setState(downloading)

	url := fmt.Sprintf("https://github.com/JanDeDobbeleer/oh-my-posh/releases/latest/download/%s", asset)

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := net.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download installer: %s", url)
	}

	defer resp.Body.Close()

	newBytes, err := io.ReadAll(resp.Body)
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

	_, err = io.Copy(fp, bytes.NewReader(newBytes))
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
