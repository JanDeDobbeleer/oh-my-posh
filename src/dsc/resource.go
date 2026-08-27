package dsc

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

type Resource[T State[T]] struct {
	// SchemaJSON is the resource's JSON schema, generated ahead of time and
	// embedded at the construction site. Reflecting it at runtime would link
	// invopop/jsonschema (and go/ast, go/parser, go/doc) into the binary for
	// output that never changes between builds; a golden test per resource
	// regenerates the schema and fails on drift.
	SchemaJSON string `json:"-"`

	States []T `json:"states,omitempty" jsonschema:"title=states,description=The different states of the resource"`
}

type State[T any] interface {
	Equal(state T) bool
	Apply() error
	Resolve() (T, bool)
}

func (resource *Resource[T]) Load() {
	states, ok := cache.Device.Get[[]T](resource.cacheKey())
	if !ok {
		log.Debug("no states found in cache")
		return
	}

	resource.States = states
}

func (resource *Resource[T]) Save() {
	cache.Device.Set(resource.cacheKey(), resource.States, cache.INFINITE)
}

func (resource *Resource[T]) Add(item T) {
	for _, existingItem := range resource.States {
		if existingItem.Equal(item) {
			log.Debug("item already exists")
			return
		}
	}

	log.Debug("adding item")

	resource.States = append(resource.States, item)
}

func (resource *Resource[T]) Resolve() {
	for _, item := range resource.States {
		if resolvedItem, ok := item.Resolve(); ok {
			resource.States = append(resource.States, resolvedItem)
		}
	}
}

func (resource *Resource[T]) Apply(schema string) error {
	log.Debug("applying items")

	err := json.Unmarshal([]byte(schema), resource)
	if err != nil {
		return newError(err.Error())
	}

	// TODO: validate if we need to filter out States
	// which are already available in the cache (and thus set)

	for _, item := range resource.States {
		if applyErr := item.Apply(); applyErr != nil {
			log.Error(applyErr)
			err = errors.Join(err, applyErr)
		}
	}

	log.Debug("items applied")

	resource.Save()

	if err != nil {
		return newError(err.Error())
	}

	return nil
}

func (resource *Resource[T]) Test(_ string) error {
	return newError("test functionality not implemented")
}

func (resource *Resource[T]) Schema() string {
	return resource.SchemaJSON
}

func (resource *Resource[T]) getItemTypeName() string {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Pointer {
		return strings.ToLower(t.Elem().Name())
	}

	return strings.ToLower(t.Name())
}

func (resource *Resource[T]) cacheKey() string {
	return "DSC_" + strings.ToUpper(resource.getItemTypeName())
}

func (resource *Resource[T]) ToJSON() string {
	var result bytes.Buffer
	jsonEncoder := json.NewEncoder(&result)
	jsonEncoder.SetEscapeHTML(false)
	_ = jsonEncoder.Encode(resource)
	return result.String()
}
