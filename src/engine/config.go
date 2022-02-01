package engine

import (
	"bytes"
	json2 "encoding/json"
	"errors"
	"fmt"
	"oh-my-posh/color"
	"oh-my-posh/console"
	"oh-my-posh/environment"
	"os"
	"path/filepath"
	"strings"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/json"
	"github.com/gookit/config/v2/toml"
	"github.com/gookit/config/v2/yaml"
	"github.com/mitchellh/mapstructure"
)

const (
	JSON string = "json"
	YAML string = "yaml"
	TOML string = "toml"
)

// Config holds all the theme for rendering the prompt
type Config struct {
	FinalSpace           bool             `json:"final_space,omitempty"`
	OSC99                bool             `json:"osc99,omitempty"`
	ConsoleTitle         bool             `json:"console_title,omitempty"`
	ConsoleTitleStyle    console.Style    `json:"console_title_style,omitempty"`
	ConsoleTitleTemplate string           `json:"console_title_template,omitempty"`
	TerminalBackground   string           `json:"terminal_background,omitempty"`
	Blocks               []*Block         `json:"blocks,omitempty"`
	Tooltips             []*Segment       `json:"tooltips,omitempty"`
	TransientPrompt      *TransientPrompt `json:"transient_prompt,omitempty"`
	Palette              color.Palette    `json:"palette,omitempty"`

	format  string
	origin  string
	eval    bool
	updated bool
}

// MakeColors creates instance of AnsiColors to use in AnsiWriter according to
// environment and configuration.
func (cfg *Config) MakeColors(env environment.Environment) color.AnsiColors {
	cacheDisabled := env.Getenv("OMP_CACHE_DISABLED") == "1"
	return color.MakeColors(cfg.Palette, !cacheDisabled)
}

type TransientPrompt struct {
	Template   string `json:"template,omitempty"`
	Background string `json:"background,omitempty"`
	Foreground string `json:"foreground,omitempty"`
}

func (cfg *Config) exitWithError(err error) {
	if err == nil {
		return
	}
	defer os.Exit(1)
	if cfg.eval {
		fmt.Println("echo \"Oh My Posh Error:\n\"", err.Error())
		return
	}
	fmt.Println("Oh My Posh Error:\n", err.Error())
}

// LoadConfig returns the default configuration including possible user overrides
func LoadConfig(env environment.Environment) *Config {
	cfg := loadConfig(env)
	return cfg
}

func loadConfig(env environment.Environment) *Config {
	var cfg Config
	configFile := *env.Args().Config
	cfg.eval = *env.Args().Eval
	if configFile == "" {
		cfg.exitWithError(errors.New("NO CONFIG"))
	}
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		cfg.exitWithError(err)
	}

	cfg.origin = configFile
	cfg.format = strings.TrimPrefix(filepath.Ext(configFile), ".")

	config.AddDriver(yaml.Driver)
	config.AddDriver(json.Driver)
	config.AddDriver(toml.Driver)
	config.WithOptions(func(opt *config.Options) {
		opt.DecoderConfig = &mapstructure.DecoderConfig{
			TagName: "json",
		}
	})

	err := config.LoadFiles(configFile)
	cfg.exitWithError(err)

	err = config.BindStruct("", &cfg)
	cfg.exitWithError(err)

	// initialize default values
	if cfg.TransientPrompt == nil {
		cfg.TransientPrompt = &TransientPrompt{}
	}

	return &cfg
}

func (cfg *Config) sync() {
	if !cfg.updated {
		return
	}
	var structMap map[string]interface{}
	inrec, _ := json2.Marshal(cfg)
	_ = json2.Unmarshal(inrec, &structMap)
	// remove empty structs
	for k, v := range structMap {
		if smap, OK := v.(map[string]interface{}); OK && len(smap) == 0 {
			delete(structMap, k)
		}
	}
	config.SetData(structMap)
}

func (cfg *Config) Export(format string) string {
	cfg.sync()

	if len(format) != 0 {
		cfg.format = format
	}

	config.AddDriver(yaml.Driver)
	config.AddDriver(toml.Driver)

	var result bytes.Buffer

	if cfg.format == JSON {
		data := config.Data()
		data["$schema"] = "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json"
		jsonEncoder := json2.NewEncoder(&result)
		jsonEncoder.SetEscapeHTML(false)
		jsonEncoder.SetIndent("", "  ")
		err := jsonEncoder.Encode(data)
		cfg.exitWithError(err)
		return escapeGlyphs(result.String())
	}

	_, err := config.DumpTo(&result, cfg.format)
	cfg.exitWithError(err)
	var prefix string
	switch cfg.format {
	case YAML:
		prefix = "# yaml-language-server: $schema=https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json\n\n"
	case TOML:
		prefix = "#:schema https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json\n\n"
	}
	return prefix + escapeGlyphs(result.String())
}

func (cfg *Config) Write() {
	content := cfg.Export(cfg.format)
	f, err := os.OpenFile(cfg.origin, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	cfg.exitWithError(err)
	_, err = f.WriteString(content)
	cfg.exitWithError(err)
	if err := f.Close(); err != nil {
		cfg.exitWithError(err)
	}
}

func escapeGlyphs(s string) string {
	var builder strings.Builder
	for _, r := range s {
		if r < 0x1000 {
			builder.WriteRune(r)
			continue
		}
		quoted := fmt.Sprintf("\\u%04x", r)
		builder.WriteString(quoted)
	}
	return builder.String()
}
