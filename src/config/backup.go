package config

import (
	"bytes"
	"io"
	"os"
	"strings"

	json "github.com/goccy/go-json"
	yaml "github.com/goccy/go-yaml"
	toml "github.com/pelletier/go-toml/v2"
)

func (cfg *Config) Backup() {
	dst := cfg.origin + ".bak"
	source, err := os.Open(cfg.origin)
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

func (cfg *Config) Export(format string) string {
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
		return escapeGlyphs(data, cfg.MigrateGlyphs)
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

func (cfg *Config) BackupAndMigrate() {
	cfg.Backup()
	cfg.Migrate()
	cfg.Write(cfg.Format)
}

func (cfg *Config) Write(format string) {
	content := cfg.Export(format)
	if len(content) == 0 {
		// we are unable to perform the export
		os.Exit(65)
		return
	}

	destination := cfg.Output
	if len(destination) == 0 {
		destination = cfg.origin
	}

	f, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
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
