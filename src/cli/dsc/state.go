package dsc

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
)

//go:embed schema.json
var Schema string

type State struct {
	Configurations Configurations `json:"configurations,omitempty"`
	Shells         Shells         `json:"shells,omitempty"`
	Fonts          Fonts          `json:"fonts,omitempty"`
}

func (s *State) empty() bool {
	return len(s.Configurations) == 0 && len(s.Shells) == 0 && len(s.Fonts) == 0
}

func (s *State) String() string {
	schemaJSON := `"$schema": "https://ohmyposh.dev/dsc.schema.json"`
	if s.empty() {
		return fmt.Sprintf("{%s}", schemaJSON)
	}

	var result bytes.Buffer
	jsonEncoder := json.NewEncoder(&result)
	jsonEncoder.SetEscapeHTML(false)
	jsonEncoder.SetIndent("", "  ")
	_ = jsonEncoder.Encode(s)
	prefix := fmt.Sprintf("{\n  %s,", schemaJSON)
	return strings.Replace(result.String(), "{", prefix, 1)
}

func (s *State) Apply(c cache.Cache) error {
	err := s.Configurations.Apply()

	shellErr := s.Shells.Apply()
	if shellErr != nil {
		err = errors.Join(err, shellErr)
	}

	fontErr := s.Fonts.Apply(c)
	if fontErr != nil {
		err = errors.Join(err, fontErr)
	}

	return err
}
