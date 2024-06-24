package upgrade

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
)

var successMsg = "Oh My Posh is installing in the background.\nRestart your shell in a minute to take full advantage of the new functionality."

func install() error {
	setState(downloading)

	temp := os.Getenv("TEMP")
	if len(temp) == 0 {
		return errors.New("failed to get TEMP environment variable")
	}

	id := uuid.New().String()
	fileName := fmt.Sprintf("oh-my-posh-install-%s.exe", id)
	path := filepath.Join(temp, fileName)

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
