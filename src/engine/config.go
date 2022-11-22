package engine

import (
	"bytes"
	json2 "encoding/json"
	"fmt"
	"io"
	"oh-my-posh/color"
	"oh-my-posh/platform"
	"oh-my-posh/properties"
	"oh-my-posh/segments"
	"oh-my-posh/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/json"
	"github.com/gookit/config/v2/toml"
	yaml "github.com/gookit/config/v2/yamlv3"
	"github.com/mitchellh/mapstructure"
)

const (
	JSON string = "json"
	YAML string = "yaml"
	TOML string = "toml"

	configVersion = 2
)

// Config holds all the theme for rendering the prompt
type Config struct {
	Version              int             `json:"version"`
	FinalSpace           bool            `json:"final_space,omitempty"`
	ConsoleTitleTemplate string          `json:"console_title_template,omitempty"`
	TerminalBackground   string          `json:"terminal_background,omitempty"`
	AccentColor          string          `json:"accent_color,omitempty"`
	Blocks               []*Block        `json:"blocks,omitempty"`
	Tooltips             []*Segment      `json:"tooltips,omitempty"`
	TransientPrompt      *Segment        `json:"transient_prompt,omitempty"`
	ValidLine            *Segment        `json:"valid_line,omitempty"`
	ErrorLine            *Segment        `json:"error_line,omitempty"`
	SecondaryPrompt      *Segment        `json:"secondary_prompt,omitempty"`
	DebugPrompt          *Segment        `json:"debug_prompt,omitempty"`
	Palette              color.Palette   `json:"palette,omitempty"`
	Palettes             *color.Palettes `json:"palettes,omitempty"`
	PWD                  string          `json:"pwd,omitempty"`

	// Deprecated
	OSC99 bool `json:"osc99,omitempty"`

	Output string `json:"-"`

	format string
	origin string
	// eval    bool
	updated bool
	env     platform.Environment
}

// MakeColors creates instance of AnsiColors to use in AnsiWriter according to
// environment and configuration.
func (cfg *Config) MakeColors() color.AnsiColors {
	cacheDisabled := cfg.env.Getenv("OMP_CACHE_DISABLED") == "1"
	return color.MakeColors(cfg.getPalette(), !cacheDisabled, cfg.AccentColor, cfg.env)
}

func (cfg *Config) getPalette() color.Palette {
	if cfg.Palettes == nil {
		return cfg.Palette
	}
	tmpl := &template.Text{
		Template: cfg.Palettes.Template,
		Env:      cfg.env,
	}
	if palette, err := tmpl.Render(); err == nil {
		if p, ok := cfg.Palettes.List[palette]; ok {
			return p
		}
	}
	return cfg.Palette
}

// LoadConfig returns the default configuration including possible user overrides
func LoadConfig(env platform.Environment) *Config {
	cfg := loadConfig(env)
	cfg.env = env
	// only migrate automatically when the switch isn't set
	if !env.Flags().Migrate && cfg.Version < configVersion {
		cfg.BackupAndMigrate(env)
	}
	return cfg
}

func loadConfig(env platform.Environment) *Config {
	defer env.Trace(time.Now(), "config.loadConfig")
	configFile := env.Flags().Config

	if len(configFile) == 0 {
		return defaultConfig(false)
	}

	var cfg Config
	cfg.origin = configFile
	cfg.format = strings.TrimPrefix(filepath.Ext(configFile), ".")
	if cfg.format == "yml" {
		cfg.format = YAML
	}

	config.AddDriver(yaml.Driver)
	config.AddDriver(json.Driver)
	config.AddDriver(toml.Driver)
	config.WithOptions(func(opt *config.Options) {
		opt.DecoderConfig = &mapstructure.DecoderConfig{
			TagName: "json",
		}
	})

	err := config.LoadFiles(configFile)
	if err != nil {
		return defaultConfig(true)
	}

	err = config.BindStruct("", &cfg)
	if err != nil {
		return defaultConfig(true)
	}

	return &cfg
}

func (cfg *Config) sync() {
	if !cfg.updated {
		return
	}
	var structMap map[string]interface{}
	inrec, err := json2.Marshal(cfg)
	if err != nil {
		return
	}
	err = json2.Unmarshal(inrec, &structMap)
	if err != nil {
		return
	}
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
		jsonEncoder := json2.NewEncoder(&result)
		jsonEncoder.SetEscapeHTML(false)
		jsonEncoder.SetIndent("", "  ")
		_ = jsonEncoder.Encode(cfg)
		prefix := "{\n  \"$schema\": \"https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json\","
		data := strings.Replace(result.String(), "{", prefix, 1)
		return escapeGlyphs(data)
	}

	_, _ = config.DumpTo(&result, cfg.format)
	switch cfg.format {
	case YAML:
		prefix := "# yaml-language-server: $schema=https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json\n\n"
		return prefix + result.String()
	case TOML:
		prefix := "#:schema https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json\n\n"
		return prefix + escapeGlyphs(result.String())
	default:
		return result.String()
	}
}

func (cfg *Config) BackupAndMigrate(env platform.Environment) {
	cfg.backup()
	cfg.Migrate(env)
	cfg.Write(cfg.format)
}

func (cfg *Config) Write(format string) {
	content := cfg.Export(format)
	destination := cfg.Output
	if len(destination) == 0 {
		destination = cfg.origin
	}
	f, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return
	}
	_, err = f.WriteString(content)
	if err != nil {
		return
	}
	_ = f.Close()
}

func (cfg *Config) backup() {
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

func defaultConfig(warning bool) *Config {
	exitBackgroundTemplate := "{{ if gt .Code 0 }}p:red{{ end }}"
	exitTemplate := " {{ if gt .Code 0 }}\uf00d{{ else }}\uf00c{{ end }} "
	if warning {
		exitBackgroundTemplate = "p:red"
		exitTemplate = " CONFIG ERROR "
	}
	cfg := &Config{
		Version:    2,
		FinalSpace: true,
		Blocks: []*Block{
			{
				Type:      Prompt,
				Alignment: Left,
				Segments: []*Segment{
					{
						Type:            SESSION,
						Style:           Diamond,
						LeadingDiamond:  "\ue0b6",
						TrailingDiamond: "\ue0b0",
						Background:      "p:yellow",
						Foreground:      "p:black",
						Template:        " {{ if .SSHSession }}\uf817 {{ end }}{{ .UserName }} ",
					},
					{
						Type:            PATH,
						Style:           Powerline,
						PowerlineSymbol: "\ue0b0",
						Background:      "p:orange",
						Foreground:      "p:white",
						Properties: properties.Map{
							properties.Style: "folder",
						},
						Template: " \uf74a {{ path .Path .Location }} ",
					},
					{
						Type:            GIT,
						Style:           Powerline,
						PowerlineSymbol: "\ue0b0",
						Background:      "p:green",
						BackgroundTemplates: []string{
							"{{ if or (.Working.Changed) (.Staging.Changed) }}p:yellow{{ end }}",
							"{{ if and (gt .Ahead 0) (gt .Behind 0) }}p:red{{ end }}",
							"{{ if gt .Ahead 0 }}#49416D{{ end }}",
							"{{ if gt .Behind 0 }}#7A306C{{ end }}",
						},
						Foreground: "p:black",
						ForegroundTemplates: []string{
							"{{ if or (.Working.Changed) (.Staging.Changed) }}p:black{{ end }}",
							"{{ if and (gt .Ahead 0) (gt .Behind 0) }}p:white{{ end }}",
							"{{ if gt .Ahead 0 }}p:white{{ end }}",
						},
						Properties: properties.Map{
							segments.BranchMaxLength:   25,
							segments.FetchStatus:       true,
							segments.FetchUpstreamIcon: true,
							segments.GithubIcon:        "\uf7a3",
						},
						Template: " {{ if .UpstreamURL }}{{ url .UpstreamIcon .UpstreamURL }} {{ end }}{{ .HEAD }}{{if .BranchStatus }} {{ .BranchStatus }}{{ end }}{{ if .Working.Changed }} \uf044 {{ .Working.String }}{{ end }}{{ if .Staging.Changed }} \uf046 {{ .Staging.String }}{{ end }} ", //nolint:lll
					},
					{
						Type:            ROOT,
						Style:           Powerline,
						PowerlineSymbol: "\ue0b0",
						Background:      "p:yellow",
						Foreground:      "p:white",
						Template:        " \uf0e7 ",
					},
					{
						Type:            EXIT,
						Style:           Diamond,
						LeadingDiamond:  "<transparent,background>\ue0b0</>",
						TrailingDiamond: "\ue0b4",
						Background:      "p:blue",
						BackgroundTemplates: []string{
							exitBackgroundTemplate,
						},
						Foreground: "p:white",
						Properties: properties.Map{
							properties.AlwaysEnabled: true,
						},
						Template: exitTemplate,
					},
				},
			},
			{
				Type: RPrompt,
				Segments: []*Segment{
					{
						Type:       NODE,
						Style:      Plain,
						Background: "transparent",
						Foreground: "p:green",
						Template:   "\uf898 ",
						Properties: properties.Map{
							segments.HomeEnabled:         false,
							segments.FetchPackageManager: false,
							segments.DisplayMode:         "files",
						},
					},
					{
						Type:       GOLANG,
						Style:      Plain,
						Background: "transparent",
						Foreground: "p:blue",
						Template:   "\ufcd1 ",
						Properties: properties.Map{
							properties.FetchVersion: false,
						},
					},
					{
						Type:       PYTHON,
						Style:      Plain,
						Background: "transparent",
						Foreground: "p:yellow",
						Template:   "\ue235 ",
						Properties: properties.Map{
							properties.FetchVersion:  false,
							segments.DisplayMode:     "files",
							segments.FetchVirtualEnv: false,
						},
					},
					{
						Type:       SHELL,
						Style:      Plain,
						Background: "transparent",
						Foreground: "p:white",
						Template:   "in <p:blue><b>{{ .Name }}</b></> ",
					},
					{
						Type:       TIME,
						Style:      Plain,
						Background: "transparent",
						Foreground: "p:white",
						Template:   "at <p:blue><b>{{ .CurrentDate | date \"15:04:05\" }}</b></>",
					},
				},
			},
		},
		ConsoleTitleTemplate: "{{ .Shell }} in {{ .Folder }}",
		Palette: color.Palette{
			"black":  "#262B44",
			"blue":   "#4B95E9",
			"green":  "#59C9A5",
			"orange": "#F07623",
			"red":    "#D81E5B",
			"white":  "#E0DEF4",
			"yellow": "#F3AE35",
		},
		SecondaryPrompt: &Segment{
			Background: "transparent",
			Foreground: "p:black",
			Template:   "<p:yellow,transparent>\ue0b6</><,p:yellow> > </><p:yellow,transparent>\ue0b0</> ",
		},
		TransientPrompt: &Segment{
			Background: "transparent",
			Foreground: "p:black",
			Template:   "<p:yellow,transparent>\ue0b6</><,p:yellow> {{ .Folder }} </><p:yellow,transparent>\ue0b0</> ",
		},
		Tooltips: []*Segment{
			{
				Type:            AWS,
				Style:           Diamond,
				LeadingDiamond:  "\ue0b0",
				TrailingDiamond: "\ue0b4",
				Background:      "p:orange",
				Foreground:      "p:white",
				Template:        " \ue7ad {{ .Profile }}{{ if .Region }}@{{ .Region }}{{ end }} ",
				Properties: properties.Map{
					properties.DisplayDefault: true,
				},
				Tips: []string{"aws"},
			},
			{
				Type:            AZ,
				Style:           Diamond,
				LeadingDiamond:  "\ue0b0",
				TrailingDiamond: "\ue0b4",
				Background:      "p:blue",
				Foreground:      "p:white",
				Template:        " \ufd03 {{ .Name }} ",
				Properties: properties.Map{
					properties.DisplayDefault: true,
				},
				Tips: []string{"az"},
			},
		},
	}
	return cfg
}
