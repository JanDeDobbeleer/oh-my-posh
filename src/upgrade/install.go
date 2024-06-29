package upgrade

import (
	"fmt"
	"os"
	"runtime"

	"github.com/inconshreveable/go-update"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
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

	data, err := downloadAsset(asset)
	if err != nil {
		return err
	}

	defer data.Close()

	return update.Apply(data, update.Options{
		TargetPath: executable,
	})
}
