package upgrade

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
)

type release struct {
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	NodeID          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TarballURL      string    `json:"tarball_url"`
	ZipballURL      string    `json:"zipball_url"`
	DiscussionURL   string    `json:"discussion_url"`
	HTMLURL         string    `json:"html_url"`
	URL             string    `json:"url"`
	UploadURL       string    `json:"upload_url"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Body            string    `json:"body"`
	AssetsURL       string    `json:"assets_url"`
	ID              int       `json:"id"`
	Prerelease      bool      `json:"prerelease"`
	Draft           bool      `json:"draft"`
}

const (
	RELEASEURL = "https://api.github.com/repos/jandedobbeleer/oh-my-posh/releases/latest"
	CACHEKEY   = "upgrade_check"

	upgradeNotice = `
A new release of Oh My Posh is available: %s → %s
To upgrade, run: 'oh-my-posh upgrade'

To enable automated upgrades, set 'auto_upgrade' to 'true' in your configuration.
`
)

func Latest(env runtime.Environment) (string, error) {
	body, err := env.HTTPRequest(RELEASEURL, nil, 5000)
	if err != nil {
		return "", err
	}
	var release release
	// this can't fail
	_ = json.Unmarshal(body, &release)

	if len(release.TagName) == 0 {
		return "", fmt.Errorf("failed to get latest release")
	}

	return release.TagName, nil
}

// Returns the upgrade notice if a new version is available
// that should be displayed to the user.
//
// The upgrade check is only performed every other week.
func Notice(env runtime.Environment, force bool) (string, bool) {
	// do not check when last validation was < 1 week ago
	if _, OK := env.Cache().Get(CACHEKEY); OK && !force {
		return "", false
	}

	if !http.IsConnected() {
		return "", false
	}

	// never validate when we install using the Windows Store
	if env.Getenv("POSH_INSTALLER") == "ws" {
		return "", false
	}

	latest, err := Latest(env)
	if err != nil {
		return "", false
	}

	env.Cache().Set(CACHEKEY, latest, "1week")

	version := fmt.Sprintf("v%s", build.Version)
	if latest == version {
		return "", false
	}

	return fmt.Sprintf(upgradeNotice, version, latest), true
}
