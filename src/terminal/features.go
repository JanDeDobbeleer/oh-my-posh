package terminal

import "encoding/gob"

func init() {
	gob.Register(&TerminalFeatures{})
}

type TerminalFeatures struct {
	Progress []string `json:"progress,omitempty" toml:"progress,omitempty" yaml:"progress,omitempty"`
}
