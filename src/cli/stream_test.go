package cli

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStreamRequest_JSONMarshaling(t *testing.T) {
	req := StreamRequest{
		ID: "test-uuid-123",
		Flags: RequestFlags{
			ConfigPath:    "/path/to/config.json",
			Shell:         "nu",
			ShellVersion:  "0.80.0",
			PWD:           "/home/user",
			Status:        0,
			NoStatus:      false,
			ExecutionTime: 1.5,
			TerminalWidth: 120,
			JobCount:      2,
			Cleared:       false,
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(&req)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Unmarshal back
	var decoded StreamRequest
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, req.ID, decoded.ID)
	assert.Equal(t, req.Flags.Shell, decoded.Flags.Shell)
	assert.Equal(t, req.Flags.PWD, decoded.Flags.PWD)
}

func TestStreamResponse_JSONMarshaling(t *testing.T) {
	resp := StreamResponse{
		ID:   "test-uuid-123",
		Type: "complete",
		Prompts: map[string]string{
			"primary": "prompt text",
			"right":   "right prompt",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(&resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Unmarshal back
	var decoded StreamResponse
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, resp.ID, decoded.ID)
	assert.Equal(t, resp.Type, decoded.Type)
	assert.Equal(t, resp.Prompts["primary"], decoded.Prompts["primary"])
}

func TestStreamResponse_WithError(t *testing.T) {
	resp := StreamResponse{
		ID:      "test-uuid-123",
		Type:    "error",
		Error:   "test error message",
		Prompts: make(map[string]string),
	}

	// Marshal to JSON
	data, err := json.Marshal(&resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "test error message")

	// Unmarshal back
	var decoded StreamResponse
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "error", decoded.Type)
	assert.Equal(t, "test error message", decoded.Error)
}
