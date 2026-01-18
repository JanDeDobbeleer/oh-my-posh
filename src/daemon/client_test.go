package daemon

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/daemon/ipc"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_DaemonNotRunning(t *testing.T) {
	// Set up temp directories to avoid interfering with real daemon
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	// No daemon running, should fail
	client, err := NewClient()
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestNewClient_DaemonRunning(t *testing.T) {
	// Set up temp directories
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	// Start daemon
	d, err := New(createTestConfig(t))
	require.NoError(t, err)

	go func() {
		_ = d.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	defer func() {
		d.shutdown()
		<-d.Done()
	}()

	// Client should connect
	client, err := NewClient()
	require.NoError(t, err)
	require.NotNil(t, client)

	err = client.Close()
	assert.NoError(t, err)
}

func TestConnectOrStart(t *testing.T) {
	// Set up temp directories
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	// Define a starter function that actually starts a daemon in this process
	// (simulating the external process start)
	startFunc := func() error {
		d, err := New(createTestConfig(t))
		if err != nil {
			return err
		}
		go func() {
			_ = d.Start()
		}()
		// Give it a tiny bit of time to start listening before the retrying NewClient hits
		for range 50 {
			if client, err := NewClient(); err == nil {
				client.Close()
				return nil
			}
			time.Sleep(10 * time.Millisecond)
		}
		return nil
	}

	// 1. Initial state: No daemon running.
	// 2. Call ConnectOrStart
	client, err := ConnectOrStart(startFunc)

	// 3. Should succeed
	require.NoError(t, err)
	require.NotNil(t, client)

	defer client.Close()

	// 4. Verify we can make a call
	assert.True(t, IsRunning())
}

func TestConnectOrStart_StartFails(t *testing.T) {
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	startFunc := func() error {
		return fmt.Errorf("simulated start failure")
	}

	client, err := ConnectOrStart(startFunc)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "simulated start failure")
}

func TestClient_RenderPrompt(t *testing.T) {
	// Set up temp directories
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)
	t.Setenv("POSH_SESSION_ID", "test-session-123")

	configPath := createTestConfig(t)
	// Start daemon
	d, err := New(configPath)
	require.NoError(t, err)

	go func() {
		_ = d.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	defer func() {
		d.shutdown()
		<-d.Done()
	}()

	// Connect client
	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	// Render prompt
	flags := &runtime.Flags{
		ConfigPath: configPath,
		PWD:        "/home/user/project",
		Shell:      "bash",
		Type:       "primary",
		IsPrimary:  true,
	}

	var responses []*ipc.PromptResponse
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.RenderPrompt(ctx, flags, 0, "", nil, func(resp *ipc.PromptResponse) bool {
		responses = append(responses, resp)
		return true
	})
	require.NoError(t, err)

	// Should have received at least one response
	require.NotEmpty(t, responses)

	// Last response should be complete
	lastResp := responses[len(responses)-1]
	assert.Equal(t, "complete", lastResp.Type)
	assert.NotEmpty(t, lastResp.RequestId)
}

func TestClient_RenderPromptSync(t *testing.T) {
	// Set up temp directories
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)
	t.Setenv("POSH_SESSION_ID", "test-session-sync")

	configPath := createTestConfig(t)
	// Start daemon
	d, err := New(configPath)
	require.NoError(t, err)

	go func() {
		_ = d.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	defer func() {
		d.shutdown()
		<-d.Done()
	}()

	// Connect client
	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	// Render prompt sync
	flags := &runtime.Flags{
		ConfigPath: configPath,
		PWD:        "/home/user/project",
		Shell:      "bash",
		Type:       "primary",
		IsPrimary:  true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.RenderPromptSync(ctx, flags, 0, "", nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "complete", resp.Type)
	assert.Contains(t, resp.Prompts, "primary")
}

func TestClient_CallbackStopsOnFalse(t *testing.T) {
	// Set up temp directories
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	configPath := createTestConfig(t)
	// Start daemon
	d, err := New(configPath)
	require.NoError(t, err)

	go func() {
		_ = d.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	defer func() {
		d.shutdown()
		<-d.Done()
	}()

	// Connect client
	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	// Render prompt with callback
	flags := &runtime.Flags{
		ConfigPath: configPath,
		PWD:        "/home/user/project",
		Shell:      "bash",
		Type:       "primary",
		IsPrimary:  true,
	}

	callCount := 0
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.RenderPrompt(ctx, flags, 0, "", nil, func(_ *ipc.PromptResponse) bool {
		callCount++
		return false // Stop after first response
	})
	require.NoError(t, err)

	assert.Equal(t, 1, callCount)
}

func TestIsRunning(t *testing.T) {
	// Set up temp directories
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	// No daemon running
	assert.False(t, IsRunning())

	// Start daemon
	d, err := New(createTestConfig(t))
	require.NoError(t, err)

	go func() {
		_ = d.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	defer func() {
		d.shutdown()
		<-d.Done()
	}()

	// Daemon should be running
	assert.True(t, IsRunning())
}

func TestExtractPrompts(t *testing.T) {
	tests := []struct {
		response *ipc.PromptResponse
		expected *PromptResult
		name     string
	}{
		{
			name:     "nil response",
			response: nil,
			expected: &PromptResult{},
		},
		{
			name: "nil prompts map",
			response: &ipc.PromptResponse{
				Type:    "complete",
				Prompts: nil,
			},
			expected: &PromptResult{},
		},
		{
			name: "primary only",
			response: &ipc.PromptResponse{
				Type: "complete",
				Prompts: map[string]*ipc.Prompt{
					"primary": {Text: "$ "},
				},
			},
			expected: &PromptResult{
				Primary: "$ ",
			},
		},
		{
			name: "all prompts",
			response: &ipc.PromptResponse{
				Type: "complete",
				Prompts: map[string]*ipc.Prompt{
					"primary":   {Text: "primary-text"},
					"right":     {Text: "right-text"},
					"secondary": {Text: "secondary-text"},
					"transient": {Text: "transient-text"},
					"debug":     {Text: "debug-text"},
					"tooltip":   {Text: "tooltip-text"},
					"valid":     {Text: "valid-text"},
					"error":     {Text: "error-text"},
				},
			},
			expected: &PromptResult{
				Primary:   "primary-text",
				Right:     "right-text",
				Secondary: "secondary-text",
				Transient: "transient-text",
				Debug:     "debug-text",
				Tooltip:   "tooltip-text",
				Valid:     "valid-text",
				Error:     "error-text",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ExtractPrompts(tc.response)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetSessionID(t *testing.T) {
	// Set up temp directories
	tmpDir := t.TempDir()
	setTestEnv(t, tmpDir)

	// Test with environment variable set
	t.Setenv("POSH_SESSION_ID", "env-session-id")
	assert.Equal(t, "env-session-id", getSessionID())

	// Test without environment variable (falls back to cache)
	t.Setenv("POSH_SESSION_ID", "")
	sessionID := getSessionID()
	assert.NotEmpty(t, sessionID)
}
