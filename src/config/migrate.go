package config

import (
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

const (
	includeFolders = properties.Property("include_folders")
	excludeFolders = properties.Property("exclude_folders")
	cacheTimeout   = properties.Property("cache_timeout")
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

	if segment.Properties == nil {
		segment.Properties = properties.Map{}
	}

	// Cache settings, the default is now 24h so we have to respect this being disabled previously
	if !segment.Properties.GetBool("cache_version", false) {
		segment.Properties[properties.CacheDuration] = cache.NONE
	}
	delete(segment.Properties, "cache_version")

	segment.IncludeFolders = segment.migrateFolders(includeFolders)
	segment.ExcludeFolders = segment.migrateFolders(excludeFolders)

	switch segment.Type { //nolint:exhaustive
	case UPGRADE:
		segment.timeoutToDuration()
	default:
		segment.timeoutToCache()
	}
}

func (segment *Segment) hasProperty(property properties.Property) bool {
	for key := range segment.Properties {
		if key == property {
			return true
		}
	}
	return false
}

func (segment *Segment) timeoutToCache() {
	if !segment.hasProperty(cacheTimeout) {
		return
	}

	timeout := segment.Properties.GetInt(cacheTimeout, 0)
	delete(segment.Properties, cacheTimeout)

	if timeout == 0 {
		return
	}

	segment.Cache = &cache.Config{
		Duration: cache.ToDuration(timeout * 60),
		Strategy: cache.Folder,
	}
}

func (segment *Segment) timeoutToDuration() {
	timeout := segment.Properties.GetInt(cacheTimeout, 0)
	delete(segment.Properties, cacheTimeout)

	if timeout == 0 {
		return
	}

	segment.Properties[properties.CacheDuration] = cache.ToDuration(timeout * 60)
}

func (segment *Segment) migrateFolders(property properties.Property) []string {
	if !segment.hasProperty(property) {
		return []string{}
	}

	array := segment.Properties.GetStringArray(property, []string{})
	delete(segment.Properties, property)

	return array
}
