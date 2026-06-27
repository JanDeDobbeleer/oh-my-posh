package config

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

const (
	configKey = "CONFIG"
	SourceKey = "CONFIG_SOURCE"
)

func (cfg *Config) Store() {
	defer log.Trace(time.Now())

	cache.Set(cache.Session, SourceKey, cfg.Source, cache.INFINITE)
	cache.Set(cache.Session, configKey, cfg.Base64(), cache.INFINITE)
}

func Get(configFile string, reload bool) *Config {
	defer log.Trace(time.Now())

	if reload {
		log.Debug("reload mode enabled")
		if source, OK := cache.Get[string](cache.Session, SourceKey); OK {
			cfg := Load(source)
			cfg.Store()
			return cfg
		}
	}

	if base64String, found := cache.Get[string](cache.Session, configKey); found {
		var cfg Config
		if err := cfg.Restore(base64String); err == nil {
			return &cfg
		}

		log.Debug("failed to restore config from cache")
	} else {
		log.Debug("no cached config found")
	}

	// A prompt render relies on the session cache for its configuration: it's
	// invoked without an explicit --config, so configFile is empty here. When the
	// session cache is missing or corrupt - the file got purged (e.g. under disk
	// pressure) or clobbered by a concurrent writer sharing the session - falling
	// back to Load("") silently renders the default theme for the rest of the
	// session. Recover the previously initialized source first so the configured
	// theme keeps rendering and the session cache self-heals.
	if configFile == "" {
		if source := recoverSource(); source != "" {
			if cfg, err := Parse(source); err == nil {
				log.Debug("recovered configuration from source: ", source)
				cfg.Store()
				return cfg
			}

			log.Debug("failed to recover configuration from source: ", source)
		}
	}

	cfg := Load(configFile)
	cfg.Store()
	return cfg
}

// recoverSource locates the most recently initialized configuration source when
// the session config cache is unavailable. It first checks the session source,
// which survives when only the config blob is corrupt, then falls back to the
// device-level DSC, which persists across sessions with an infinite TTL and is
// populated on every init (where --config is required).
func recoverSource() string {
	defer log.Trace(time.Now())

	if source, OK := cache.Get[string](cache.Session, SourceKey); OK && source != "" {
		return source
	}

	resource := DSC()
	resource.Load()

	if len(resource.States) == 0 {
		return ""
	}

	// the last added state reflects the most recent init.
	if last := resource.States[len(resource.States)-1]; last != nil {
		return last.Source
	}

	return ""
}

func (cfg *Config) Base64() string {
	defer log.Trace(time.Now())

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(cfg)
	if err != nil {
		log.Error(err)
		return ""
	}

	return base64.StdEncoding.EncodeToString(buffer.Bytes())
}

func (cfg *Config) Restore(base64String string) error {
	defer log.Trace(time.Now())

	data, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		log.Error(err)
		return err
	}

	var buffer bytes.Buffer
	buffer.Write(data)
	decoder := gob.NewDecoder(&buffer)
	err = decoder.Decode(cfg)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
