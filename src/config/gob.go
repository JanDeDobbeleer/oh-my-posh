package config

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

func init() {
	// Register types that can appear in any values for gob encoding/decoding
	// This is necessary for properties.Map which contains map[Property]any
	gob.Register([]any{})
	gob.Register(map[string]any{})
	gob.Register(map[any]any{})
	gob.Register([]string{})
	gob.Register(map[string]string{})
	gob.Register([]int{})
	gob.Register([]float64{})
	gob.Register([]bool{})
	gob.Register(int64(0))
	gob.Register(uint64(0))
	gob.Register(float32(0))
	gob.Register(properties.Map{})
	gob.Register(properties.Property(""))
	gob.Register(map[properties.Property]any{})
}

const (
	key = "CONFIG_GOB"
)

func (cfg *Config) Store(session cache.Cache) {
	defer log.Trace(time.Now())

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(cfg)
	if err != nil {
		log.Error(err)
		return
	}

	// Encode the binary gob data as base64 string
	gobBase64 := base64.StdEncoding.EncodeToString(buffer.Bytes())
	session.Set(key, gobBase64, cache.INFINITE)
}

func Get(session cache.Cache, configFile string, edit bool) *Config {
	defer log.Trace(time.Now())

	if edit {
		log.Debug("edit mode enabled")
		cfg, _ := Load(configFile, false)
		return cfg
	}

	gobBase64, found := session.Get(key)
	if !found {
		log.Debug("no cached config found")
		cfg, _ := Load(configFile, false)
		return cfg
	}

	// Decode base64 back to binary
	gobData, err := base64.StdEncoding.DecodeString(gobBase64)
	if err != nil {
		log.Error(err)
		cfg, _ := Load(configFile, false)
		return cfg
	}

	var cfg Config
	decoder := gob.NewDecoder(bytes.NewReader(gobData))
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Error(err)
		cfg, _ := Load(configFile, false)
		return cfg
	}

	return &cfg
}
