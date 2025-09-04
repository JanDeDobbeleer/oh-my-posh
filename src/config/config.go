package config

import (
	"encoding/gob"
	"slices"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/upgrade"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

func init() {
	gob.Register(&Config{})
}

const (
	JSON string = "json"
	YAML string = "yaml"
	TOML string = "toml"

	TML   string = "tml"
	YML   string = "yml"
	JSONC string = "jsonc"

	AUTOUPGRADE   = "upgrade"
	UPGRADENOTICE = "notice"
	RELOAD        = "reload"

	Version = 3
)

type Action string

func (a Action) IsDefault() bool {
	return a != Prepend && a != Extend
}

const (
	Prepend Action = "prepend"
	Extend  Action = "extend"
)

// Config holds all the theme for rendering the prompt
type Config struct {
	Palette                 color.Palette          `json:"palette,omitempty" toml:"palette,omitempty" yaml:"palette,omitempty"`
	DebugPrompt             *Segment               `json:"debug_prompt,omitempty" toml:"debug_prompt,omitempty" yaml:"debug_prompt,omitempty"`
	Var                     map[string]any         `json:"var,omitempty" toml:"var,omitempty" yaml:"var,omitempty"`
	Palettes                *color.Palettes        `json:"palettes,omitempty" toml:"palettes,omitempty" yaml:"palettes,omitempty"`
	ValidLine               *Segment               `json:"valid_line,omitempty" toml:"valid_line,omitempty" yaml:"valid_line,omitempty"`
	SecondaryPrompt         *Segment               `json:"secondary_prompt,omitempty" toml:"secondary_prompt,omitempty" yaml:"secondary_prompt,omitempty"`
	TransientPrompt         *Segment               `json:"transient_prompt,omitempty" toml:"transient_prompt,omitempty" yaml:"transient_prompt,omitempty"`
	ErrorLine               *Segment               `json:"error_line,omitempty" toml:"error_line,omitempty" yaml:"error_line,omitempty"`
	Maps                    *maps.Config           `json:"maps,omitempty" toml:"maps,omitempty" yaml:"maps,omitempty"`
	Upgrade                 *upgrade.Config        `json:"upgrade,omitempty" toml:"upgrade,omitempty" yaml:"upgrade,omitempty"`
	Extends                 string                 `json:"extends,omitempty" toml:"extends,omitempty" yaml:"extends,omitempty"`
	AccentColor             color.Ansi             `json:"accent_color,omitempty" toml:"accent_color,omitempty" yaml:"accent_color,omitempty"`
	ConsoleTitleTemplate    string                 `json:"console_title_template,omitempty" toml:"console_title_template,omitempty" yaml:"console_title_template,omitempty"`
	PWD                     string                 `json:"pwd,omitempty" toml:"pwd,omitempty" yaml:"pwd,omitempty"`
	Source                  string                 `json:"-" toml:"-" yaml:"-"`
	Format                  string                 `json:"-" toml:"-" yaml:"-"`
	TerminalBackground      color.Ansi             `json:"terminal_background,omitempty" toml:"terminal_background,omitempty" yaml:"terminal_background,omitempty"`
	ToolTipsAction          Action                 `json:"tooltips_action,omitempty" toml:"tooltips_action,omitempty" yaml:"tooltips_action,omitempty"`
	Tooltips                []*Segment             `json:"tooltips,omitempty" toml:"tooltips,omitempty" yaml:"tooltips,omitempty"`
	Blocks                  []*Block               `json:"blocks,omitempty" toml:"blocks,omitempty" yaml:"blocks,omitempty"`
	ITermFeatures           terminal.ITermFeatures `json:"iterm_features,omitempty" toml:"iterm_features,omitempty" yaml:"iterm_features,omitempty"`
	Cycle                   color.Cycle            `json:"cycle,omitempty" toml:"cycle,omitempty" yaml:"cycle,omitempty"`
	Version                 int                    `json:"version" toml:"version" yaml:"version"`
	EnableCursorPositioning bool                   `json:"enable_cursor_positioning,omitempty" toml:"enable_cursor_positioning,omitempty" yaml:"enable_cursor_positioning,omitempty"`
	ShellIntegration        bool                   `json:"shell_integration,omitempty" toml:"shell_integration,omitempty" yaml:"shell_integration,omitempty"`
	MigrateGlyphs           bool                   `json:"-" toml:"-" yaml:"-"`
	Async                   bool                   `json:"async,omitempty" toml:"async,omitempty" yaml:"async,omitempty"`
	FinalSpace              bool                   `json:"final_space,omitempty" toml:"final_space,omitempty" yaml:"final_space,omitempty"`
	UpgradeNotice           bool                   `json:"-" toml:"-" yaml:"-"`
	updated                 bool
	extended                bool
	PatchPwshBleed          bool `json:"patch_pwsh_bleed,omitempty" toml:"patch_pwsh_bleed,omitempty" yaml:"patch_pwsh_bleed,omitempty"`
	AutoUpgrade             bool `json:"-" toml:"-" yaml:"-"`
}

func (cfg *Config) MakeColors(env runtime.Environment) color.String {
	cacheDisabled := env.Getenv("OMP_CACHE_DISABLED") == "1"
	return color.MakeColors(cfg.getPalette(), !cacheDisabled, cfg.AccentColor, env)
}

func (cfg *Config) getPalette() color.Palette {
	if cfg.Palettes == nil {
		return cfg.Palette
	}

	key, err := template.Render(cfg.Palettes.Template, nil)
	if err != nil {
		return cfg.Palette
	}

	palette, ok := cfg.Palettes.List[key]
	if !ok {
		return cfg.Palette
	}

	for key, color := range cfg.Palette {
		if _, ok := palette[key]; ok {
			continue
		}

		palette[key] = color
	}

	return palette
}

func (cfg *Config) Features(env runtime.Environment) shell.Features {
	var feats shell.Features

	asyncShells := []string{shell.BASH, shell.ZSH, shell.FISH, shell.PWSH, shell.PWSH5}

	if cfg.Async && slices.Contains(asyncShells, env.Shell()) {
		log.Debug("async enabled")
		feats |= shell.Async
	}

	if cfg.TransientPrompt != nil {
		log.Debug("transient prompt enabled")
		feats |= shell.Transient
	}

	unsupportedShells := []string{shell.ELVISH, shell.XONSH}
	if slices.Contains(unsupportedShells, env.Shell()) {
		cfg.ShellIntegration = false
	}

	if cfg.ShellIntegration {
		log.Debug("shell integration enabled")
		feats |= shell.FTCSMarks
	}

	// do not enable upgrade features when async is enabled
	if feats&shell.Async == 0 {
		feats |= cfg.upgradeFeatures()
	}

	if cfg.ErrorLine != nil || cfg.ValidLine != nil {
		log.Debug("error or valid line enabled")
		feats |= shell.LineError
	}

	if len(cfg.Tooltips) > 0 {
		log.Debug("tooltips enabled")
		feats |= shell.Tooltips
	}

	if env.Shell() == shell.FISH && cfg.ITermFeatures != nil && cfg.ITermFeatures.Contains(terminal.PromptMark) {
		log.Debug("prompt mark enabled")
		feats |= shell.PromptMark
	}

	for i, block := range cfg.Blocks {
		if (i == 0 && block.Newline) && cfg.EnableCursorPositioning {
			log.Debug("cursor positioning enabled")
			feats |= shell.CursorPositioning
		}

		if block.Type == RPrompt {
			log.Debug("rprompt enabled")
			feats |= shell.RPrompt
		}

		for _, segment := range block.Segments {
			if segment.Type == AZ {
				source := segment.Properties.GetString(segments.Source, segments.FirstMatch)
				if strings.Contains(source, segments.Pwsh) {
					log.Debug("azure enabled")
					feats |= shell.Azure
				}
			}

			if segment.Type == GIT {
				source := segment.Properties.GetString(segments.Source, segments.Cli)
				if source == segments.Pwsh {
					log.Debug("posh-git enabled")
					feats |= shell.PoshGit
				}
			}
		}
	}

	return feats
}

func (cfg *Config) upgradeFeatures() shell.Features {
	var feats shell.Features

	if _, OK := cache.Get[string](cache.Device, upgrade.CACHEKEY); OK && !cfg.Upgrade.Force {
		log.Debug("upgrade cache key found and not forced, skipping upgrade")
		return feats
	}

	autoUpgrade := cfg.Upgrade.Auto
	if _, OK := cache.Get[string](cache.Device, AUTOUPGRADE); OK {
		log.Debug("auto upgrade key found")
		autoUpgrade = true
	}

	upgradeNotice := cfg.Upgrade.DisplayNotice
	if _, OK := cache.Get[string](cache.Device, UPGRADENOTICE); OK {
		log.Debug("upgrade notice key found")
		upgradeNotice = true
	}

	if upgradeNotice && !autoUpgrade {
		log.Debug("notice enabled, no auto upgrade")
		feats |= shell.Notice
	}

	if autoUpgrade {
		log.Debug("auto upgrade enabled")
		feats |= shell.Upgrade
	}

	return feats
}
