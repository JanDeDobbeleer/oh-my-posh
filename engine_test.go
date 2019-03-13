package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLenWithoutANSI(t *testing.T) {
	block := &Block{
		Type:      Prompt,
		Alignment: Right,
		Segments: []*Segment{
			{
				Type:       Time,
				Style:      Plain,
				Background: "#B8B80A",
				Foreground: "#ffffff",
			},
		},
	}
	engine := &engine{
		renderer: &ColorWriter{
			Buffer: new(bytes.Buffer),
		},
	}
	blockText := engine.renderBlockSegments(block)
	strippedLength := engine.lenWithoutANSI(blockText)
	assert.Equal(t, 10, strippedLength)
}
