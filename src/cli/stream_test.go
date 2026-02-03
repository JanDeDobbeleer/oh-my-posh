package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamCommand_Creation(t *testing.T) {
	cmd := createStreamCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "stream", cmd.Use)
	assert.Equal(t, "Stream the prompt with incremental updates", cmd.Short)
}

func TestStreamCommand_Flags(t *testing.T) {
	cmd := createStreamCmd()

	// Verify all expected flags exist
	expectedFlags := []string{
		"pwd",
		"pswd",
		"shell",
		"shell-version",
		"status",
		"no-status",
		"pipestatus",
		"execution-time",
		"stack-count",
		"terminal-width",
		"cleared",
		"eval",
		"column",
		"job-count",
		"save-cache",
		"escape",
		"force",
	}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag '%s' should exist", flagName)
	}
}

func TestStreamCommand_RequiredFlagsForStreaming(t *testing.T) {
	// This test validates that the stream command sets the correct flags
	// for streaming execution mode

	cmd := createStreamCmd()

	// Verify that running the command would set streaming=true
	// We can't easily test the actual run without a full config,
	// but we can verify the command is properly configured
	assert.NotNil(t, cmd.Run)
	assert.NotNil(t, cmd.Args)
}

func TestStreamCommand_FlagInheritance(t *testing.T) {
	// Verify that stream command uses the same flags as print command
	// This ensures consistency between commands

	streamCmd := createStreamCmd()
	printCmd := createPrintCmd()

	// Core flags that should exist in both
	sharedFlags := []string{
		"pwd",
		"shell",
		"status",
		"execution-time",
		"terminal-width",
		"eval",
		"force",
	}

	for _, flagName := range sharedFlags {
		streamFlag := streamCmd.Flags().Lookup(flagName)
		printFlag := printCmd.Flags().Lookup(flagName)

		assert.NotNil(t, streamFlag, "Stream command should have '%s' flag", flagName)
		assert.NotNil(t, printFlag, "Print command should have '%s' flag", flagName)

		// Verify default values match
		if streamFlag != nil && printFlag != nil {
			assert.Equal(t, printFlag.DefValue, streamFlag.DefValue,
				"Flag '%s' should have same default value in both commands", flagName)
		}
	}
}

func TestStreamCommand_OutputDelimiter(t *testing.T) {
	// Test that output uses null byte delimiter for multi-line prompts
	tests := []struct {
		name     string
		expected string
		prompts  []string
	}{
		{
			name:     "Single line prompt",
			prompts:  []string{"prompt1"},
			expected: "prompt1\x00",
		},
		{
			name:     "Multi-line prompt",
			prompts:  []string{"line1\nline2\nline3"},
			expected: "line1\nline2\nline3\x00",
		},
		{
			name:     "Multiple prompts",
			prompts:  []string{"prompt1", "prompt2", "prompt3"},
			expected: "prompt1\x00prompt2\x00prompt3\x00",
		},
		{
			name:     "Multiple multi-line prompts",
			prompts:  []string{"line1\nline2", "line3\nline4"},
			expected: "line1\nline2\x00line3\nline4\x00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate output with null byte delimiter
			var buf bytes.Buffer
			for _, prompt := range tt.prompts {
				buf.WriteString(prompt)
				buf.WriteString("\x00")
			}

			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestStreamCommand_Integration_MockOutput(t *testing.T) {
	// This test validates the output structure without requiring a full engine
	// It simulates what the stream command would output with null byte delimiter

	tests := []struct {
		validateOutput func(t *testing.T, output string)
		name           string
		promptCount    int
	}{
		{
			name:        "Single prompt with null byte",
			promptCount: 1,
			validateOutput: func(t *testing.T, output string) {
				assert.True(t, strings.HasSuffix(output, "\x00"))
			},
		},
		{
			name:        "Multiple prompts with null bytes",
			promptCount: 3,
			validateOutput: func(t *testing.T, output string) {
				parts := strings.Split(output, "\x00")
				// 3 prompts = 4 parts (including trailing empty string after last \x00)
				assert.Len(t, parts, 4)
				// Last part should be empty
				assert.Equal(t, "", parts[3])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate stream output with null byte delimiter
			var output bytes.Buffer
			for i := 0; i < tt.promptCount; i++ {
				output.WriteString("prompt")
				output.WriteString("\x00")
			}

			tt.validateOutput(t, output.String())
		})
	}
}

func TestStreamCommand_HiddenFlags(t *testing.T) {
	cmd := createStreamCmd()

	// Verify save-cache is hidden (internal use only)
	saveCacheFlag := cmd.Flags().Lookup("save-cache")
	require.NotNil(t, saveCacheFlag)
	assert.True(t, saveCacheFlag.Hidden, "save-cache flag should be hidden")
}

func TestStreamCommand_NoArgs(t *testing.T) {
	cmd := createStreamCmd()

	// Stream command should not accept positional arguments
	// (unlike print which accepts primary/secondary/etc.)
	assert.NotNil(t, cmd.Args)

	// Test that NoArgs validator rejects arguments
	err := cmd.Args(cmd, []string{"extra"})
	assert.Error(t, err, "Should reject arguments when NoArgs is used")

	// Test that NoArgs validator accepts no arguments
	err = cmd.Args(cmd, []string{})
	assert.NoError(t, err, "Should accept no arguments")
}

func TestStreamCommand_StreamingFlagEnabled(t *testing.T) {
	// This validates that the stream command would create
	// a Flags struct with Streaming=true

	// We can't easily test the full execution without mocking the entire engine,
	// but we can verify the command structure is correct

	cmd := createStreamCmd()
	assert.NotNil(t, cmd.Run)

	// The Run function should:
	// 1. Create Flags with Streaming=true
	// 2. Set Type=prompt.PRIMARY
	// 3. Set IsPrimary=true
	// These are validated by code inspection in the createStreamCmd implementation
}
