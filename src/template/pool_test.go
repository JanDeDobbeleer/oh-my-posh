package template

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

func TestTextPool(t *testing.T) {
	env := new(mock.Environment)
	env.On("Shell").Return("foo")
	Cache = new(cache.Template)
	Init(env, nil, nil)

	// Test that New() returns a Text instance
	text1 := New("Hello {{ .Name }}", map[string]any{"Name": "World"})
	assert.NotNil(t, text1)

	// Test rendering
	result, err := text1.Render()
	assert.NoError(t, err)
	assert.Equal(t, "Hello World", result)

	// Release back to pool
	text1.Release()

	// Verify fields are reset (we can't check them directly since they're unexported)
	// But we can test by creating a new instance and verifying it works
	text2 := New("", nil)
	assert.NotNil(t, text2)

	// Test empty template
	result2, err := text2.Render()
	assert.NoError(t, err)
	assert.Equal(t, "", result2)

	text2.Release()
}

func TestTextPoolFallback(t *testing.T) {
	// Test fallback when pool is not initialized
	originalPool := textPool
	textPool = nil

	text := New("test", nil)
	assert.NotNil(t, text)

	// Should work without panic
	text.Release()

	// Restore
	textPool = originalPool
}
