package template

import (
	"bytes"
	"errors"
	"strings"
	"sync"
	"text/template"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

type Data any

type context struct {
	Data
	Getenv func(string) string
	cache.Template
}

func (c *context) init(t *Text) {
	c.Data = t.Context
	c.Getenv = env.Getenv
	c.Template = *Cache
}

var renderPool sync.Pool

type renderer struct {
	template *template.Template
	context  *context
	buffer   bytes.Buffer
}

func newTextPoolObject() *renderer {
	return &renderer{
		template: template.New("cache").Funcs(funcMap()),
		context:  &context{},
	}
}

func (t *renderer) release() {
	t.buffer.Reset()
	t.context.Data = nil
	t.template.New("cache")
	renderPool.Put(t)
}

func (t *renderer) execute(text *Text) (string, error) {
	tmpl, err := t.template.Parse(text.Template)
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
