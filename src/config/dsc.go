package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/dsc"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

type Resource struct {
	dsc.Resource[*Configuration]
}

func DSC() *Resource {
	return &Resource{
		Resource: dsc.Resource[*Configuration]{
			JSONSchemaURL: "https://ohmyposh.dev/dsc.configuration.schema.json",
		},
	}
}

type Configuration struct {
	Format string `json:"format,omitempty" jsonschema:"title=Format,description=The format of the configuration file,enum=json,enum=jsonc,enum=yaml,enum=yml,enum=toml,enum=tml"`
	Source string `json:"source,omitempty" jsonschema:"title=Source,description=The source of the configuration file"`
	Config
	resolved bool `json:"-"`
}

func (s *Resource) Add(configPath string) {
	if configPath == "" || strings.HasPrefix(configPath, "http") {
		log.Debug("Invalid configuration path:", configPath)
		return
	}

	// replace $HOME with tilde as we can't guarantee the home path
	configPath = filepath.Clean(configPath)
	configPath = strings.ReplaceAll(configPath, path.Home(), "~")

	s.Resource.Add(&Configuration{
		Source: configPath,
	})
}

func (s *Resource) ToJSON() string {
	output := s.Resource.ToJSON()
	return EscapeGlyphs(output, false)
}

func (c *Configuration) Apply() error {
	if c == nil {
		return nil
	}

	formats := map[string][]string{
		JSON: {".json", ".jsonc"},
		YAML: {".yaml", ".yml"},
		TOML: {".toml", ".tml"},
	}

	if !slices.Contains(formats[c.Format], filepath.Ext(c.Source)) {
		return fmt.Errorf("source file %s does not match format %s", c.Source, c.Format)
	}

	log.Debug("Applying configuration %s", c.Source)

	// Expand tilde to home directory for file operations
	filePath := strings.ReplaceAll(c.Source, "~", path.Home())

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	data := c.Export(c.Format)

	// Write file
	if err := os.WriteFile(filePath, []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to write configuration file %s: %w", filePath, err)
	}

	log.Debug("Configuration written to %s", filePath)
	return nil
}

func (c *Configuration) Equal(config *Configuration) bool {
	if config == nil {
		return false
	}

	return c.Source == config.Source
}

func (c *Configuration) Resolve() (*Configuration, bool) {
	log.Debug("Resolving configuration %s", c.Source)

	if c.resolved {
		log.Debug("Configuration already resolved")
		return c, true
	}

	c.resolved = true

	// we use pwsh as that will never omit any feature
	data, _ := Load(c.Source, shell.PWSH, false)
	if data == nil {
		log.Debug("No configuration data found")
		return nil, false
	}

	c.Config = *data
	c.Format = data.Format

	// Skip if no extends, http URL
	if data.Extends == "" || strings.HasPrefix(data.Extends, "http") {
		log.Debug("No extends found or remote configuration")
		return c, false
	}

	// Resolve the extends configuration
	parent := &Configuration{
		Source: data.Extends,
	}

	return parent, true
}
