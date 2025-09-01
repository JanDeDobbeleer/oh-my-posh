package segments

import (
	"errors"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/upgrade"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
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
	latest, err := u.cachedLatest()
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

func (u *Upgrade) cachedLatest() (*UpgradeCache, error) {
	data, ok := cache.Get[*UpgradeCache](cache.Device, UPGRADECACHEKEY)
	if !ok {
		return nil, errors.New("no cache data")
	}

	return data, nil
}

func (u *Upgrade) checkUpdate(current string) (*UpgradeCache, error) {
	duration := u.props.GetString(properties.CacheDuration, string(cache.ONEWEEK))
	source := u.props.GetString(Source, string(upgrade.CDN))

	cfg := &upgrade.Config{
		Source:   upgrade.Source(source),
		Interval: cache.Duration(duration),
	}

	latest, err := cfg.FetchLatest()
	if err != nil {
		return nil, err
	}

	cacheData := &UpgradeCache{
		Latest:  latest,
		Current: current,
	}

	cache.Set(cache.Device, UPGRADECACHEKEY, cacheData, cache.Duration(duration))

	return cacheData, nil
}
