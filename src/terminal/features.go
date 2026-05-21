package terminal

import "encoding/gob"

func init() {
	gob.Register(&Features{})
}

type Features struct {
	Progress []string `json:"progress,omitempty" toml:"progress,omitempty" yaml:"progress,omitempty"`
}

func (f *Features) Apply() {
	// an empty list can't override the default: the gob-encoded session cache
	// collapses empty slices to nil, so it can't be told apart from unset
	if f == nil || len(f.Progress) == 0 {
		return
	}

	progressTerminals = f.Progress
}
