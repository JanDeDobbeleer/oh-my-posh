package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAsyncTimeoutFromConfig(t *testing.T) {
	jsonConfig := `{
		"version": 3,
		"blocks": [
			{
				"type": "prompt",
				"segments": [
					{
						"type": "git",
						"style": "plain",
						"async_timeout": 100,
						"properties": {
							"fetch_status": true
						}
					}
				]
			}
		]
	}`
	
	// Create a temporary file for the config
	tmpFile := "/tmp/test_config.json"
	err := os.WriteFile(tmpFile, []byte(jsonConfig), 0644)
	assert.NoError(t, err)
	defer os.Remove(tmpFile)
	
	// Load the config
	cfg, _ := Load(tmpFile, "generic", false)
	assert.NotNil(t, cfg)
	
	// Verify the async timeout is loaded correctly
	assert.Len(t, cfg.Blocks, 1)
	assert.Len(t, cfg.Blocks[0].Segments, 1)
	
	gitSegment := cfg.Blocks[0].Segments[0]
	assert.Equal(t, "git", string(gitSegment.Type))
	assert.Equal(t, 100*time.Nanosecond, gitSegment.AsyncTimeout)
}