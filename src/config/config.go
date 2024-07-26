package config

import (
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

const (
	JSON string = "json"
	YAML string = "yaml"
	TOML string = "toml"

	Version = 2
)

// Config holds all the theme for rendering the prompt
type Config struct {
	Version                 int                    `json:"version" toml:"version"`
	FinalSpace              bool                   `json:"final_space,omitempty" toml:"final_space,omitempty"`
	ConsoleTitleTemplate    string                 `json:"console_title_template,omitempty" toml:"console_title_template,omitempty"`
	TerminalBackground      color.Ansi             `json:"terminal_background,omitempty" toml:"terminal_background,omitempty"`
	AccentColor             color.Ansi             `json:"accent_color,omitempty" toml:"accent_color,omitempty"`
	Blocks                  []*Block               `json:"blocks,omitempty" toml:"blocks,omitempty"`
	Tooltips                []*Segment             `json:"tooltips,omitempty" toml:"tooltips,omitempty"`
	TransientPrompt         *Segment               `json:"transient_prompt,omitempty" toml:"transient_prompt,omitempty"`
	ValidLine               *Segment               `json:"valid_line,omitempty" toml:"valid_line,omitempty"`
	ErrorLine               *Segment               `json:"error_line,omitempty" toml:"error_line,omitempty"`
	SecondaryPrompt         *Segment               `json:"secondary_prompt,omitempty" toml:"secondary_prompt,omitempty"`
	DebugPrompt             *Segment               `json:"debug_prompt,omitempty" toml:"debug_prompt,omitempty"`
	Palette                 color.Palette          `json:"palette,omitempty" toml:"palette,omitempty"`
	Palettes                *color.Palettes        `json:"palettes,omitempty" toml:"palettes,omitempty"`
	Cycle                   color.Cycle            `json:"cycle,omitempty" toml:"cycle,omitempty"`
	ShellIntegration        bool                   `json:"shell_integration,omitempty" toml:"shell_integration,omitempty"`
	PWD                     string                 `json:"pwd,omitempty" toml:"pwd,omitempty"`
	Var                     map[string]any         `json:"var,omitempty" toml:"var,omitempty"`
	EnableCursorPositioning bool                   `json:"enable_cursor_positioning,omitempty" toml:"enable_cursor_positioning,omitempty"`
	PatchPwshBleed          bool                   `json:"patch_pwsh_bleed,omitempty" toml:"patch_pwsh_bleed,omitempty"`
	DisableNotice           bool                   `json:"disable_notice,omitempty" toml:"disable_notice,omitempty"`
	AutoUpgrade             bool                   `json:"auto_upgrade,omitempty" toml:"auto_upgrade,omitempty"`
	ITermFeatures           terminal.ITermFeatures `json:"iterm_features,omitempty" toml:"iterm_features,omitempty"`

	// Deprecated
	OSC99 bool `json:"osc99,omitempty" toml:"osc99,omitempty"`

	Output        string `json:"-" toml:"-"`
	MigrateGlyphs bool   `json:"-" toml:"-"`
	Format        string `json:"-" toml:"-"`

	origin string
	// eval    bool
	updated bool
	env     runtime.Environment
}

func (cfg *Config) MakeColors() color.String {
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

func (cfg *Config) Features() shell.Features {
	var feats shell.Features

	if cfg.TransientPrompt != nil {
		feats = append(feats, shell.Transient)
	}

	if cfg.ShellIntegration {
		feats = append(feats, shell.FTCSMarks)
	}

	if !cfg.AutoUpgrade && !cfg.DisableNotice {
		feats = append(feats, shell.Notice)
	}

	if cfg.AutoUpgrade {
		feats = append(feats, shell.Upgrade)
	}

	if cfg.ErrorLine != nil || cfg.ValidLine != nil {
		feats = append(feats, shell.LineError)
	}

	if len(cfg.Tooltips) > 0 {
		feats = append(feats, shell.Tooltips)
	}

	if cfg.env.Shell() == shell.FISH && cfg.ITermFeatures != nil && cfg.ITermFeatures.Contains(terminal.PromptMark) {
		feats = append(feats, shell.PromptMark)
	}

	for i, block := range cfg.Blocks {
		if (i == 0 && block.Newline) && cfg.EnableCursorPositioning {
			feats = append(feats, shell.CursorPositioning)
		}

		if block.Type == RPrompt {
			feats = append(feats, shell.RPrompt)
		}

		for _, segment := range block.Segments {
			if segment.Type == AZ {
				source := segment.Properties.GetString(segments.Source, segments.FirstMatch)
				if source == segments.Pwsh || source == segments.FirstMatch {
					feats = append(feats, shell.Azure)
				}
			}

			if segment.Type == GIT {
				source := segment.Properties.GetString(segments.Source, segments.Cli)
				if source == segments.Pwsh {
					feats = append(feats, shell.PoshGit)
				}
			}
		}
	}

	return feats
}
