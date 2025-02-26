package cache

import (
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
)

type Template struct {
	SegmentsCache maps.Simple
	Segments      *maps.Concurrent
	Var           maps.Simple
	PWD           string
	Folder        string
	PSWD          string
	UserName      string
	HostName      string
	ShellVersion  string
	Shell         string
	AbsolutePWD   string
	OS            string
	Version       string
	PromptCount   int
	SHLVL         int
	Jobs          int
	Code          int
	WSL           bool
	Root          bool
}

func (t *Template) AddSegmentData(key string, value any) {
	t.Segments.Set(key, value)
}

func (t *Template) RemoveSegmentData(key string) {
	t.Segments.Delete(key)
}
