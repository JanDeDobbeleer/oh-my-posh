package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCodePoints(t *testing.T) {
	codepoints, err := getGlyphCodePoints()
	if connectionError, ok := err.(*ConnectionError); ok {
		t.Log(connectionError.Error())
		return
	}
	assert.Equal(t, 1939, len(codepoints))
}
