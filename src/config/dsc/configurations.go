package dsc

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/dsc"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

//go:embed schema.json
var Schema string

type State struct {
	Configurations []*Configuration `json:"configurations,omitempty"`
}

func (c *State) Add(configPath string) {
	log.Debug("Adding configuration %s", configPath)

	if configPath == "" || strings.HasPrefix(configPath, "http") {
		return
	}

	// replace $HOME with tilde as we can't guarantee the home path
	configPath = filepath.Clean(configPath)
	configPath = strings.ReplaceAll(configPath, path.Home(), "~")

	if c.exists(configPath) {
		log.Debug("Configuration already exists")
		return
	}

	cfg := &Configuration{
		Path: configPath,
	}

	c.Configurations = append(c.Configurations, cfg)
}

func (c *State) Test(_ dsc.State) error {
	return errors.New("test not implemented for configurations")
}

func (c *State) Resolve() {
	for _, cfg := range c.Configurations {
		c.resolveConfig(cfg)
	}
}

func (c *State) Apply(_ cache.Cache) error {
	log.Debug("Applying configurations")

	var err error

	for _, cfg := range c.Configurations {
		if applyErr := cfg.Apply(); applyErr != nil {
			log.Error(applyErr)
			err = errors.Join(err, applyErr)
		}
	}

	log.Debug("Configurations applied")

	return err
}

func (c *State) String() string {
	schemaJSON := `"$schema": "https://ohmyposh.dev/dsc.config.schema.json"`
	if len(c.Configurations) == 0 {
		return fmt.Sprintf("{%s}", schemaJSON)
	}

	var result bytes.Buffer
	jsonEncoder := json.NewEncoder(&result)
	jsonEncoder.SetEscapeHTML(false)
	jsonEncoder.SetIndent("", "  ")
	_ = jsonEncoder.Encode(c)
	prefix := fmt.Sprintf("{\n  %s,", schemaJSON)
	return strings.Replace(result.String(), "{", prefix, 1)
}

func (c *State) CacheKey() string {
	return "DSC_CONFIGURATIONS"
}

func (c *State) New() dsc.State {
	return &State{}
}

func (c *State) Schema() string {
	return Schema
}

func (c *State) exists(configPath string) bool {
	for _, cfg := range c.Configurations {
		if cfg.Path == configPath {
			return true
		}
	}

	return false
}

func (c *State) resolveConfig(cfg *Configuration) {
	log.Debug("Resolving configuration %s", cfg.Path)
	if cfg.resolved {
		log.Debug("Configuration already resolved")
		return
	}

	cfg.resolved = true

	// we use pwsh as that will never omit any feature
	data, _ := config.Load(cfg.Path, shell.PWSH, false)
	if data == nil {
		log.Debug("No configuration data found")
		return
	}

	cfg.Data = data
	cfg.Format = data.Format

	// Skip if no extends, http URL, or already processed
	if data.Extends == "" || strings.HasPrefix(data.Extends, "http") {
		log.Debug("No extends found or already processed")
		return
	}

	var extCfg *Configuration

	// Create new configuration if it doesn't exist
	if !c.exists(data.Extends) {
		log.Debug("Adding extended configuration %s", data.Extends)
		extCfg = &Configuration{
			Path: data.Extends,
		}

		c.Configurations = append(c.Configurations, extCfg)
	}

	// Recursively resolve the extended configuration
	c.resolveConfig(extCfg)
}
