package config

import (
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
)

func (cfg *Config) Migrate() {
	for _, block := range cfg.Blocks {
		for _, segment := range block.Segments {
			segment.migrate(cfg.Version)
		}
	}

	cfg.updated = true
	cfg.Version = Version
}

func (segment *Segment) migrate(version int) {
	// configs older than 2 are no longer supported
	if version != 2 {
		return
	}

	// Cache settings
	delete(segment.Properties, "cache_version")
	cacheTimeout := segment.Properties.GetInt("cache_timeout", 0)
	if cacheTimeout != 0 {
		segment.Cache = &cache.Config{
			Duration: cache.ToDuration(cacheTimeout * 60),
			Strategy: cache.Folder,
		}
	}
}
