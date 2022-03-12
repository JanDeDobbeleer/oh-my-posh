package engine

import (
	"bytes"
	json2 "encoding/json"
	"fmt"
	"io"
	"oh-my-posh/color"
	"oh-my-posh/console"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
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

	configVersion = 1
)

// Config holds all the theme for rendering the prompt
type Config struct {
	Version              int           `json:"version"`
	FinalSpace           bool          `json:"final_space,omitempty"`
	OSC99                bool          `json:"osc99,omitempty"`
	ConsoleTitle         bool          `json:"console_title,omitempty"`
	ConsoleTitleStyle    console.Style `json:"console_title_style,omitempty"`
	ConsoleTitleTemplate string        `json:"console_title_template,omitempty"`
	TerminalBackground   string        `json:"terminal_background,omitempty"`
	Blocks               []*Block      `json:"blocks,omitempty"`
	Tooltips             []*Segment    `json:"tooltips,omitempty"`
	TransientPrompt      *ExtraPrompt  `json:"transient_prompt,omitempty"`
	ValidLine            *ExtraPrompt  `json:"valid_line,omitempty"`
	ErrorLine            *ExtraPrompt  `json:"error_line,omitempty"`
	SecondaryPrompt      *ExtraPrompt  `json:"secondary_prompt,omitempty"`
	Palette              color.Palette `json:"palette,omitempty"`

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

type ExtraPrompt struct {
	Template   string `json:"template,omitempty"`
	Background string `json:"background,omitempty"`
	Foreground string `json:"foreground,omitempty"`
}

func (cfg *Config) print(message string) {
	if cfg.eval {
		fmt.Printf("echo \"%s\"", message)
		return
	}
	fmt.Println(message)
}

func (cfg *Config) exitWithError(err error) {
	if err == nil {
		return
	}
	defer os.Exit(1)
	message := "Oh My Posh Error:\n\n" + err.Error()
	cfg.print(message)
}

// LoadConfig returns the default configuration including possible user overrides
func LoadConfig(env environment.Environment) *Config {
	cfg := loadConfig(env)
	// only migrate automatically when the switch isn't set
	if !env.Flags().Migrate && cfg.Version != configVersion {
		cfg.BackupAndMigrate(env)
	}
	return cfg
}

func loadConfig(env environment.Environment) *Config {
	var cfg Config
	configFile := env.Flags().Config
	if configFile == "" {
		return defaultConfig()
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
		cfg.TransientPrompt = &ExtraPrompt{}
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

func (cfg *Config) BackupAndMigrate(env environment.Environment) {
	origin := cfg.backup()
	cfg.Migrate(env)
	cfg.Write()
	cfg.print(fmt.Sprintf("\nOh My Posh config migrated to version %d\nBackup config available at %s\n\n", cfg.Version, origin))
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

func (cfg *Config) backup() string {
	dst := cfg.origin + ".bak"
	source, err := os.Open(cfg.origin)
	cfg.exitWithError(err)
	defer source.Close()
	destination, err := os.Create(dst)
	cfg.exitWithError(err)
	defer destination.Close()
	_, err = io.Copy(destination, source)
	cfg.exitWithError(err)
	return dst
}

func escapeGlyphs(s string) string {
	var builder strings.Builder
	for _, r := range s {
		// exclude regular characters and emoji
		if r < 0x1000 || r > 0x10000 {
			builder.WriteRune(r)
			continue
		}
		quoted := fmt.Sprintf("\\u%04x", r)
		builder.WriteString(quoted)
	}
	return builder.String()
}

func defaultConfig() *Config {
	cfg := &Config{
		Version:    1,
		FinalSpace: true,
		Blocks: []*Block{
			{
				Type:      Prompt,
				Alignment: Left,
				Segments: []*Segment{
					{
						Type:            SESSION,
						Style:           Diamond,
						Background:      "#c386f1",
						Foreground:      "#ffffff",
						LeadingDiamond:  "\uE0B6",
						TrailingDiamond: "\uE0B0",
					},
					{
						Type:            PATH,
						Style:           Powerline,
						PowerlineSymbol: "\uE0B0",
						Background:      "#ff479c",
						Foreground:      "#ffffff",
						Properties: properties.Map{
							properties.Style: "folder",
						},
					},
					{
						Type:            SHELL,
						Style:           Powerline,
						PowerlineSymbol: "\uE0B0",
						Background:      "#0077c2",
						Foreground:      "#ffffff",
					},
					{
						Type:            TEXT,
						Style:           Powerline,
						PowerlineSymbol: "\uE0B0",
						Background:      "#ffffff",
						Foreground:      "#111111",
						Properties: properties.Map{
							properties.SegmentTemplate: " no config ",
						},
					},
					{
						Type:            EXIT,
						Style:           Diamond,
						Background:      "#2e9599",
						Foreground:      "#ffffff",
						LeadingDiamond:  "<transparent,background>\uE0B0</>",
						TrailingDiamond: "\uE0B4",
						BackgroundTemplates: []string{
							"{{ if gt .Code 0 }}#f1184c{{ end }}",
						},
						Properties: properties.Map{
							properties.AlwaysEnabled:   true,
							properties.SegmentTemplate: " \uE23A",
						},
					},
				},
			},
		},
	}
	return cfg
}
