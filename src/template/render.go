package template

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"text/template"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/generics"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

type Data any

type context struct {
	Data
	Getenv func(string) string
	cache.Template
}

func (c *context) init(t *Text) {
	c.Data = t.context
	c.Getenv = env.Getenv
	c.Template = *Cache
}

var (
	renderPool *generics.Pool[*renderer]
)

// renderer holds only per-render mutable state: an output buffer and a context.
// The parsed template is now looked up from the shared parsedTemplates cache.
type renderer struct {
	context *context
	buffer  bytes.Buffer
}

func (t *renderer) release() {
	t.buffer.Reset()
	t.context.Data = nil
	renderPool.Put(t)
}

// templateCacheKey returns the key used to look up a cached *template.Template.
// It encodes the patched template text and, when context is non-nil, the
// reflect.Type of the context so that two different struct types whose
// patchTemplate output differs (via hasField) get separate cache entries.
func templateCacheKey(patchedText string, ctx any) string {
	if ctx == nil {
		return patchedText
	}
	return patchedText + "\x00" + reflect.TypeOf(ctx).String()
}

// parsedTemplate returns a fully-parsed *template.Template for text.
// The first call for a given key parses and stores it; subsequent calls
// return the cached value. Concurrent first-renders of the same template
// may both parse, but LoadOrStore ensures only one result is shared.
func parsedTemplate(text *Text) (*template.Template, error) {
	// patchTemplate rewrites the raw template string into the patched form
	// and stores it back in text.template. We need the patched text as part
	// of the cache key, so we must patch first.
	text.patchTemplate()

	key := templateCacheKey(text.template, text.context)

	if cached, ok := parsedTemplates.Load(key); ok {
		return cached.(*template.Template), nil
	}

	// Parse into a fresh template with the shared func map and settings.
	tmpl, err := template.New("cache").Funcs(funcMap()).Parse(text.template)
	if err != nil {
		return nil, err
	}

	// Store; if another goroutine already stored an equivalent template, use theirs.
	actual, _ := parsedTemplates.LoadOrStore(key, tmpl)
	return actual.(*template.Template), nil
}

func (t *renderer) execute(text *Text) (string, error) {
	tmpl, err := parsedTemplate(text)
	if err != nil {
		log.Error(err)
		return "", errors.New(InvalidTemplate)
	}

	t.context.init(text)

	err = tmpl.Execute(&t.buffer, t.context)
	if err != nil {
		log.Error(err)
		return "", errors.New(IncorrectTemplate)
	}

	output := t.buffer.String()

	// issue with missingkey=zero ignored for map[string]any
	// https://github.com/golang/go/issues/24963
	output = strings.ReplaceAll(output, "<no value>", "")

	return output, nil
}
