package properties

import (
	"encoding/gob"
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

func init() {
	gob.Register([]any{})
	gob.Register(map[string]any{})
	gob.Register(map[any]any{})
	gob.Register([]string{})
	gob.Register(map[string]string{})
	gob.Register([]int{})
	gob.Register([]float64{})
	gob.Register([]bool{})
	gob.Register(int64(0))
	gob.Register(uint64(0))
	gob.Register(float32(0))
	gob.Register(Map{})
	gob.Register((*Property)(nil))
	gob.Register(map[Property]any{})
}

type Wrapper struct {
	Properties Map
}

func (w *Wrapper) GetColor(property Property, defaultColor color.Ansi) color.Ansi {
	value := w.Properties.GetColor(property, defaultColor)
	log.Debug(fmt.Sprintf("%s: %s", property, value))
	return value
}

func (w *Wrapper) GetBool(property Property, defaultValue bool) bool {
	value := w.Properties.GetBool(property, defaultValue)
	log.Debug(fmt.Sprintf("%s: %t", property, value))
	return value
}

func (w *Wrapper) GetString(property Property, defaultValue string) string {
	value := w.Properties.GetString(property, defaultValue)
	log.Debug(fmt.Sprintf("%s: %s", property, value))
	return value
}

func (w *Wrapper) GetFloat64(property Property, defaultValue float64) float64 {
	value := w.Properties.GetFloat64(property, defaultValue)
	log.Debug(fmt.Sprintf("%s: %f", property, value))
	return value
}

func (w *Wrapper) GetInt(property Property, defaultValue int) int {
	value := w.Properties.GetInt(property, defaultValue)
	log.Debug(fmt.Sprintf("%s: %d", property, value))
	return value
}

func (w *Wrapper) GetKeyValueMap(property Property, defaultValue map[string]string) map[string]string {
	value := w.Properties.GetKeyValueMap(property, defaultValue)
	log.Debug(fmt.Sprintf("%s: %v", property, value))
	return value
}

func (w *Wrapper) GetStringArray(property Property, defaultValue []string) []string {
	value := w.Properties.GetStringArray(property, defaultValue)
	log.Debug(fmt.Sprintf("%s: %v", property, value))
	return value
}

func (w *Wrapper) Get(property Property, defaultValue any) any {
	value := w.Properties.Get(property, defaultValue)
	log.Debug(fmt.Sprintf("%s: %v", property, value))
	return value
}
