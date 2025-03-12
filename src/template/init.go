package template

import (
	"sync"

	"github.com/jandedobbeleer/oh-my-posh/src/maps"
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
	shell       string
	env         runtime.Environment
	knownFields sync.Map
)

func Init(environment runtime.Environment, vars maps.Simple) {
	env = environment
	shell = env.Shell()
	knownFields = sync.Map{}

	renderPool = sync.Pool{
		New: func() any {
			return newTextPoolObject()
		},
	}

	if Cache != nil {
		return
	}

	loadCache(vars)
}
