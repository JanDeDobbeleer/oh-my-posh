package upgrade

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var successMsg = "Oh My Posh is installing in the background.\nRestart your shell in a minute to take full advantage of the new functionality."

func install() error {
	setState(downloading)

	temp := os.Getenv("TEMP")
	if len(temp) == 0 {
		return errors.New("failed to get TEMP environment variable")
	}

	path := filepath.Join(temp, "install.exe")

	if _, err := os.Stat(path); err == nil {
		err := os.Remove(path)
		if err != nil {
			return errors.New("unable to remove existing installer")
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	asset := fmt.Sprintf("install-%s.exe", runtime.GOARCH)

	data, err := downloadAsset(asset)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, data)
	if err != nil {
		return err
	}

	data.Close()
	file.Close()

	// We need to run the installer in a separate process to avoid being blocked by the current process
	go func() {
		cmd := exec.Command(path, "/VERYSILENT", "/FORCECLOSEAPPLICATIONS")
		_ = cmd.Run()
	}()

	return nil
}
