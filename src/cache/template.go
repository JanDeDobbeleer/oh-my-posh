package cache

import (
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
)

type Template struct {
	Segments *maps.Concurrent[any]
	SimpleTemplate
}

type SimpleTemplate struct {
	SegmentsCache maps.Simple[any]
	Var           maps.Simple[any]
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
