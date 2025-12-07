package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

// SegmentBase provides base functionality for typed segments
type SegmentBase struct {
	// Runtime-only fields (not serialized)
	env   runtime.Environment `json:"-" toml:"-" yaml:"-"`
	text  string              `json:"-" toml:"-" yaml:"-"`
	index int                 `json:"-" toml:"-" yaml:"-"`
}

// SetText sets the rendered text
func (s *SegmentBase) SetText(text string) {
	s.text = text
}

// Text returns the rendered text
func (s *SegmentBase) Text() string {
	return s.text
}

// SetIndex sets the segment index
func (s *SegmentBase) SetIndex(index int) {
	s.index = index
}

// Init initializes the segment with the runtime environment
func (s *SegmentBase) Init(env runtime.Environment) {
	s.env = env
}

// Env returns the runtime environment
func (s *SegmentBase) Env() runtime.Environment {
	return s.env
}

// CacheKey returns an empty cache key by default
func (s *SegmentBase) CacheKey() (string, bool) {
	return "", false
}

// TypedSegmentMarker is a marker interface to identify typed segments
// Typed segments should implement this to distinguish from legacy property-based segments
type TypedSegmentMarker interface {
	IsTypedSegment()
}
