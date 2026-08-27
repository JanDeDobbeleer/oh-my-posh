// Package ini is a minimal INI reader covering what the aws, gcp and git
// segments need: sections in file order, first key wins, values cut at the
// first # or ; (unless loaded verbatim) and surrounded quotes trimmed.
package ini

import (
	"errors"
	"strings"
)

type File struct {
	sections map[string]*Section
	order    []*Section
}

type Section struct {
	name  string
	keys  map[string]*Key
	order []*Key
}

type Key struct {
	name  string
	value string
}

// Load parses src, cutting values at the first # or ; by default.
func Load(src string) (*File, error) {
	return parse(src, false)
}

// LoadVerbatim parses src keeping values as-is, without cutting at inline
// comment markers.
func LoadVerbatim(src string) (*File, error) {
	return parse(src, true)
}

func parse(src string, verbatim bool) (*File, error) {
	file := &File{sections: make(map[string]*Section)}

	src = strings.TrimPrefix(src, "\ufeff")
	current := file.section("")

	for line := range strings.Lines(src) {
		line = strings.TrimSpace(line)

		if line == "" || line[0] == '#' || line[0] == ';' {
			continue
		}

		if line[0] == '[' {
			name, _, found := strings.CutLast(line, "]")
			if !found {
				return nil, errors.New("unclosed section: " + line)
			}

			current = file.section(strings.TrimSpace(name[1:]))
			continue
		}

		delim := strings.IndexAny(line, "=:")
		if delim < 0 {
			return nil, errors.New("key-value delimiter not found: " + line)
		}

		name := strings.TrimSpace(line[:delim])
		value := parseValue(line[delim+1:], verbatim)

		if _, ok := current.keys[name]; ok {
			continue
		}

		key := &Key{name: name, value: value}
		current.keys[name] = key
		current.order = append(current.order, key)
	}

	return file, nil
}

func parseValue(value string, verbatim bool) string {
	value = strings.TrimSpace(value)

	if !verbatim {
		if i := strings.IndexAny(value, "#;"); i > -1 {
			value = strings.TrimSpace(value[:i])
		}
	}

	if len(value) > 1 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
		value = value[1 : len(value)-1]
	}

	return value
}

func (f *File) section(name string) *Section {
	if section, ok := f.sections[name]; ok {
		return section
	}

	section := &Section{name: name, keys: make(map[string]*Key)}
	f.sections[name] = section
	f.order = append(f.order, section)

	return section
}

// Sections returns all sections in file order, including the unnamed default
// section.
func (f *File) Sections() []*Section {
	return f.order
}

// GetSection returns the named section, or an error when it doesn't exist.
func (f *File) GetSection(name string) (*Section, error) {
	if section, ok := f.sections[name]; ok {
		return section, nil
	}

	return nil, errors.New("section does not exist: " + name)
}

// Section returns the named section, or an empty one when it doesn't exist.
func (f *File) Section(name string) *Section {
	if section, ok := f.sections[name]; ok {
		return section
	}

	return &Section{name: name, keys: make(map[string]*Key)}
}

func (s *Section) Name() string {
	return s.name
}

// Keys returns the section's keys in file order.
func (s *Section) Keys() []*Key {
	return s.order
}

// Key returns the named key, or an empty one when it doesn't exist.
func (s *Section) Key(name string) *Key {
	if key, ok := s.keys[name]; ok {
		return key
	}

	return &Key{name: name}
}

func (k *Key) Name() string {
	return k.name
}

func (k *Key) Value() string {
	return k.value
}

func (k *Key) String() string {
	return k.value
}
