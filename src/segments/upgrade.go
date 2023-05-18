package segments

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/upgrade"
)

type UpgradeCache struct {
	Latest  string `json:"latest"`
	Current string `json:"current"`
}

type Upgrade struct {
	props properties.Properties
	env   platform.Environment

	Version string
}

const UPGRADECACHEKEY = "upgrade_segment"

func (u *Upgrade) Template() string {
	return " \uf019 "
}

func (u *Upgrade) Init(props properties.Properties, env platform.Environment) {
	u.props = props
	u.env = env
}

func (u *Upgrade) Enabled() bool {
	version := fmt.Sprintf("v%s", u.env.Flags().Version)

	if shouldDisplay, err := u.ShouldDisplay(version); err == nil {
		return shouldDisplay
	}

	latest, err := upgrade.Latest(u.env)
	if err != nil {
		return false
	}

	if latest == version {
		return false
	}

	cacheData := &UpgradeCache{
		Latest:  latest,
		Current: version,
	}
	cacheJSON, err := json.Marshal(cacheData)
	if err != nil {
		return false
	}

	oneWeek := 10080
	cacheTimeout := u.props.GetInt(properties.CacheTimeout, oneWeek)
	u.env.Cache().Set(UPGRADECACHEKEY, string(cacheJSON), cacheTimeout)

	u.Version = latest
	return true
}

func (u *Upgrade) ShouldDisplay(version string) (bool, error) {
	data, OK := u.env.Cache().Get(UPGRADECACHEKEY)
	if !OK {
		return false, errors.New("no cache")
	}

	var cacheJSON UpgradeCache
	err := json.Unmarshal([]byte(data), &cacheJSON)
	if err != nil {
		return false, errors.New("invalid cache data")
	}

	if version == cacheJSON.Current {
		return cacheJSON.Current != cacheJSON.Latest, nil
	}

	return false, errors.New("version changed, run the check again")
}
