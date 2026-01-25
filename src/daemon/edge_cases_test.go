package daemon

import (
	"context"
	"os"
	"path/filepath"
	goruntime "runtime"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/daemon/ipc"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEdgeCase_SocketDeleted verifies behavior when the socket file is deleted
// while the daemon is running.
// Skipped on Windows: named pipes cannot be deleted while in use.
func TestEdgeCase_SocketDeleted(t *testing.T) {
	if goruntime.GOOS == windowsOS {
		t.Skip("Windows named pipes cannot be deleted while in use")
	}
	// Setup daemon
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

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

	// Verify connection works
	client, err := NewClient()
	require.NoError(t, err)
	client.Close()

	// Delete the socket file
	socketPath := ipc.SocketPath()
	err = os.Remove(socketPath)
	require.NoError(t, err)

	// Try to connect again - should fail
	// Note: NewClient has a retry loop, so it will wait until timeout
	// We expect it to eventually fail because the listener is gone (from filesystem perspective)
	// although the daemon process still has the file descriptor open.
	// On Unix, connecting to a deleted socket path fails.
	client2, err := NewClient()
	assert.Error(t, err)
	if client2 != nil {
		client2.Close()
	}
}

// TestEdgeCase_CorruptConfig verifies that the daemon handles corrupt config files gracefully.
func TestEdgeCase_CorruptConfig(t *testing.T) {
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

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

	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	// Create a corrupt config file
	configPath := filepath.Join(tmpDir, "corrupt.json")
	err = os.WriteFile(configPath, []byte(`{ "invalid json": `), 0644)
	require.NoError(t, err)

	flags := &runtime.Flags{
		ConfigPath: configPath,
		PWD:        "/tmp",
		Shell:      "bash",
		Type:       "primary",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Request should succeed with default config that displays the error
	// config.Load() falls back to default config on parse errors
	resp, err := client.RenderPromptSync(ctx, flags, 0, "", nil)

	// Should not error - daemon gracefully falls back to default config
	require.NoError(t, err)
	require.NotNil(t, resp)
	// The default config includes the error message in the prompt
	assert.NotEmpty(t, resp.Prompts["primary"].Text)
}
