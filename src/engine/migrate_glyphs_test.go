package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCodePoints(t *testing.T) {
	codepoints := getGlyphCodePoints()
	assert.Equal(t, 1939, len(codepoints))
}
