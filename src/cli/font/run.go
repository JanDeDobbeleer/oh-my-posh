package font

import (
	"fmt"
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/cli/ui"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

// List returns the installable Nerd Fonts, newest release first, for `oh-my-posh font list`.
func List() ([]*Asset, error) {
	return fonts()
}

// Install downloads and installs one font by name, URL, or local zip path, reporting each step on
// a status line.
//
// This replaced a five-stage terminal UI - fetch the list, pick from it, download, unzip, install
// - that existed only so a name could be chosen interactively. `font list` answers that question
// better: it can be read, grepped, scripted, and piped, none of which a picker allows. What is
// left is a linear sequence, which is what it always was underneath.
func Install(name, zipFolder string) (string, error) {
	status := ui.NewStatus(os.Stdout)
	programStatus := ui.NewProgramStatus(os.Stdout)

	// A local zip is already on disk: nothing to resolve and nothing to download.
	if IsLocalZipFile(name) {
		data, err := os.ReadFile(name)
		if err != nil {
			programStatus.Error(err.Error())
			return "", err
		}

		message := fmt.Sprintf("Installing %s", name)
		programStatus.Working(message)
		status.Start(message)

		families, err := InstallZIP(data, zipFolder)
		if err != nil {
			programStatus.Error(err.Error())
			status.Stop("")
			return "", err
		}

		message = installed(name, families)
		programStatus.Done(message)
		status.Stop(message)

		return name, nil
	}

	programStatus.Working("Resolving font")
	status.Start("Resolving font")

	asset, err := ResolveFontAsset(name)
	if err != nil {
		programStatus.Error(err.Error())
		status.Stop("")
		return "", err
	}

	if asset.Folder != "" && zipFolder == "" {
		zipFolder = asset.Folder
	}

	status.Stop("")

	message := fmt.Sprintf("Downloading %s", asset.Name)
	programStatus.Progress(message, 0)
	bar := ui.NewProgress(os.Stdout, message)

	zipFile, err := download(asset.URL, func(fraction float64) {
		programStatus.Progress(message, int(fraction*100))
		bar.Set(fraction)
	})
	if err != nil {
		programStatus.Error(err.Error())
		bar.Done()
		return "", err
	}

	bar.Done()

	message = fmt.Sprintf("Installing %s", asset.Name)
	programStatus.Working(message)
	status.Start(message)

	families, err := InstallZIP(zipFile, zipFolder)
	if err != nil {
		programStatus.Error(err.Error())
		status.Stop("")
		return "", err
	}

	message = installed(asset.Name, families)
	programStatus.Done(message)
	status.Stop(message)

	return asset.Name, nil
}

// installed words the closing line: which families a shell can now be pointed at, since the
// family name is rarely the name of the font that was asked for.
func installed(name string, families []string) string {
	if len(families) == 0 {
		return fmt.Sprintf("No matching font families were installed. Try --zip-folder when installing %s or a custom font zip.", name)
	}

	sb := text.NewBuilder()

	sb.WriteString(fmt.Sprintf("Installed %s 🚀\n\nThe following font families are now available for configuration:\n", name))

	for _, family := range families {
		sb.WriteString(fmt.Sprintf("\n  • %s", family))
	}

	return sb.String()
}
