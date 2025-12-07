package config

import (
	"encoding/json"
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/segments"
)

// BlockType type of block
type BlockType string

// BlockAlignment aligment of a Block
type BlockAlignment string

// Overflow defines how to handle a right block that overflows with the previous block
type Overflow string

const (
	// Prompt writes one or more Segments
	Prompt BlockType = "prompt"
	// RPrompt is a right aligned prompt
	RPrompt BlockType = "rprompt"
	// Left aligns left
	Left BlockAlignment = "left"
	// Right aligns right
	Right BlockAlignment = "right"
	// Break adds a line break
	Break Overflow = "break"
	// Hide hides the block
	Hide Overflow = "hide"
)

// Block defines a part of the prompt with optional segments
type Block struct {
	Type            BlockType      `json:"type,omitempty" toml:"type,omitempty" yaml:"type,omitempty"`
	Alignment       BlockAlignment `json:"alignment,omitempty" toml:"alignment,omitempty" yaml:"alignment,omitempty"`
	Filler          string         `json:"filler,omitempty" toml:"filler,omitempty" yaml:"filler,omitempty"`
	Overflow        Overflow       `json:"overflow,omitempty" toml:"overflow,omitempty" yaml:"overflow,omitempty"`
	LeadingDiamond  string         `json:"leading_diamond,omitempty" toml:"leading_diamond,omitempty" yaml:"leading_diamond,omitempty"`
	TrailingDiamond string         `json:"trailing_diamond,omitempty" toml:"trailing_diamond,omitempty" yaml:"trailing_diamond,omitempty"`
	Segments        []*Segment     `json:"segments,omitempty" toml:"segments,omitempty" yaml:"segments,omitempty"`
	MaxWidth        int            `json:"max_width,omitempty" toml:"max_width,omitempty" yaml:"max_width,omitempty"`
	MinWidth        int            `json:"min_width,omitempty" toml:"min_width,omitempty" yaml:"min_width,omitempty"`
	Newline         bool           `json:"newline,omitempty" toml:"newline,omitempty" yaml:"newline,omitempty"`
	Force           bool           `json:"force,omitempty" toml:"force,omitempty" yaml:"force,omitempty"`
	Index           int            `json:"index,omitempty" toml:"index,omitempty" yaml:"index,omitempty"`
}

func (b *Block) key() any {
	if b.Index > 0 {
		return b.Index - 1
	}

	return fmt.Sprintf("%s-%s", b.Type, b.Alignment)
}

// UnmarshalJSON implements custom unmarshaling to support polymorphic segments
func (b *Block) UnmarshalJSON(data []byte) error {
	// Use type alias to avoid recursion
	type Alias Block
	aux := &struct {
		RawSegments []json.RawMessage `json:"segments"`
		*Alias
	}{
		Alias: (*Alias)(b),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Clear segments before repopulating
	b.Segments = nil

	for i, rawSeg := range aux.RawSegments {
		segment, err := unmarshalSegment(rawSeg)
		if err != nil {
			return fmt.Errorf("segment %d: %w", i, err)
		}
		b.Segments = append(b.Segments, segment)
	}

	return nil
}

func unmarshalSegment(data []byte) (*Segment, error) {
	// Peek at type field
	var typeCheck struct {
		Type SegmentType `json:"type"`
	}
	if err := json.Unmarshal(data, &typeCheck); err != nil {
		return nil, err
	}

	// Try to create a segment writer using the factory
	f, ok := Segments[typeCheck.Type]
	if !ok {
		return nil, fmt.Errorf("unknown segment type: %s", typeCheck.Type)
	}

	writer := f()

	if typedSeg, isTyped := writer.(segments.TypedSegmentMarker); isTyped {
		// Unmarshal into the typed segment
		if err := json.Unmarshal(data, writer); err != nil {
			return nil, err
		}

		// Apply defaults
		if err := ApplyDefaults(typedSeg); err != nil {
			return nil, err
		}

		// Create segment wrapper with writer already set
		seg := &Segment{
			Type:   typeCheck.Type,
			writer: writer,
		}

		return seg, nil
	}

	// Fall back to old property-based system for non-migrated segments
	var seg Segment
	if err := json.Unmarshal(data, &seg); err != nil {
		return nil, err
	}

	return &seg, nil
}
