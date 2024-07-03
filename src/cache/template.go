package cache

import (
	"sync"

	"github.com/jandedobbeleer/oh-my-posh/src/maps"
)

type Template struct {
	Root          bool
	PWD           string
	AbsolutePWD   string
	Folder        string
	Shell         string
	ShellVersion  string
	UserName      string
	HostName      string
	Code          int
	Env           map[string]string
	Var           maps.Simple
	OS            string
	WSL           bool
	PromptCount   int
	SHLVL         int
	Segments      *maps.Concurrent
	SegmentsCache maps.Simple

	Initialized bool
	sync.RWMutex
}

func (t *Template) AddSegmentData(key string, value any) {
	t.Segments.Set(key, value)
}

func (t *Template) RemoveSegmentData(key string) {
	t.Segments.Delete(key)
}
