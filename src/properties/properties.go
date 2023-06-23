package properties

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
)

type Wrapper struct {
	Properties Map
	Env        platform.Environment
}

func (w *Wrapper) GetColor(property Property, defaultColor string) string {
	value := w.Properties.GetColor(property, defaultColor)
	w.Env.Debug(fmt.Sprintf("%s: %s", property, value))
	return value
}

func (w *Wrapper) GetBool(property Property, defaultValue bool) bool {
	value := w.Properties.GetBool(property, defaultValue)
	w.Env.Debug(fmt.Sprintf("%s: %t", property, value))
	return value
}

func (w *Wrapper) GetString(property Property, defaultValue string) string {
	value := w.Properties.GetString(property, defaultValue)
	w.Env.Debug(value)
	return value
}

func (w *Wrapper) GetFloat64(property Property, defaultValue float64) float64 {
	value := w.Properties.GetFloat64(property, defaultValue)
	w.Env.Debug(fmt.Sprintf("%s: %f", property, value))
	return value
}

func (w *Wrapper) GetInt(property Property, defaultValue int) int {
	value := w.Properties.GetInt(property, defaultValue)
	w.Env.Debug(fmt.Sprintf("%s: %d", property, value))
	return value
}

func (w *Wrapper) GetKeyValueMap(property Property, defaultValue map[string]string) map[string]string {
	value := w.Properties.GetKeyValueMap(property, defaultValue)
	w.Env.Debug(fmt.Sprintf("%s: %v", property, value))
	return value
}

func (w *Wrapper) GetStringArray(property Property, defaultValue []string) []string {
	value := w.Properties.GetStringArray(property, defaultValue)
	w.Env.Debug(fmt.Sprintf("%s: %v", property, value))
	return value
}
