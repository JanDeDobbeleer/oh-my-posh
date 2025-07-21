package dsc

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

type Resource[T State[T]] struct {
	Cache         cache.Cache `json:"-"`
	JSONSchemaURL string      `json:"$schema,omitempty"`
	States        []T         `json:"states,omitempty" jsonschema:"title=states,description=The different states of the resource"`
}

type State[T any] interface {
	Equal(state T) bool
	Apply() error
	Resolve() (T, bool)
}

func (resource *Resource[T]) SetCache(c cache.Cache) {
	resource.Cache = c
}

func (resource *Resource[T]) Load(c cache.Cache) {
	resource.Cache = c

	cached, ok := resource.Cache.Get(resource.cacheKey())
	if !ok {
		return
	}

	var items []T
	if err := json.Unmarshal([]byte(cached), &items); err != nil {
		log.Error(err)
		return
	}

	resource.States = items
}

func (resource *Resource[T]) Save() {
	data, err := json.Marshal(resource.States)
	if err != nil {
		log.Error(err)
		return
	}

	resource.Cache.Set(resource.cacheKey(), string(data), cache.INFINITE)
}

func (resource *Resource[T]) Add(item T) {
	for _, existingItem := range resource.States {
		if existingItem.Equal(item) {
			log.Debug("Item already exists")
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
	reflector := jsonschema.Reflector{
		ExpandedStruct: true,
		DoNotReference: true,
	}

	schema := reflector.Reflect(resource)
	schema.ID = jsonschema.ID(resource.getItemTypeName())
	schema.Properties.Delete("$schema")
	schemaJSON, _ := json.MarshalIndent(schema, "", "  ")

	return string(schemaJSON)
}

func (resource *Resource[T]) getItemTypeName() string {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
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
	jsonEncoder.SetIndent("", "  ")
	_ = jsonEncoder.Encode(resource)
	return result.String()
}
