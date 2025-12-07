package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockUnmarshalTypedSegment(t *testing.T) {
	jsonData := `{
		"type": "prompt",
		"segments": [
			{
				"type": "status",
				"status_template": "custom template",
				"always_enabled": true
			}
		]
	}`

	var block Block
	err := json.Unmarshal([]byte(jsonData), &block)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(block.Segments))
	assert.Equal(t, STATUS, block.Segments[0].Type)

	// The writer field should be set (it's private but accessible in same package)
	segment := block.Segments[0]
	assert.NotNil(t, segment.writer, "writer should be set")

	// Check it's marked as typed
	type typedMarker interface {
		IsTypedSegment()
	}
	_, isTyped := segment.writer.(typedMarker)
	assert.True(t, isTyped, "segment should be marked as typed")
}

func TestBlockUnmarshalLegacySegment(t *testing.T) {
	jsonData := `{
		"type": "prompt",
		"segments": [
			{
				"type": "text",
				"properties": {
					"text": "hello"
				}
			}
		]
	}`

	var block Block
	err := json.Unmarshal([]byte(jsonData), &block)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(block.Segments))
	assert.Equal(t, TEXT, block.Segments[0].Type)

	// For legacy segments, properties should be preserved
	assert.NotNil(t, block.Segments[0].Properties)
}
