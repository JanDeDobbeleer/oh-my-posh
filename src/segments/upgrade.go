package segments

import (
	"errors"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/upgrade"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

type UpgradeCache struct {
	Latest  string `json:"latest"`
	Current string `json:"current"`
}

type Upgrade struct {
	Base

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
	upgradeCache, err := u.upgradeCache()
	if err != nil || upgradeCache.Current != u.Current {
		upgradeCache, err = u.checkUpdate(u.Current)
	}

	if err != nil || u.Current == upgradeCache.Latest {
		return false
	}

	u.UpgradeCache = *upgradeCache
	u.Version = u.Latest
	return true
}

func (u *Upgrade) upgradeCache() (*UpgradeCache, error) {
	data, ok := cache.Get[*UpgradeCache](cache.Device, UPGRADECACHEKEY)
	if !ok {
		return nil, errors.New("no cache data")
	}

	return data, nil
}

func (u *Upgrade) checkUpdate(current string) (*UpgradeCache, error) {
	duration := u.options.String(options.CacheDuration, string(cache.ONEWEEK))
	source := u.options.String(Source, string(upgrade.CDN))

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
