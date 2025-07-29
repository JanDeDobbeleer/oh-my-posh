package template

import (
	"bytes"
	"errors"
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

var renderPool *generics.Pool[*renderer]

type renderer struct {
	template *template.Template
	context  *context
	buffer   bytes.Buffer
}

func (t *renderer) release() {
	t.buffer.Reset()
	t.context.Data = nil
	t.template.New("cache")
	renderPool.Put(t)
}

func (t *renderer) execute(text *Text) (string, error) {
	tmpl, err := t.template.Parse(text.template)
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
