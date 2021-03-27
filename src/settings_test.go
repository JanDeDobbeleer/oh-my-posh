package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSettingsExportJSON(t *testing.T) {
	content := exportConfig("../themes/jandedobbeleer.omp.json", "json")
	assert.NotContains(t, content, "\\u003ctransparent\\u003e")
	assert.Contains(t, content, "<transparent>")
}
