package template

import (
	"sync"

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
	knownFields sync.Map // keyed by reflect.Type → fieldSet (immutable once stored)
	textPool    *generics.Pool[*Text]

	// parsedTemplates caches fully-parsed *template.Template values keyed by
	// a string that encodes both the raw (unpatched) template text and the
	// context type, so cache hits can skip patchTemplate entirely.
	// Execute on a cached template is concurrency-safe; we never re-Parse it.
	parsedTemplates sync.Map
)

func Init(environment runtime.Environment, vars maps.Simple[any], aliases *maps.Config) {
	env = environment
	shell = env.Shell()

	// Reset per-process caches so tests that call Init multiple times stay isolated.
	knownFields = sync.Map{}
	parsedTemplates = sync.Map{}

	renderPool = generics.NewPool(func() *renderer {
		return &renderer{
			context: &context{},
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
