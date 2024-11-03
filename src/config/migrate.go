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

	// Cache settings
	delete(segment.Properties, "cache_version")
	segment.Cache = segment.migrateCache()

	segment.IncludeFolders = segment.migrateFolders(includeFolders)
	segment.ExcludeFolders = segment.migrateFolders(excludeFolders)
}

func (segment *Segment) hasProperty(property properties.Property) bool {
	for key := range segment.Properties {
		if key == property {
			return true
		}
	}
	return false
}

func (segment *Segment) migrateCache() *cache.Config {
	if !segment.hasProperty(cacheTimeout) {
		return nil
	}

	timeout := segment.Properties.GetInt(cacheTimeout, 0)
	delete(segment.Properties, cacheTimeout)

	if timeout == 0 {
		return nil
	}

	return &cache.Config{
		Duration: cache.ToDuration(timeout * 60),
		Strategy: cache.Folder,
	}
}

func (segment *Segment) migrateFolders(property properties.Property) []string {
	if !segment.hasProperty(property) {
		return []string{}
	}

	array := segment.Properties.GetStringArray(property, []string{})
	delete(segment.Properties, property)

	return array
}
