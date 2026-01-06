package config

import "github.com/jandedobbeleer/oh-my-posh/src/cache"

type Cache struct {
	Duration cache.Duration `json:"duration,omitempty" toml:"duration,omitempty" yaml:"duration,omitempty"`
	Strategy Strategy       `json:"strategy,omitempty" toml:"strategy,omitempty" yaml:"strategy,omitempty"`
}

type Strategy string

const (
	Folder  Strategy = "folder"
	Session Strategy = "session"
	Device  Strategy = "device"
)
