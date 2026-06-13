package config

import (
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
)

// TestExportDoesNotLeakRuntimeContext asserts that the __template_context__
// sentinel injected by MapSegmentWithWriter is never present in any export
// format (JSON, YAML, TOML, gob) even when SetContext has run before export.
//
// This locks in the property that runtime state stays runtime-only.
func TestExportDoesNotLeakRuntimeContext(t *testing.T) {
	env := new(mock.Environment)
	env.On("TerminalWidth").Return(120, nil)
	env.On("Shell").Return(shell.BASH)

	template.Cache = new(cache.Template)
	template.Init(env, nil, nil)

	segment := &Segment{
		Type: PATH,
		Options: options.Map{
			"max_width": "{{ sub .TerminalWidth 30 }}",
		},
	}

	assert.NoError(t, segment.MapSegmentWithWriter(env), "MapSegmentWithWriter should succeed")
	// Sanity: the sentinel is in the in-memory map after rendering setup.
	assert.NotNil(t, segment.Options[options.TemplateContextKey],
		"expected sentinel injected by SetContext")

	cfg := &Config{
		Blocks: []*Block{{Segments: []*Segment{segment}}},
	}

	formats := []string{JSON, YAML, TOML}
	for _, format := range formats {
		out := cfg.Export(format)
		assert.NotEmpty(t, out, "%s export must not be empty", format)
		assert.False(t, strings.Contains(out, string(options.TemplateContextKey)),
			"%s export leaked sentinel: %s", format, out)
	}

	gobOut := cfg.Base64()
	assert.NotEmpty(t, gobOut, "gob/Base64 must not be empty")
	assert.False(t, strings.Contains(gobOut, string(options.TemplateContextKey)),
		"gob/Base64 leaked sentinel: %s", gobOut)

	// Also assert post-export: the in-memory map was cleaned, so subsequent
	// debug logging or third-party iteration cannot surface it.
	assert.Nil(t, segment.Options[options.TemplateContextKey],
		"expected sentinel cleared after export")
}
