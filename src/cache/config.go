package cache

type Config struct {
	Duration Duration `json:"duration,omitempty" toml:"duration,omitempty"`
	Strategy Strategy `json:"strategy,omitempty" toml:"strategy,omitempty"`
}

type Strategy string

const (
	Folder  Strategy = "folder"
	Session Strategy = "session"
)
