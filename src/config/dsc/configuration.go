package dsc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
)

type Configuration struct {
	Data   *config.Config `json:"data,omitempty"`
	Path   string         `json:"path,omitempty"`
	Format string         `json:"format,omitempty"`

	resolved bool `json:"-"`
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
