package upgrade

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
)

type release struct {
	URL             string    `json:"url"`
	HTMLURL         string    `json:"html_url"`
	AssetsURL       string    `json:"assets_url"`
	UploadURL       string    `json:"upload_url"`
	TarballURL      string    `json:"tarball_url"`
	ZipballURL      string    `json:"zipball_url"`
	DiscussionURL   string    `json:"discussion_url"`
	ID              int       `json:"id"`
	NodeID          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Body            string    `json:"body"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
}

const (
	releaseURL    = "https://api.github.com/repos/jandedobbeleer/oh-my-posh/releases/latest"
	upgradeNotice = `
A new release of Oh My Posh is available: %s → %s
%s
`
	windows = `To upgrade, use the guide for your favorite package manager in the documentation:
https://ohmyposh.dev/docs/installation/windows#update`
	unix   = "To upgrade, use your favorite package manager or, if you used Homebrew to install, run: 'brew update && brew upgrade oh-my-posh'"
	darwin = "To upgrade, run: 'brew update && brew upgrade oh-my-posh'"

	CACHEKEY = "upgrade_check"
)

func getLatestVersion(env platform.Environment) (string, error) {
	body, err := env.HTTPRequest(releaseURL, nil, 1000)
	if err != nil {
		return "", err
	}
	var release release
	// this can't fail
	_ = json.Unmarshal(body, &release)
	return release.TagName, nil
}

// Returns the upgrade notice if a new version is available
// that should be displayed to the user.
//
// The upgrade check is only performed every other week.
func Notice(env platform.Environment) (string, bool) {
	// never validate when we install using the Windows Store
	if env.Getenv("POSH_INSTALLER") == "ws" {
		return "", false
	}
	// do not check when last validation was < 1 week ago
	if _, OK := env.Cache().Get(CACHEKEY); OK {
		return "", false
	}

	latest, err := getLatestVersion(env)
	if err != nil {
		return "", false
	}

	oneWeek := 10080
	env.Cache().Set(CACHEKEY, latest, oneWeek)

	version := fmt.Sprintf("v%s", env.Flags().Version)
	if latest == version {
		return "", false
	}

	var notice string
	switch env.GOOS() {
	case platform.WINDOWS:
		notice = windows
	case platform.DARWIN:
		notice = darwin
	case platform.LINUX:
		notice = unix
	}

	return fmt.Sprintf(upgradeNotice, version, latest, notice), true
}
