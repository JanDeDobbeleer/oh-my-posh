package config

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
	Type            BlockType      `json:"type,omitempty" toml:"type,omitempty"`
	Alignment       BlockAlignment `json:"alignment,omitempty" toml:"alignment,omitempty"`
	Filler          string         `json:"filler,omitempty" toml:"filler,omitempty"`
	Overflow        Overflow       `json:"overflow,omitempty" toml:"overflow,omitempty"`
	LeadingDiamond  string         `json:"leading_diamond,omitempty" toml:"leading_diamond,omitempty"`
	TrailingDiamond string         `json:"trailing_diamond,omitempty" toml:"trailing_diamond,omitempty"`
	Segments        []*Segment     `json:"segments,omitempty" toml:"segments,omitempty"`
	MaxWidth        int            `json:"max_width,omitempty" toml:"max_width,omitempty"`
	MinWidth        int            `json:"min_width,omitempty" toml:"min_width,omitempty"`
	Newline         bool           `json:"newline,omitempty" toml:"newline,omitempty"`
	Force           bool           `json:"force,omitempty" toml:"force,omitempty"`
}
