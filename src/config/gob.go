package config

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"os"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

const (
	configKey = "CONFIG"
	SourceKey = "CONFIG_SOURCE"

	// envKey is resolved into the --config flag by the root command when no
	// explicit flag is given. The shell's init script pins it to the session's
	// resolved configuration source so the configuration can be recovered when
	// the session cache is lost or corrupted. A healthy session cache always
	// wins, so it can not be used to change the configuration mid-session.
	envKey = "POSH_CONFIG"
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
		// the entry is corrupted beyond repair, drop it so subsequent renders
		// skip the decode attempt and go straight to recovery
		cache.Delete(cache.Session, configKey)
	} else {
		log.Debug("no cached config found")
	}

	// A prompt render is invoked without an explicit --config and relies on the
	// session cache for its configuration. When that cache is missing or corrupt
	// - the file got purged (e.g. under disk pressure) or clobbered by a
	// concurrent writer sharing the session - the shell session still knows its
	// source: init exported it as POSH_CONFIG, and the root command resolved it
	// into configFile. Store the parsed result again so the session cache
	// self-heals. When the source is unavailable, return the default without
	// caching it so the next render retries the recovery instead of pinning
	// the default theme for the rest of the session.
	if configFile != "" && configFile == os.Getenv(envKey) {
		cfg, err := Parse(configFile)
		if err == nil {
			log.Debug("recovered configuration from environment: ", configFile)
			cfg.Store()
			return cfg
		}

		log.Debug("failed to recover configuration from environment: ", configFile)
		return Default(err)
	}

	cfg := Load(configFile)
	cfg.Store()
	return cfg
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
