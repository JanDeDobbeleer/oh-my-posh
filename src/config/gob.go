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
	configKey       = "CONFIG"
	configSourceKey = "CONFIG_SOURCE"
)

func (cfg *Config) Store() {
	defer log.Trace(time.Now())

	cache.Set(cache.Session, configSourceKey, cfg.Source, cache.INFINITE)
	cache.Set(cache.Session, configKey, cfg, cache.INFINITE)
}

func Get(configFile string, reload bool) *Config {
	defer log.Trace(time.Now())

	if reload {
		log.Debug("reload mode enabled")
		if source, OK := cache.Get[string](cache.Session, configSourceKey); OK {
			return Load(source, false)
		}
	}

	cfg, found := cache.Get[*Config](cache.Session, configKey)
	if !found {
		log.Debug("no cached config found")
		return Load(configFile, false)
	}

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
