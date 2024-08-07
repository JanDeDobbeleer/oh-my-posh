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
	env                     runtime.Environment
	Palette                 color.Palette   `json:"palette,omitempty" toml:"palette,omitempty"`
	DebugPrompt             *Segment        `json:"debug_prompt,omitempty" toml:"debug_prompt,omitempty"`
	Var                     map[string]any  `json:"var,omitempty" toml:"var,omitempty"`
	Palettes                *color.Palettes `json:"palettes,omitempty" toml:"palettes,omitempty"`
	ValidLine               *Segment        `json:"valid_line,omitempty" toml:"valid_line,omitempty"`
	SecondaryPrompt         *Segment        `json:"secondary_prompt,omitempty" toml:"secondary_prompt,omitempty"`
	TransientPrompt         *Segment        `json:"transient_prompt,omitempty" toml:"transient_prompt,omitempty"`
	ErrorLine               *Segment        `json:"error_line,omitempty" toml:"error_line,omitempty"`
	ConsoleTitleTemplate    string          `json:"console_title_template,omitempty" toml:"console_title_template,omitempty"`
	Format                  string          `json:"-" toml:"-"`
	origin                  string
	PWD                     string                 `json:"pwd,omitempty" toml:"pwd,omitempty"`
	AccentColor             color.Ansi             `json:"accent_color,omitempty" toml:"accent_color,omitempty"`
	Output                  string                 `json:"-" toml:"-"`
	TerminalBackground      color.Ansi             `json:"terminal_background,omitempty" toml:"terminal_background,omitempty"`
	Cycle                   color.Cycle            `json:"cycle,omitempty" toml:"cycle,omitempty"`
	ITermFeatures           terminal.ITermFeatures `json:"iterm_features,omitempty" toml:"iterm_features,omitempty"`
	Blocks                  []*Block               `json:"blocks,omitempty" toml:"blocks,omitempty"`
	Tooltips                []*Segment             `json:"tooltips,omitempty" toml:"tooltips,omitempty"`
	Version                 int                    `json:"version" toml:"version"`
	DisableNotice           bool                   `json:"disable_notice,omitempty" toml:"disable_notice,omitempty"`
	AutoUpgrade             bool                   `json:"auto_upgrade,omitempty" toml:"auto_upgrade,omitempty"`
	ShellIntegration        bool                   `json:"shell_integration,omitempty" toml:"shell_integration,omitempty"`
	MigrateGlyphs           bool                   `json:"-" toml:"-"`
	PatchPwshBleed          bool                   `json:"patch_pwsh_bleed,omitempty" toml:"patch_pwsh_bleed,omitempty"`
	EnableCursorPositioning bool                   `json:"enable_cursor_positioning,omitempty" toml:"enable_cursor_positioning,omitempty"`
	updated                 bool
	FinalSpace              bool `json:"final_space,omitempty" toml:"final_space,omitempty"`
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
	}

	key, err := tmpl.Render()
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
