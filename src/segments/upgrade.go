package segments

import (
	"encoding/json"
	"errors"

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

	// deprecated
	Version string

	UpgradeCache
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
	u.Current = build.Version
	latest, err := u.cachedLatest(u.Current)
	if err != nil {
		latest, err = u.checkUpdate(u.Current)
	}

	if err != nil || u.Current == latest.Latest {
		return false
	}

	u.UpgradeCache = *latest
	u.Version = u.Latest
	return true
}

func (u *Upgrade) cachedLatest(current string) (*UpgradeCache, error) {
	data, ok := u.env.Cache().Get(UPGRADECACHEKEY)
	if !ok {
		return nil, errors.New("no cache data")
	}

	var cacheJSON UpgradeCache
	err := json.Unmarshal([]byte(data), &cacheJSON)
	if err != nil {
		return nil, err // invalid cache data
	}

	if current != cacheJSON.Current {
		return nil, errors.New("version changed, run the check again")
	}

	return &cacheJSON, nil
}

func (u *Upgrade) checkUpdate(current string) (*UpgradeCache, error) {
	tag, err := upgrade.Latest(u.env)
	if err != nil {
		return nil, err
	}

	latest := tag[1:]
	cacheData := &UpgradeCache{
		Latest:  latest,
		Current: current,
	}
	cacheJSON, err := json.Marshal(cacheData)
	if err != nil {
		return nil, err
	}

	oneWeek := 10080
	cacheTimeout := u.props.GetInt(properties.CacheTimeout, oneWeek)
	// update cache
	u.env.Cache().Set(UPGRADECACHEKEY, string(cacheJSON), cacheTimeout)

	return cacheData, nil
}
