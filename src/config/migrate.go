package config

import (
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

const (
	includeFolders = options.Option("include_folders")
	excludeFolders = options.Option("exclude_folders")
	cacheTimeout   = options.Option("cache_timeout")
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
	if version < 2 {
		return
	}

	// Ensure Options is initialized
	if segment.Options == nil {
		segment.Options = options.Map{}
	}

	if version < 4 && version != 2 {
		return
	}

	// Version 2 migrations work on Options
	// Cache settings, the default is now 24h so we have to respect this being disabled previously
	if !segment.Options.Bool("cache_version", false) {
		segment.Options[options.CacheDuration] = cache.NONE
	}
	delete(segment.Options, "cache_version")

	segment.IncludeFolders = segment.migrateFolders(includeFolders)
	segment.ExcludeFolders = segment.migrateFolders(excludeFolders)

	switch segment.Type { //nolint:exhaustive
	case UPGRADE:
		segment.timeoutToDuration()
	default:
		segment.timeoutToCache()
	}
}

func (segment *Segment) hasProperty(property options.Option) bool {
	for key := range segment.Options {
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

	timeout := segment.Options.Int(cacheTimeout, 0)
	delete(segment.Options, cacheTimeout)

	if timeout == 0 {
		return
	}

	segment.Cache = &Cache{
		Duration: cache.ToDuration(timeout * 60),
		Strategy: Folder,
	}
}

func (segment *Segment) timeoutToDuration() {
	timeout := segment.Options.Int(cacheTimeout, 0)
	delete(segment.Options, cacheTimeout)

	if timeout == 0 {
		return
	}

	segment.Options[options.CacheDuration] = cache.ToDuration(timeout * 60)
}

func (segment *Segment) migrateFolders(property options.Option) []string {
	if !segment.hasProperty(property) {
		return []string{}
	}

	array := segment.Options.StringArray(property, []string{})
	delete(segment.Options, property)

	return array
}
