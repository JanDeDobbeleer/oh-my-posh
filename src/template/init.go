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

	if Cache != nil && !refreshCache {
		return
	}

	refreshCache = false
	loadCache(vars, aliases)
}

var refreshCache bool

// ResetCache marks the template cache stale so the next Init rebuilds it.
// One-shot commands never need this - the cache is per-process by design
// (and tests rely on injecting a canned Cache before Init). The serve daemon
// does: its process outlives many prompts, and a surviving Cache pins
// per-prompt context (PWD, Folder, Code, Jobs, ...) to the values of the
// first render. The rebuild is deferred to Init (not done here) so Cache is
// never nil or partially built under a template render from an abandoned
// cycle's segment goroutine.
func ResetCache() {
	refreshCache = true
}
