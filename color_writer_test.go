package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteAndRemoveText(t *testing.T) {
	writer := &ColorWriter{
		Buffer: new(bytes.Buffer),
	}
	inputText := "This is white, <#ff5733>this is orange</>, white again"
	text := writer.writeAndRemoveText("#193549", "#fff", "This is white, ", "This is white, ", inputText)
	assert.Equal(t, "<#ff5733>this is orange</>, white again", text)
	assert.NotContains(t, writer.string(), "<#ff5733>")
}

func TestWriteAndRemoveTextColored(t *testing.T) {
	writer := &ColorWriter{
		Buffer: new(bytes.Buffer),
	}
	inputText := "This is white, <#ff5733>this is orange</>, white again"
	text := writer.writeAndRemoveText("#193549", "#ff5733", "this is orange", "<#ff5733>this is orange</>", inputText)
	assert.Equal(t, "This is white, , white again", text)
	assert.NotContains(t, writer.string(), "<#ff5733>")
}

func TestWriteColorOverride(t *testing.T) {
	writer := &ColorWriter{
		Buffer: new(bytes.Buffer),
	}
	text := "This is white, <#ff5733>this is orange</>, white again"
	writer.write("#193549", "#ff5733", text)
	assert.NotContains(t, writer.string(), "<#ff5733>")
}
