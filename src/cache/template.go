package cache

import (
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
)

type Template struct {
	SegmentsCache maps.Simple
	Segments      *maps.Concurrent
	Var           maps.Simple
	ShellVersion  string
	AbsolutePWD   string
	PSWD          string
	UserName      string
	HostName      string
	PWD           string
	Shell         string
	Folder        string
	OS            string
	Code          int
	PromptCount   int
	SHLVL         int
	Jobs          int
	WSL           bool
	Root          bool
	Initialized   bool
}

func (t *Template) AddSegmentData(key string, value any) {
	t.Segments.Set(key, value)
}

func (t *Template) RemoveSegmentData(key string) {
	t.Segments.Delete(key)
}
