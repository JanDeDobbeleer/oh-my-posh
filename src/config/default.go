package config

import (
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/upgrade"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

const (
	paletteBlack          = "p:black"
	paletteBlue           = "p:blue"
	paletteGreen          = "p:green"
	paletteOrange         = "p:orange"
	paletteWhite          = "p:white"
	paletteYellow         = "p:yellow"
	backgroundTransparent = "transparent"
)

func Default(configError error) *Config {
	exitBackgroundTemplate := "{{ if gt .Code 0 }}p:red{{ end }}"
	exitTemplate := " {{ if gt .Code 0 }}\uf00d{{ else }}\uf00c{{ end }} "

	if configError != nil && configError != ErrNoConfig {
		exitBackgroundTemplate = "p:red"
		exitTemplate = configError.Error()
	}

	cfg := &Config{
		hash:       1234567890, // placeholder hash value
		Version:    4,
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
						Foreground:      paletteBlack,
						Background:      paletteYellow,
						Template:        " {{ if .SSHSession }}\ueba9 {{ end }}{{ .UserName }} ",
					},
					{
						Type:            PATH,
						Style:           Powerline,
						PowerlineSymbol: "\ue0b0",
						Foreground:      paletteWhite,
						Background:      paletteOrange,
						Options: options.Map{
							options.Style: "folder",
						},
						Template: " \uea83 {{ path .Path .Location }} ",
					},
					{
						Type:            GIT,
						Style:           Powerline,
						PowerlineSymbol: "\ue0b0",
						Foreground:      paletteBlack,
						Background:      paletteGreen,
						BackgroundTemplates: []string{
							"{{ if or (.Working.Changed) (.Staging.Changed) }}p:yellow{{ end }}",
							"{{ if and (gt .Ahead 0) (gt .Behind 0) }}p:red{{ end }}",
							"{{ if gt .Ahead 0 }}#49416D{{ end }}",
							"{{ if gt .Behind 0 }}#7A306C{{ end }}",
						},
						ForegroundTemplates: []string{
							"{{ if or (.Working.Changed) (.Staging.Changed) }}p:black{{ end }}",
							"{{ if and (gt .Ahead 0) (gt .Behind 0) }}p:white{{ end }}",
							"{{ if gt .Ahead 0 }}p:white{{ end }}",
						},
						Options: options.Map{
							segments.BranchTemplate:    "{{ trunc 25 .Branch }}",
							segments.FetchStatus:       true,
							segments.FetchUpstreamIcon: true,
						},
						Template: " {{ if .UpstreamURL }}{{ url .UpstreamIcon .UpstreamURL }} {{ end }}{{ .HEAD }}{{if .BranchStatus }} {{ .BranchStatus }}{{ end }}{{ if .Working.Changed }} \uf044 {{ .Working.String }}{{ end }}{{ if .Staging.Changed }} \uf046 {{ .Staging.String }}{{ end }} ", //nolint:lll
					},
					{
						Type:            ROOT,
						Style:           Powerline,
						PowerlineSymbol: "\ue0b0",
						Foreground:      paletteWhite,
						Background:      paletteYellow,
						Template:        " \uf0e7 ",
					},
					{
						Type:            STATUS,
						Style:           Diamond,
						LeadingDiamond:  "<transparent,background>\ue0b0</>",
						TrailingDiamond: "\ue0b4",
						Foreground:      paletteWhite,
						Background:      paletteBlue,
						BackgroundTemplates: []string{
							exitBackgroundTemplate,
						},
						Options: options.Map{
							options.AlwaysEnabled: true,
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
						Foreground: paletteGreen,
						Background: backgroundTransparent,
						Template:   "\ue718 ",
						Options: options.Map{
							segments.HomeEnabled:         false,
							segments.FetchPackageManager: false,
							segments.DisplayMode:         "files",
						},
					},
					{
						Type:       GOLANG,
						Style:      Plain,
						Foreground: paletteBlue,
						Background: backgroundTransparent,
						Template:   "\ue626 ",
						Options: options.Map{
							options.FetchVersion: false,
						},
					},
					{
						Type:       PYTHON,
						Style:      Plain,
						Foreground: paletteYellow,
						Background: backgroundTransparent,
						Template:   "\ue235 ",
						Options: options.Map{
							options.FetchVersion:     false,
							segments.DisplayMode:     "files",
							segments.FetchVirtualEnv: false,
						},
					},
					{
						Type:       SHELL,
						Style:      Plain,
						Foreground: paletteWhite,
						Background: backgroundTransparent,
						Template:   "in <p:blue><b>{{ .Name }}</b></> ",
					},
					{
						Type:       TIME,
						Style:      Plain,
						Foreground: paletteWhite,
						Background: backgroundTransparent,
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
			Foreground: paletteBlack,
			Background: backgroundTransparent,
			Template:   "<p:yellow,transparent>\ue0b6</><,p:yellow> > </><p:yellow,transparent>\ue0b0</> ",
		},
		TransientPrompt: &Segment{
			Foreground: paletteBlack,
			Background: backgroundTransparent,
			Template:   "<p:yellow,transparent>\ue0b6</><,p:yellow> {{ .Folder }} </><p:yellow,transparent>\ue0b0</> ",
		},
		Tooltips: []*Segment{
			{
				Type:            AWS,
				Style:           Diamond,
				LeadingDiamond:  "\ue0b0",
				TrailingDiamond: "\ue0b4",
				Foreground:      paletteWhite,
				Background:      paletteOrange,
				Template:        " \ue7ad {{ .Profile }}{{ if .Region }}@{{ .Region }}{{ end }} ",
				Options: options.Map{
					options.DisplayDefault: true,
				},
				Tips: []string{"aws"},
			},
			{
				Type:            AZ,
				Style:           Diamond,
				LeadingDiamond:  "\ue0b0",
				TrailingDiamond: "\ue0b4",
				Foreground:      paletteWhite,
				Background:      paletteBlue,
				Template:        " \uebd8 {{ .Name }} ",
				Options: options.Map{
					options.DisplayDefault: true,
				},
				Tips: []string{"az"},
			},
		},
		Upgrade: &upgrade.Config{
			Source:   upgrade.CDN,
			Interval: cache.ONEWEEK,
		},
	}

	return cfg
}

func Claude() *Config {
	cfg := &Config{
		hash:    1234567890, // placeholder hash value
		Version: 4,
		Blocks: []*Block{
			{
				Type:      Prompt,
				Alignment: Left,
				Segments: []*Segment{
					{
						Type:           PATH,
						Style:          Diamond,
						LeadingDiamond: "\ue0b6",
						Foreground:     paletteWhite,
						Background:     paletteOrange,
						Options: options.Map{
							segments.DirLength:           3,
							segments.FolderSeparatorIcon: "\ue0bb",
							options.Style:                "fish",
						},
						Template: "{{ if .Segments.Git.Dir }} \uf1d2 <i><b>{{ .Segments.Git.RepoName }}{{ if .Segments.Git.IsWorkTree }} \ue21c{{ end }}</b></i>{{ $rel :=  .Segments.Git.RelativeDir }}{{ if $rel }} \ueaf7 {{ .Format $rel }}{{ end }}{{ else }} \uea83 {{ path .Path .Location }}{{ end }} ", //nolint:lll
					},
					{
						Type:            GIT,
						Style:           Diamond,
						LeadingDiamond:  "<parentBackground,background>\ue0b0</>",
						TrailingDiamond: "\ue0b4",
						Foreground:      paletteBlack,
						Background:      paletteGreen,
						BackgroundTemplates: []string{
							"{{ if or (.Working.Changed) (.Staging.Changed) }}p:yellow{{ end }}",
							"{{ if and (gt .Ahead 0) (gt .Behind 0) }}p:red{{ end }}",
							"{{ if gt .Ahead 0 }}#49416D{{ end }}",
							"{{ if gt .Behind 0 }}#7A306C{{ end }}",
						},
						ForegroundTemplates: []string{
							"{{ if or (.Working.Changed) (.Staging.Changed) }}p:black{{ end }}",
							"{{ if or (gt .Ahead 0) (gt .Behind 0) }}p:white{{ end }}",
						},
						Options: options.Map{
							segments.FetchStatus:       true,
							segments.FetchUpstreamIcon: false,
						},
						Template: " {{ if .UpstreamURL }}{{ url .UpstreamIcon .UpstreamURL }} {{ end }}{{ .HEAD }}{{if .BranchStatus }} {{ .BranchStatus }}{{ end }}{{ if .Working.Changed }} \uf044 {{ nospace .Working.String }}{{ end }}{{ if .Staging.Changed }} \uf046 {{ .Staging.String }}{{ end }} ", //nolint:lll
					},
				},
			},
			{
				Type:      Prompt,
				Alignment: Right,
				Segments: []*Segment{
					{
						Type:            CLAUDE,
						Style:           Diamond,
						LeadingDiamond:  "\ue0b6",
						TrailingDiamond: "\ue0b4",
						Foreground:      paletteBlack,
						Background:      paletteBlue,
						Template:        " \U000f0bc9 {{ .Model.DisplayName }} \uf2d0 {{ .TokenUsagePercent.Gauge }} ",
					},
				},
			},
		},
		Palette: color.Palette{
			"black":  "#262B44",
			"blue":   "#4B95E9",
			"green":  "#59C9A5",
			"orange": "#F07623",
			"red":    "#D81E5B",
			"white":  "#E0DEF4",
			"yellow": "#F3AE35",
		},
	}

	return cfg
}
