package config

import (
	"sync"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
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
	// LineBreak creates a line break in the prompt
	LineBreak BlockType = "newline"
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
	Type      BlockType      `json:"type,omitempty" toml:"type,omitempty"`
	Alignment BlockAlignment `json:"alignment,omitempty" toml:"alignment,omitempty"`
	Segments  []*Segment     `json:"segments,omitempty" toml:"segments,omitempty"`
	Newline   bool           `json:"newline,omitempty" toml:"newline,omitempty"`
	Filler    string         `json:"filler,omitempty" toml:"filler,omitempty"`
	Overflow  Overflow       `json:"overflow,omitempty" toml:"overflow,omitempty"`

	LeadingDiamond  string `json:"leading_diamond,omitempty" toml:"leading_diamond,omitempty"`
	TrailingDiamond string `json:"trailing_diamond,omitempty" toml:"trailing_diamond,omitempty"`

	// Deprecated: keep the logic for legacy purposes
	HorizontalOffset int `json:"horizontal_offset,omitempty" toml:"horizontal_offset,omitempty"`
	VerticalOffset   int `json:"vertical_offset,omitempty" toml:"vertical_offset,omitempty"`

	MaxWidth int `json:"max_width,omitempty" toml:"max_width,omitempty"`
	MinWidth int `json:"min_width,omitempty" toml:"min_width,omitempty"`

	env runtime.Environment
}

func (b *Block) Init(env runtime.Environment) {
	b.env = env
	b.executeSegmentLogic()
}

func (b *Block) Enabled() bool {
	if b.Type == LineBreak {
		return true
	}

	for _, segment := range b.Segments {
		if segment.Enabled {
			return true
		}
	}

	return false
}

func (b *Block) executeSegmentLogic() {
	if shouldHideForWidth(b.env, b.MinWidth, b.MaxWidth) {
		return
	}

	b.setEnabledSegments()
}

func (b *Block) setEnabledSegments() {
	wg := sync.WaitGroup{}
	wg.Add(len(b.Segments))
	defer wg.Wait()

	for _, segment := range b.Segments {
		go func(s *Segment) {
			defer wg.Done()
			s.SetEnabled(b.env)
		}(segment)
	}
}
