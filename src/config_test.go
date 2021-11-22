package main

import (
	"testing"

	"github.com/gookit/config/v2"
	"github.com/stretchr/testify/assert"
)

func TestSettingsExportJSON(t *testing.T) {
	defer testClearDefaultConfig()
	content := exportConfig("../themes/jandedobbeleer.omp.json", "json")
	assert.NotContains(t, content, "\\u003ctransparent\\u003e")
	assert.Contains(t, content, "<transparent>")
}

func testClearDefaultConfig() {
	config.Default().ClearAll()
}
