package dsc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

type Configurations []*Configuration

type Configuration struct {
	Data   *config.Config `json:"data,omitempty"`
	Path   string         `json:"path,omitempty"`
	Format string         `json:"format,omitempty"`

	resolved bool `json:"-"`
}

func (c *Configurations) exists(configPath string) bool {
	for _, cfg := range *c {
		if cfg.Path == configPath {
			return true
		}
	}

	return false
}

func (c *Configurations) Add(configPath string) {
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

	*c = append(*c, cfg)
}

func (c *Configurations) Resolve() {
	for _, cfg := range *c {
		c.resolveConfig(cfg)
	}
}

func (c *Configurations) resolveConfig(cfg *Configuration) {
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

		*c = append(*c, extCfg)
	}

	// Recursively resolve the extended configuration
	c.resolveConfig(extCfg)
}

func (c *Configurations) Apply() error {
	log.Debug("Applying configurations")

	var err error

	for _, cfg := range *c {
		if applyErr := cfg.Apply(); applyErr != nil {
			log.Error(applyErr)
			err = errors.Join(err, applyErr)
		}
	}

	log.Debug("Configurations applied")

	return err
}

func (c *Configuration) Apply() error {
	if c.Data == nil {
		return nil
	}

	log.Debug("Applying configuration %s", c.Path)

	// Expand tilde to home directory for file operations
	filePath := strings.ReplaceAll(c.Path, "~", path.Home())

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	data := c.Data.Export(c.Format)

	// Write file
	if err := os.WriteFile(filePath, []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to write configuration file %s: %w", filePath, err)
	}

	log.Debug("Configuration written to %s", filePath)
	return nil
}
