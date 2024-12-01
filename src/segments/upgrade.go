package segments

import (
	"encoding/json"
	"errors"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/upgrade"
)

type UpgradeCache struct {
	Latest  string `json:"latest"`
	Current string `json:"current"`
}

type Upgrade struct {
	base

	// deprecated
	Version string

	UpgradeCache
}

const (
	UPGRADECACHEKEY = "upgrade_segment"
)

func (u *Upgrade) Template() string {
	return " \uf019 "
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
	duration := u.props.GetString(properties.CacheDuration, string(cache.ONEWEEK))
	source := u.props.GetString(Source, string(upgrade.CDN))

	cfg := &upgrade.Config{
		Source:   upgrade.Source(source),
		Interval: cache.Duration(duration),
	}

	latest, err := cfg.Latest()
	if err != nil {
		return nil, err
	}

	cacheData := &UpgradeCache{
		Latest:  latest,
		Current: current,
	}

	cacheJSON, err := json.Marshal(cacheData)
	if err != nil {
		return nil, err
	}

	u.env.Cache().Set(UPGRADECACHEKEY, string(cacheJSON), cache.Duration(duration))

	return cacheData, nil
}
