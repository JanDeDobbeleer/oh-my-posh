package template

import (
	"bytes"
	"errors"
	"reflect"
	"sort"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

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
// It encodes the raw (unpatched) template text and, when context is non-nil,
// the reflect.Type of the context so that two different struct types whose
// patchTemplate output would differ (via hasField) get separate cache entries.
// For map[string]any contexts the key also includes the sorted exported key
// names, since patchTemplate output depends on which keys are present.
func templateCacheKey(rawText string, ctx any) string {
	if ctx == nil {
		return rawText
	}

	t := reflect.TypeOf(ctx)
	if t.Kind() == reflect.Map {
		if m, ok := ctx.(map[string]any); ok {
			keys := make([]string, 0, len(m))
			for k := range m {
				if r, _ := utf8.DecodeRuneInString(k); unicode.IsUpper(r) {
					keys = append(keys, k)
				}
			}
			sort.Strings(keys)
			return rawText + "\x00" + t.String() + "\x00" + strings.Join(keys, "\x01")
		}
	}

	return rawText + "\x00" + t.String()
}

// parsedTemplate returns a fully-parsed *template.Template for text.
// The first call for a given key parses and stores it; subsequent calls
// return the cached value. Concurrent first-renders of the same template
// may both parse, but LoadOrStore ensures only one result is shared.
func parsedTemplate(text *Text) (*template.Template, error) {
	// Key on the raw, unpatched template text plus the context type so that
	// cache hits can skip patchTemplate entirely: patching is only needed
	// the first time a given (raw template, context type) pair is seen.
	key := templateCacheKey(text.template, text.context)

	if cached, ok := parsedTemplates.Load(key); ok {
		return cached.(*template.Template), nil
	}

	// Cache miss: patch the raw template into its executable form.
	text.patchTemplate()

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
