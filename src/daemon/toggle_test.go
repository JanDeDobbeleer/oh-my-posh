package daemon

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDaemon_ToggleMismatch confirms that ToggleSegment correctly targets the session
// identified by the shell's PID when POSH_SESSION_ID is missing.
func TestDaemon_ToggleMismatch(t *testing.T) {
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	// Ensure POSH_SESSION_ID is NOT set, so we rely on PID vs UUID matching
	t.Setenv("POSH_SESSION_ID", "")

	// Start daemon
	configPath := createTestConfig(t)
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

	// 1. Simulate Shell: RenderPrompt with PID
	clientA, err := NewClient()
	require.NoError(t, err)
	defer clientA.Close()

	shellPID := 12345
	flags := &runtime.Flags{
		ConfigPath: configPath,
		PWD:        "/home/user",
		Shell:      "zsh",
		Type:       "primary",
	}

	// Render first time to register session (using PID)
	ctxA, cancelA := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelA()

	// We pass PID 12345. Client.RenderPrompt uses this PID for the session ID.
	respA, err := clientA.RenderPromptSync(ctxA, flags, shellPID, "", nil)
	require.NoError(t, err)
	require.Equal(t, "complete", respA.Type)

	// 2. Simulate Toggle: ToggleSegment (simulating 'omp toggle segment')
	clientB, err := NewClient()
	require.NoError(t, err)
	defer clientB.Close()

	// Toggle "text" segment (present in createTestConfig)
	ctxB, cancelB := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelB()

	// We pass the shellPID to ToggleSegment. It should prioritize this over generating a new UUID.
	err = clientB.ToggleSegment(ctxB, shellPID, []string{"text"})
	require.NoError(t, err)

	// 3. Verify Shell session
	// We inspect the daemon's cache directly to see if shellPID ("12345") has the toggle
	shellSessionID := fmt.Sprintf("%d", shellPID)
	toggleMapRaw, ok := d.cache.Get(shellSessionID, cache.TOGGLECACHE)

	require.True(t, ok, "Shell session (PID based) should have received the toggle")
	m := toggleMapRaw.(map[string]bool)
	assert.True(t, m["text"], "Segment 'text' should be toggled in shell session")
}

// TestDaemon_ToggleMultipleSegments verifies toggling multiple segments in a single call.
func TestDaemon_ToggleMultipleSegments(t *testing.T) {
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)
	t.Setenv("POSH_SESSION_ID", "multi-toggle-session")

	// Start daemon
	configPath := createTestConfig(t)
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

	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Toggle "seg1" and "seg2"
	err = client.ToggleSegment(ctx, 0, []string{"seg1", "seg2"})
	require.NoError(t, err)

	// Verify cache
	toggleMapRaw, ok := d.cache.Get("multi-toggle-session", cache.TOGGLECACHE)
	require.True(t, ok)
	m := toggleMapRaw.(map[string]bool)
	assert.True(t, m["seg1"])
	assert.True(t, m["seg2"])

	// Toggle "seg1" again (should untoggle) and "seg3" (should toggle)
	err = client.ToggleSegment(ctx, 0, []string{"seg1", "seg3"})
	require.NoError(t, err)

	toggleMapRaw, ok = d.cache.Get("multi-toggle-session", cache.TOGGLECACHE)
	require.True(t, ok)
	m = toggleMapRaw.(map[string]bool)
	assert.False(t, m["seg1"], "seg1 should be untoggled")
	assert.True(t, m["seg2"], "seg2 should remain toggled")
	assert.True(t, m["seg3"], "seg3 should be toggled")
}

// TestDaemon_ToggleWithSessionID verifies that ToggleSegment respects POSH_SESSION_ID
// when it is provided in the environment.
func TestDaemon_ToggleWithSessionID(t *testing.T) {
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	sessionID := "explicit-session-id"
	t.Setenv("POSH_SESSION_ID", sessionID)

	// Start daemon
	configPath := createTestConfig(t)
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

	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Toggle without passing a PID. It should use POSH_SESSION_ID.
	err = client.ToggleSegment(ctx, 0, []string{"text"})
	require.NoError(t, err)

	// Verify cache using the explicit session ID
	toggleMapRaw, ok := d.cache.Get(sessionID, cache.TOGGLECACHE)
	require.True(t, ok, "Session ID from environment should have been used")
	m := toggleMapRaw.(map[string]bool)
	assert.True(t, m["text"])
}
