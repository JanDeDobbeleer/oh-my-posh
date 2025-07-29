package template

import (
	"sync"
	"text/template"

	"github.com/jandedobbeleer/oh-my-posh/src/generics"
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
	textPool    *generics.Pool[*Text]
)

func Init(environment runtime.Environment, vars maps.Simple[any], aliases *maps.Config) {
	env = environment
	shell = env.Shell()
	knownFields = sync.Map{}

	renderPool = generics.NewPool(func() *renderer {
		return &renderer{
			template: template.New("cache").Funcs(funcMap()),
			context:  &context{},
		}
	})

	textPool = generics.NewPool(func() *Text {
		return &Text{}
	})

	if Cache != nil {
		return
	}

	loadCache(vars, aliases)
}
