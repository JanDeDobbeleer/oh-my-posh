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
	key = "CONFIG_GOB"
)

func (cfg *Config) Store() {
	defer log.Trace(time.Now())

	// TODO: Save this as a Config and no longer parse, we can deep equal to see if we need to reload
	cache.Set(cache.Session, key, cfg.Base64(), cache.INFINITE)
}

func (cfg *Config) Base64() string {
	defer log.Trace(time.Now())

	if cfg.base64 != "" {
		return cfg.base64
	}

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(cfg)
	if err != nil {
		log.Error(err)
		return ""
	}

	gobBase64 := base64.StdEncoding.EncodeToString(buffer.Bytes())
	cfg.base64 = gobBase64

	return gobBase64
}

func Get(configFile string, reload bool) *Config {
	defer log.Trace(time.Now())

	gobBase64, found := cache.Get[string](cache.Session, key)
	if !found {
		log.Debug("no cached config found")
		cfg := Load(configFile, false)
		return cfg
	}

	// Decode base64 back to binary
	gobData, err := base64.StdEncoding.DecodeString(gobBase64)
	if err != nil {
		log.Error(err)
		cfg := Load(configFile, false)
		return cfg
	}

	cfg := &Config{}
	decoder := gob.NewDecoder(bytes.NewReader(gobData))
	err = decoder.Decode(cfg)
	if err != nil {
		log.Error(err)
		cfg = Load(configFile, false)
		return cfg
	}

	if reload {
		log.Debug("edit mode enabled")
		cfg = Load(cfg.Source, false)
		return cfg
	}

	return cfg
}
