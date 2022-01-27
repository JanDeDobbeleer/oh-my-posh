package engine

import (
	// "encoding/json"

	"bytes"
	json2 "encoding/json"
	"errors"
	"fmt"
	"oh-my-posh/color"
	"oh-my-posh/console"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"oh-my-posh/segments"
	"os"
	"strconv"
	"strings"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/json"
	"github.com/gookit/config/v2/toml"
	"github.com/gookit/config/v2/yaml"
	"github.com/mitchellh/mapstructure"
)

// Config holds all the theme for rendering the prompt
type Config struct {
	FinalSpace           bool             `config:"final_space"`
	OSC99                bool             `config:"osc99"`
	ConsoleTitle         bool             `config:"console_title"`
	ConsoleTitleStyle    console.Style    `config:"console_title_style"`
	ConsoleTitleTemplate string           `config:"console_title_template"`
	TerminalBackground   string           `config:"terminal_background"`
	Blocks               []*Block         `config:"blocks"`
	Tooltips             []*Segment       `config:"tooltips"`
	TransientPrompt      *TransientPrompt `config:"transient_prompt"`
	Palette              color.Palette    `config:"palette"`
}

// MakeColors creates instance of AnsiColors to use in AnsiWriter according to
// environment and configuration.
func (cfg *Config) MakeColors(env environment.Environment) color.AnsiColors {
	cacheDisabled := env.Getenv("OMP_CACHE_DISABLED") == "1"
	return color.MakeColors(cfg.Palette, !cacheDisabled)
}

type TransientPrompt struct {
	Template   string `config:"template"`
	Background string `config:"background"`
	Foreground string `config:"foreground"`
}

func printConfigError(err error, eval bool) {
	if eval {
		fmt.Println("echo \"Oh My Posh Error:\n\"", err.Error())
		return
	}
	fmt.Println("Oh My Posh Error:\n", err.Error())
}

// GetConfig returns the default configuration including possible user overrides
func GetConfig(env environment.Environment) *Config {
	cfg, err := loadConfig(env)
	if err != nil {
		return getDefaultConfig(err.Error())
	}
	return cfg
}

func loadConfig(env environment.Environment) (*Config, error) {
	var cfg Config
	configFile := *env.Args().Config
	eval := *env.Args().Eval
	if configFile == "" {
		return nil, errors.New("NO CONFIG")
	}
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		printConfigError(err, eval)
		return nil, errors.New("INVALID CONFIG PATH")
	}

	config.AddDriver(yaml.Driver)
	config.AddDriver(json.Driver)
	config.AddDriver(toml.Driver)
	config.WithOptions(func(opt *config.Options) {
		opt.DecoderConfig = &mapstructure.DecoderConfig{
			TagName: "config",
		}
	})

	err := config.LoadFiles(configFile)
	if err != nil {
		printConfigError(err, eval)
		return nil, errors.New("UNABLE TO OPEN CONFIG")
	}

	err = config.BindStruct("", &cfg)
	if err != nil {
		printConfigError(err, eval)
		return nil, errors.New("INVALID CONFIG")
	}

	// initialize default values
	if cfg.TransientPrompt == nil {
		cfg.TransientPrompt = &TransientPrompt{}
	}

	return &cfg, nil
}

func ExportConfig(configFile, format string) string {
	if len(format) == 0 {
		format = config.JSON
	}

	config.AddDriver(yaml.Driver)
	config.AddDriver(json.Driver)
	config.AddDriver(toml.Driver)

	err := config.LoadFiles(configFile)
	if err != nil {
		printConfigError(err, false)
		return fmt.Sprintf("INVALID CONFIG:\n\n%s", err.Error())
	}

	schemaKey := "$schema"
	if format == config.JSON && !config.Exists(schemaKey) {
		data := config.Data()
		data[schemaKey] = "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json"
		config.SetData(data)
	}

	buf := new(bytes.Buffer)
	_, err = config.DumpTo(buf, format)
	if err != nil {
		printConfigError(err, false)
		return "UNABLE TO DUMP CONFIG"
	}

	switch format {
	case config.JSON:
		var prettyJSON bytes.Buffer
		err := json2.Indent(&prettyJSON, buf.Bytes(), "", "  ")
		if err == nil {
			unescapeUnicodeCharactersInJSON := func(rawJSON []byte) string {
				str, err := strconv.Unquote(strings.ReplaceAll(strconv.Quote(string(rawJSON)), `\\u`, `\u`))
				if err != nil {
					return err.Error()
				}
				return str
			}
			return unescapeUnicodeCharactersInJSON(prettyJSON.Bytes())
		}
	case config.Yaml:
		prefix := "# yaml-language-server: $schema=https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json\n\n"
		content := buf.String()
		return prefix + content

	case config.Toml:
		prefix := "#:schema https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json\n\n"
		content := buf.String()
		return prefix + content
	}

	return buf.String()
}

func getDefaultConfig(info string) *Config {
	cfg := &Config{
		FinalSpace:        true,
		ConsoleTitle:      true,
		ConsoleTitleStyle: console.FolderName,
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
							properties.Prefix: " \uE5FF ",
							properties.Style:  "folder",
						},
					},
					{
						Type:            GIT,
						Style:           Powerline,
						PowerlineSymbol: "\uE0B0",
						Background:      "#fffb38",
						Foreground:      "#193549",
						Properties: properties.Map{
							segments.FetchStashCount:   true,
							segments.FetchUpstreamIcon: true,
						},
					},
					{
						Type:            BATTERY,
						Style:           Powerline,
						PowerlineSymbol: "\uE0B0",
						Background:      "#f36943",
						Foreground:      "#193549",
						Properties: properties.Map{
							properties.Postfix: "\uF295 ",
						},
					},
					{
						Type:            NODE,
						Style:           Powerline,
						PowerlineSymbol: "\uE0B0",
						Background:      "#6CA35E",
						Foreground:      "#ffffff",
						Properties: properties.Map{
							properties.Prefix:       " \uE718",
							properties.FetchVersion: false,
						},
					},
					{
						Type:            SHELL,
						Style:           Powerline,
						PowerlineSymbol: "\uE0B0",
						Background:      "#0077c2",
						Foreground:      "#ffffff",
						Properties: properties.Map{
							properties.Prefix: " \uFCB5 ",
						},
					},
					{
						Type:            ROOT,
						Style:           Powerline,
						PowerlineSymbol: "\uE0B0",
						Background:      "#ffff66",
						Foreground:      "#ffffff",
					},
					{
						Type:            TEXT,
						Style:           Powerline,
						PowerlineSymbol: "\uE0B0",
						Background:      "#ffffff",
						Foreground:      "#111111",
						Properties: properties.Map{
							properties.SegmentTemplate: info,
						},
					},
					{
						Type:            EXIT,
						Style:           Diamond,
						Background:      "#2e9599",
						Foreground:      "#ffffff",
						LeadingDiamond:  "<transparent,#2e9599>\uE0B0</>",
						TrailingDiamond: "\uE0B4",
						Properties: properties.Map{
							properties.AlwaysEnabled: true,
							properties.Prefix:        " \uE23A",
						},
					},
				},
			},
		},
	}
	return cfg
}
