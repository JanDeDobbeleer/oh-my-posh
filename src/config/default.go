package config

import (
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

func Default(env platform.Environment, warning bool) *Config {
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
						Template:        " {{ if .SSHSession }}\ueba9 {{ end }}{{ .UserName }} ",
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
						Template: " \uea83 {{ path .Path .Location }} ",
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
						Type:            STATUS,
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
						Template:   "\ue718 ",
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
						Template:   "\ue626 ",
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
		Palette: terminal.Palette{
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
				Template:        " \uebd8 {{ .Name }} ",
				Properties: properties.Map{
					properties.DisplayDefault: true,
				},
				Tips: []string{"az"},
			},
		},
	}
	cfg.env = env
	return cfg
}
