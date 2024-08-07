package template

import (
	"bytes"
	"sync"
	"text/template"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

const (
	// Errors to show when the template handling fails
	InvalidTemplate   = "invalid template text"
	IncorrectTemplate = "unable to create text based on template"

	globalRef = ".$"

	elvish = "elvish"
	xonsh  = "xonsh"
)

var (
	shell          string
	tmplFunc       *template.Template
	contextPool    sync.Pool
	buffPool       sync.Pool
	env            runtime.Environment
	knownVariables []string
)

type buff bytes.Buffer

func (b *buff) release() {
	(*bytes.Buffer)(b).Reset()
	buffPool.Put(b)
}

func (b *buff) Write(p []byte) (n int, err error) {
	return (*bytes.Buffer)(b).Write(p)
}

func (b *buff) String() string {
	return (*bytes.Buffer)(b).String()
}

func Init(environment runtime.Environment) {
	env = environment
	shell = env.Shell()

	tmplFunc = template.New("cache").Funcs(funcMap())

	contextPool = sync.Pool{
		New: func() any {
			return &context{}
		},
	}

	buffPool = sync.Pool{
		New: func() any {
			return &buff{}
		},
	}

	knownVariables = []string{
		"Root",
		"PWD",
		"AbsolutePWD",
		"Folder",
		"Shell",
		"ShellVersion",
		"UserName",
		"HostName",
		"Code",
		"Env",
		"OS",
		"WSL",
		"PromptCount",
		"Segments",
		"SHLVL",
		"Templates",
		"Var",
		"Data",
		"Jobs",
	}
}
