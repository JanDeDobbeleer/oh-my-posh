package config

// TmuxConfig holds the configuration for rendering tmux status bar sections.
type TmuxConfig struct {
	StatusLeft  TmuxStatusSection `json:"status_left"  yaml:"status_left"  toml:"status_left"`
	StatusRight TmuxStatusSection `json:"status_right" yaml:"status_right" toml:"status_right"`
}

// TmuxStatusSection holds a list of blocks to render for one tmux status section.
type TmuxStatusSection struct {
	Blocks []*Block `json:"blocks" yaml:"blocks" toml:"blocks"`
}
