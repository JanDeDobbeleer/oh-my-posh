package config

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	toml "github.com/pelletier/go-toml/v2"
	yaml "go.yaml.in/yaml/v3"
)

func (cfg *Config) Backup() {
	dst := cfg.Source + ".bak"
	source, err := os.Open(cfg.Source)
	if err != nil {
		return
	}
	defer source.Close()
	destination, err := os.Create(dst)
	if err != nil {
		return
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	if err != nil {
		return
	}
}

// stripRuntimeState removes runtime-only state (e.g. the template context
// injected by Segment.MapSegmentWithWriter) from every segment's options map.
// It must be called before any encoder runs over the config so that internal
// state never reaches a serialized form.
func (cfg *Config) stripRuntimeState() {
	for _, block := range cfg.Blocks {
		for _, segment := range block.Segments {
			delete(segment.Options, options.TemplateContextKey)
		}
	}
}

func (cfg *Config) Export(format string) string {
	cfg.stripRuntimeState()

	if len(format) != 0 {
		cfg.Format = format
	}

	var result bytes.Buffer

	switch cfg.Format {
	case YAML:
		prefix := "# yaml-language-server: $schema=https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json\n\n"
		yamlEncoder := yaml.NewEncoder(&result)

		err := yamlEncoder.Encode(cfg)
		if err != nil {
			return ""
		}

		return prefix + result.String()
	case JSON:
		jsonEncoder := json.NewEncoder(&result)
		jsonEncoder.SetEscapeHTML(false)
		jsonEncoder.SetIndent("", "  ")
		_ = jsonEncoder.Encode(cfg)
		prefix := "{\n  \"$schema\": \"https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json\","
		data := strings.Replace(result.String(), "{", prefix, 1)
		return EscapeGlyphs(data, cfg.MigrateGlyphs)
	case TOML:
		tomlEncoder := toml.NewEncoder(&result)
		tomlEncoder.SetIndentTables(true)

		err := tomlEncoder.Encode(cfg)
		if err != nil {
			return ""
		}

		return result.String()
	}

	// unsupported format
	return ""
}

func (cfg *Config) Write(format string) {
	content := cfg.Export(format)
	if content == "" {
		// we are unable to perform the export
		return
	}

	f, err := os.OpenFile(cfg.Source, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return
	}

	defer func() {
		_ = f.Close()
	}()

	_, err = f.WriteString(content)
	if err != nil {
		return
	}
}
