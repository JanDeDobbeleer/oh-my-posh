package upgrade

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
)

var successMsg = "🚀  Upgrade successful, restart your shell to take full advantage of the new functionality."

func install() error {
	setState(validating)
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(executable, os.O_WRONLY, 0755)
	if err != nil {
		return errors.New("we don't have permissions to upgrade oh-my-posh, please use elevated permissions to upgrade")
	}

	defer file.Close()

	setState(downloading)

	asset := fmt.Sprintf("posh-%s-%s", runtime.GOOS, runtime.GOARCH)

	data, err := downloadAsset(asset)
	if err != nil {
		return err
	}

	defer data.Close()

	setState(installing)

	_, err = io.Copy(file, data)
	if err != nil {
		return err
	}

	return nil
}
