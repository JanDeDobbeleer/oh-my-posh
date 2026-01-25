package daemon

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/daemon/ipc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSocketDir creates a short temp directory suitable for Unix socket paths.
// Unix sockets have a path length limit (~104 chars on macOS), so we use /tmp directly.
func testSocketDir(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		return t.TempDir()
	}
	dir, err := os.MkdirTemp("/tmp", "omp")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestNewDaemon(t *testing.T) {
	// Set up temp directory for state path
	tmpDir := t.TempDir()
	setTestEnv(t, tmpDir)

	d, err := New(createTestConfig(t))
	require.NoError(t, err)
	require.NotNil(t, d)

	assert.NotNil(t, d.lockFile)
	assert.NotNil(t, d.cache)
	assert.NotNil(t, d.activeRenders)
	assert.NotNil(t, d.sessions)
	assert.NotNil(t, d.done)

	// Clean up
	d.shutdown()
	<-d.Done()
}

func TestDaemonSingleInstance(t *testing.T) {
	// Set up temp directory for state path
	tmpDir := t.TempDir()
	setTestEnv(t, tmpDir)

	// First daemon should succeed
	d1, err := New(createTestConfig(t))
	require.NoError(t, err)
	defer func() {
		d1.shutdown()
		<-d1.Done()
	}()

	// Second daemon should fail due to lock
	d2, err := New(createTestConfig(t))
	assert.Error(t, err)
	assert.Nil(t, d2)
	assert.Contains(t, err.Error(), "failed to acquire lock")
}

func TestDaemonStartAndShutdown(t *testing.T) {
	// Set up temp directories (use short path for socket)
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	d, err := New(createTestConfig(t))
	require.NoError(t, err)

	// Start daemon in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- d.Start()
	}()

	// Give daemon time to start
	time.Sleep(100 * time.Millisecond)

	// Shutdown daemon
	d.shutdown()

	// Wait for daemon to stop
	select {
	case <-d.Done():
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Daemon did not stop within timeout")
	}

	// Check Start returned without error
	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("Start did not return")
	}
}

func TestDaemonRenderPrompt(t *testing.T) {
	// Set up temp directories (use short path for socket)
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	d, err := New(createTestConfig(t))
	require.NoError(t, err)

	// Start daemon in background
	go func() {
		_ = d.Start()
	}()

	// Give daemon time to start
	time.Sleep(100 * time.Millisecond)

	defer func() {
		d.shutdown()
		<-d.Done()
	}()

	// Connect to daemon
	conn, err := ipc.Dial()
	require.NoError(t, err)
	defer conn.Close()

	client := ipc.NewDaemonServiceClient(conn)

	// Send render request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.RenderPrompt(ctx, &ipc.PromptRequest{
		Version:   ipc.ProtocolVersion,
		SessionId: "test-session",
		RequestId: "test-request-1",
		Flags:     &ipc.Flags{Pwd: "/home/user"},
	})
	require.NoError(t, err)

	// Receive response
	resp, err := stream.Recv()
	require.NoError(t, err)

	assert.Equal(t, "complete", resp.Type)
	assert.Equal(t, "test-request-1", resp.RequestId)
	assert.Contains(t, resp.Prompts, "primary")
}

func TestDaemonVersionMismatch(t *testing.T) {
	// Set up temp directories (use short path for socket)
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	d, err := New(createTestConfig(t))
	require.NoError(t, err)

	// Start daemon in background
	go func() {
		_ = d.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	defer func() {
		d.shutdown()
		<-d.Done()
	}()

	// Connect to daemon
	conn, err := ipc.Dial()
	require.NoError(t, err)
	defer conn.Close()

	client := ipc.NewDaemonServiceClient(conn)

	// Send request with wrong version
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.RenderPrompt(ctx, &ipc.PromptRequest{
		Version:   999, // Wrong version
		SessionId: "test-session",
		RequestId: "test-request-1",
	})
	require.NoError(t, err)

	// Should get error on recv
	_, err = stream.Recv()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "version mismatch")
}

func TestDaemonCancelsExistingSessionRequest(t *testing.T) {
	// Set up temp directories (use short path for socket)
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	d, err := New(createTestConfig(t))
	require.NoError(t, err)

	// Start daemon in background
	go func() {
		_ = d.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	defer func() {
		d.shutdown()
		<-d.Done()
	}()

	// Connect to daemon
	conn, err := ipc.Dial()
	require.NoError(t, err)
	defer conn.Close()

	client := ipc.NewDaemonServiceClient(conn)

	// Send first request
	ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel1()

	stream1, err := client.RenderPrompt(ctx1, &ipc.PromptRequest{
		Version:   ipc.ProtocolVersion,
		SessionId: "same-session",
		RequestId: "request-1",
		Flags:     &ipc.Flags{Pwd: "/home/user"},
	})
	require.NoError(t, err)

	// Receive first response
	resp1, err := stream1.Recv()
	require.NoError(t, err)
	assert.Equal(t, "request-1", resp1.RequestId)

	// Send second request for same session
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	stream2, err := client.RenderPrompt(ctx2, &ipc.PromptRequest{
		Version:   ipc.ProtocolVersion,
		SessionId: "same-session",
		RequestId: "request-2",
		Flags:     &ipc.Flags{Pwd: "/home/user"},
	})
	require.NoError(t, err)

	// Receive second response
	resp2, err := stream2.Recv()
	require.NoError(t, err)
	assert.Equal(t, "request-2", resp2.RequestId)
}

func TestDaemonCacheGetTTL(t *testing.T) {
	// Set up temp directories (use short path for socket)
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	d, err := New(createTestConfig(t))
	require.NoError(t, err)

	// Start daemon in background
	go func() {
		_ = d.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	defer func() {
		d.shutdown()
		<-d.Done()
	}()

	// Connect to daemon
	conn, err := ipc.Dial()
	require.NoError(t, err)
	defer conn.Close()

	client := ipc.NewDaemonServiceClient(conn)

	// Get default TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.CacheGetTTL(ctx, &ipc.CacheGetTTLRequest{})
	require.NoError(t, err)

	// Default TTL should be 7 days
	assert.Equal(t, int32(7), resp.Days)
}

func TestDaemonCacheSetTTL(t *testing.T) {
	// Set up temp directories (use short path for socket)
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	d, err := New(createTestConfig(t))
	require.NoError(t, err)

	// Start daemon in background
	go func() {
		_ = d.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	defer func() {
		d.shutdown()
		<-d.Done()
	}()

	// Connect to daemon
	conn, err := ipc.Dial()
	require.NoError(t, err)
	defer conn.Close()

	client := ipc.NewDaemonServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set TTL to 14 days
	setResp, err := client.CacheSetTTL(ctx, &ipc.CacheSetTTLRequest{Days: 14})
	require.NoError(t, err)
	assert.True(t, setResp.Success)

	// Verify TTL was changed
	getResp, err := client.CacheGetTTL(ctx, &ipc.CacheGetTTLRequest{})
	require.NoError(t, err)
	assert.Equal(t, int32(14), getResp.Days)
}

func TestDaemonCacheClear(t *testing.T) {
	// Set up temp directories (use short path for socket)
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	d, err := New(createTestConfig(t))
	require.NoError(t, err)

	// Pre-populate cache with some data
	d.cache.Set("session1", "key1", "value1", 0)
	d.cache.Set("session2", "key2", "value2", 0)
	assert.Equal(t, 2, d.cache.SessionCount())

	// Start daemon in background
	go func() {
		_ = d.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	defer func() {
		d.shutdown()
		<-d.Done()
	}()

	// Connect to daemon
	conn, err := ipc.Dial()
	require.NoError(t, err)
	defer conn.Close()

	client := ipc.NewDaemonServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Clear cache
	resp, err := client.CacheClear(ctx, &ipc.CacheClearRequest{})
	require.NoError(t, err)
	assert.True(t, resp.Success)

	// Verify cache was cleared
	assert.Equal(t, 0, d.cache.SessionCount())
	assert.Equal(t, 0, d.cache.EntryCount())
}
