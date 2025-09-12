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
			return Load(source, false)
		}
	}

	base64String, found := cache.Get[string](cache.Session, configKey)
	if !found {
		log.Debug("no cached config found")
		return Load(configFile, false)
	}

	var cfg Config
	if err := cfg.Restore(base64String); err != nil {
		log.Debug("failed to restore config from cache")
		return Load(configFile, false)
	}

	return &cfg
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
