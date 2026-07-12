package upgrade

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
)

const (
	CACHEKEY = "upgrade_check"

	upgradeNotice = `
A new release of Oh My Posh is available: v%s → v%s
To upgrade, run: 'oh-my-posh upgrade%s'

To enable automated upgrades, run: 'oh-my-posh enable upgrade'.
`
)

// Returns the upgrade notice if a new version is available
// that should be displayed to the user.
//
// The upgrade check is only performed every other week.
func (cfg *Config) Notice() (string, bool) {
	if !http.IsConnected() {
		return "", false
	}

	latest, err := cfg.FetchLatest()
	if err != nil {
		return "", false
	}

	if latest == build.Version {
		return "", false
	}

	var forceUpdate string
	if IsMajorUpgrade(build.Version, latest) {
		forceUpdate = " --force"
	}

	return fmt.Sprintf(upgradeNotice, build.Version, latest, forceUpdate), true
}
