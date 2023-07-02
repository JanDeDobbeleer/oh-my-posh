package segments

import (
	"encoding/json"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
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
	current := build.Version
	latest := u.cachedLatest(current)
	if len(latest) == 0 {
		latest = u.checkUpdate(current)
	}

	if len(latest) == 0 || current == latest {
		return false
	}

	u.Version = latest
	return true
}

func (u *Upgrade) cachedLatest(current string) string {
	data, ok := u.env.Cache().Get(UPGRADECACHEKEY)
	if !ok {
		return "" // no cache
	}

	var cacheJSON UpgradeCache
	err := json.Unmarshal([]byte(data), &cacheJSON)
	if err != nil {
		return "" // invalid cache data
	}

	if current != cacheJSON.Current {
		return "" // version changed, run the check again
	}

	return cacheJSON.Latest
}

func (u *Upgrade) checkUpdate(current string) string {
	tag, err := upgrade.Latest(u.env)
	if err != nil {
		return ""
	}

	latest := tag[1:]
	cacheData := &UpgradeCache{
		Latest:  latest,
		Current: current,
	}
	cacheJSON, err := json.Marshal(cacheData)
	if err != nil {
		return ""
	}

	oneWeek := 10080
	cacheTimeout := u.props.GetInt(properties.CacheTimeout, oneWeek)
	// update cache
	u.env.Cache().Set(UPGRADECACHEKEY, string(cacheJSON), cacheTimeout)

	return latest
}
