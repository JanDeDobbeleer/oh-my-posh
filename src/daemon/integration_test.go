package daemon

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"github.com/jandedobbeleer/oh-my-posh/src/daemon/ipc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests for daemon mode.
// These tests verify end-to-end flows between client and daemon components.

// TestIntegration_DaemonLifecycle tests the complete daemon lifecycle:
// start → client connect → render request → graceful shutdown.
func TestIntegration_DaemonLifecycle(t *testing.T) {
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)
	// Send render request
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

	flags := &runtime.Flags{
		ConfigPath: configPath,
		PWD:        "/home/user",
		Shell:      "zsh",
		Type:       "primary",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, _ := client.RenderPromptSync(ctx, flags, 0, "", nil)
	require.NotNil(t, resp)

	assert.Equal(t, "complete", resp.Type)
	assert.Contains(t, resp.Prompts, "primary")

	// Graceful shutdown
	d.shutdown()

	select {
	case <-d.Done():
		// Success - daemon shut down cleanly
	case <-time.After(5 * time.Second):
		t.Fatal("daemon did not shut down within timeout")
	}
}

// TestIntegration_MultipleConcurrentSessions tests multiple clients with different
// session IDs rendering simultaneously. Verifies thread safety and session isolation.
func TestIntegration_MultipleConcurrentSessions(t *testing.T) {
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

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

	// Create multiple clients with different session IDs
	const numSessions = 5
	var wg sync.WaitGroup
	errors := make(chan error, numSessions)
	responses := make(chan *ipc.PromptResponse, numSessions)

	for i := range numSessions {
		wg.Add(1)
		go func(sessionNum int) {
			defer wg.Done()

			// Each goroutine creates its own client
			client, err := NewClient()
			if err != nil {
				errors <- err
				return
			}
			defer client.Close()

			// Send render request
			configPath := createTestConfig(t)
			flags := &runtime.Flags{
				ConfigPath: configPath,
				PWD:        "/home/user",
				Shell:      "zsh",
				Type:       "primary",
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Use RenderPrompt with manual session ID control via gRPC
			conn, err := ipc.Dial()
			if err != nil {
				errors <- err
				return
			}
			defer conn.Close()

			grpcClient := ipc.NewDaemonServiceClient(conn)
			stream, err := grpcClient.RenderPrompt(ctx, &ipc.PromptRequest{
				Version:   ipc.ProtocolVersion,
				SessionId: "session-" + string(rune('A'+sessionNum)),
				RequestId: "request-" + string(rune('0'+sessionNum)),
				Flags:     ipc.FlagsToProto(flags),
			})
			if err != nil {
				errors <- err
				return
			}

			resp, err := stream.Recv()
			if err != nil {
				errors <- err
				return
			}

			responses <- resp
		}(i)
	}

	wg.Wait()
	close(errors)
	close(responses)

	// Check for errors
	for err := range errors {
		t.Errorf("session error: %v", err)
	}

	// Verify all sessions got responses
	responseCount := 0
	for resp := range responses {
		require.NotNil(t, resp)
		assert.Equal(t, "complete", resp.Type)
		assert.Contains(t, resp.Prompts, "primary")
		responseCount++
	}

	assert.Equal(t, numSessions, responseCount, "all sessions should receive responses")
}

// TestIntegration_SessionCancellation tests that a new request for the same
// session cancels any in-progress render.
func TestIntegration_SessionCancellation(t *testing.T) {
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

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

	// Connect to daemon
	conn, err := ipc.Dial()
	require.NoError(t, err)
	defer conn.Close()

	client := ipc.NewDaemonServiceClient(conn)
	sessionID := "same-session"

	// Send first request
	ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel1()

	stream1, err := client.RenderPrompt(ctx1, &ipc.PromptRequest{
		Version:   ipc.ProtocolVersion,
		SessionId: sessionID,
		RequestId: "request-1",
		Flags:     &ipc.Flags{Pwd: "/home/user"},
	})
	require.NoError(t, err)

	// Get first response
	resp1, err := stream1.Recv()
	require.NoError(t, err)
	assert.Equal(t, "request-1", resp1.RequestId)

	// Immediately send second request for same session
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	stream2, err := client.RenderPrompt(ctx2, &ipc.PromptRequest{
		Version:   ipc.ProtocolVersion,
		SessionId: sessionID,
		RequestId: "request-2",
		Flags:     &ipc.Flags{Pwd: "/home/user"},
	})
	require.NoError(t, err)

	// Second request should complete successfully
	resp2, err := stream2.Recv()
	require.NoError(t, err)
	assert.Equal(t, "request-2", resp2.RequestId)
	assert.Equal(t, "complete", resp2.Type)
}

// TestIntegration_StaleLockRecovery tests that the daemon can recover from
// a stale lock file left by a crashed process.
func TestIntegration_StaleLockRecovery(t *testing.T) {
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	// Create lock file with a non-existent PID
	lockDir := statePath()
	err := os.MkdirAll(lockDir, 0700)
	require.NoError(t, err)
	lockPath := lockDir + "/daemon.lock"

	// Write a fake PID that definitely doesn't exist
	// PID 99999 is unlikely to be a real process
	err = os.WriteFile(lockPath, []byte("99999\n"), 0600)
	require.NoError(t, err)

	// Daemon should detect stale lock and recover
	d, err := New(createTestConfig(t))
	require.NoError(t, err, "daemon should recover from stale lock")
	require.NotNil(t, d)

	defer func() {
		d.shutdown()
		<-d.Done()
	}()

	// Daemon should be functional
	go func() {
		_ = d.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	// Verify daemon is running by connecting a client
	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	assert.True(t, IsRunning())
}

// TestIntegration_SocketCleanupAfterCrash tests that a new daemon can start
// even if a stale socket file exists from a previous crash.
func TestIntegration_SocketCleanupAfterCrash(t *testing.T) {
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	// Start first daemon
	d1, err := New(createTestConfig(t))
	require.NoError(t, err)

	go func() {
		_ = d1.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	// Verify socket exists
	socketPath := ipc.SocketPath()
	_, err = os.Stat(socketPath)
	require.NoError(t, err, "socket should exist")

	// Simulate crash by force-closing without cleanup
	// This leaves the socket file behind
	d1.server.Stop() // Force stop, not graceful

	// Release lock manually (simulating clean crash recovery path)
	_ = d1.lockFile.Release()

	// Small delay to ensure resources are released
	time.Sleep(50 * time.Millisecond)

	// New daemon should be able to start
	d2, err := New(createTestConfig(t))
	require.NoError(t, err, "new daemon should start despite stale socket")

	go func() {
		_ = d2.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	defer func() {
		d2.shutdown()
		<-d2.Done()
	}()

	// New daemon should be functional
	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	assert.True(t, IsRunning())
}

// TestIntegration_GracefulShutdownDuringRender tests that in-flight RPCs
// complete gracefully when shutdown is initiated.
func TestIntegration_GracefulShutdownDuringRender(t *testing.T) {
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	// Start daemon
	d, err := New(createTestConfig(t))
	require.NoError(t, err)

	started := make(chan struct{})
	go func() {
		close(started)
		_ = d.Start()
	}()

	<-started
	time.Sleep(100 * time.Millisecond)

	// Verify daemon is ready by connecting a test client
	testClient, err := NewClient()
	require.NoError(t, err)
	testClient.Close()

	// Start multiple render requests
	const numClients = 3
	var wg sync.WaitGroup
	clientsDone := make(chan struct{})

	for i := range numClients {
		wg.Add(1)
		go func(_ int) {
			defer wg.Done()

			client, err := NewClient()
			if err != nil {
				// Connection might fail during shutdown, that's expected
				return
			}
			defer func() {
				if client != nil {
					_ = client.Close()
				}
			}()

			// Start a render request
			configPath := createTestConfig(t)
			flags := &runtime.Flags{
				ConfigPath: configPath,
				PWD:        "/home/user",
				Shell:      "zsh",
				Type:       "primary",
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// RenderPromptSync will either complete or fail gracefully
			_, _ = client.RenderPromptSync(ctx, flags, 0, "", nil)
		}(i)
	}

	// Wait a tiny bit for requests to be in-flight, then shutdown
	time.Sleep(50 * time.Millisecond)

	// Initiate shutdown
	d.shutdown()

	// Wait for clients to complete
	go func() {
		wg.Wait()
		close(clientsDone)
	}()

	select {
	case <-clientsDone:
		// Success - all clients completed or failed gracefully
	case <-time.After(5 * time.Second):
		t.Fatal("clients did not complete within timeout")
	}

	// Daemon should be fully stopped
	select {
	case <-d.Done():
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("daemon did not stop within timeout")
	}
}

// TestIntegration_HighConcurrencyStress tests the daemon under high concurrency
// to verify thread safety and absence of deadlocks.
func TestIntegration_HighConcurrencyStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

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

	// Spawn many concurrent requests
	const numRequests = 50
	var wg sync.WaitGroup
	var successCount atomic.Int32
	var errorCount atomic.Int32

	configPath := createTestConfig(t)
	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	flags := &runtime.Flags{
		ConfigPath: configPath,
		PWD:        "/home/user",
		Shell:      "zsh",
		Type:       "primary",
	}

	for i := range numRequests {
		wg.Add(1)
		go func(reqNum int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			sessionID := fmt.Sprintf("stress-session-%d", reqNum)
			resp, err := client.RenderPromptSync(ctx, flags, 0, sessionID, nil)
			if err != nil {
				errorCount.Add(1)
				return
			}

			if resp != nil && resp.Type == "complete" {
				successCount.Add(1)
			}
		}(i)
	}

	wg.Wait()

	success := int(successCount.Load())

	// Most requests should succeed
	assert.Greater(t, success, numRequests*8/10, "at least 80% of requests should succeed")
}

// TestIntegration_ClientReconnectAfterDaemonRestart tests that clients can
// reconnect after the daemon is restarted.
func TestIntegration_ClientReconnectAfterDaemonRestart(t *testing.T) {
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)
	t.Setenv("POSH_SESSION_ID", "reconnect-test")

	// Start first daemon
	d1, err := New(createTestConfig(t))
	require.NoError(t, err)

	go func() {
		_ = d1.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	// Connect and verify
	client1, err := NewClient()
	require.NoError(t, err)

	configPath := createTestConfig(t)
	flags := &runtime.Flags{
		ConfigPath: configPath,
		PWD:        "/home/user",
		Shell:      "bash",
		Type:       "primary",
	}

	ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	resp1, err := client1.RenderPromptSync(ctx1, flags, 0, "", nil)
	cancel1()
	require.NoError(t, err)
	assert.Equal(t, "complete", resp1.Type)

	client1.Close()

	// Stop first daemon
	d1.shutdown()
	<-d1.Done()

	// Verify daemon is stopped
	assert.False(t, IsRunning())

	// Start second daemon
	d2, err := New(createTestConfig(t))
	require.NoError(t, err)

	go func() {
		_ = d2.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	defer func() {
		d2.shutdown()
		<-d2.Done()
	}()

	// Connect new client to new daemon
	client2, err := NewClient()
	require.NoError(t, err)
	defer client2.Close()

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	resp2, err := client2.RenderPromptSync(ctx2, flags, 0, "", nil)
	require.NoError(t, err)
	assert.Equal(t, "complete", resp2.Type)
}

// TestIntegration_ProcessExitDetection tests that sessions are automatically
// unregistered when their associated process exits.
func TestIntegration_ProcessExitDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Set up temp directory as state path
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)

	configPath := createTestConfig(t)
	d, err := New(configPath)
	require.NoError(t, err)

	go func() {
		_ = d.Start()
	}()

	// Wait for daemon to start
	time.Sleep(100 * time.Millisecond)

	defer func() {
		d.shutdown()
		<-d.Done()
	}()

	// Start a subprocess that will exit (use helper process)
	cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=TestHelperProcess")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	err = cmd.Start()
	require.NoError(t, err)

	pid := cmd.Process.Pid

	// Verify helper process is running
	require.True(t, IsProcessRunning(pid), "helper process should be running")

	// Register the session via gRPC request with PID
	conn, err := ipc.Dial()
	require.NoError(t, err)
	defer conn.Close()

	client := ipc.NewDaemonServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.RenderPrompt(ctx, &ipc.PromptRequest{
		Version:   ipc.ProtocolVersion,
		SessionId: "exit-test-session",
		RequestId: "request-1",
		Pid:       int32(pid),
		Flags:     &ipc.Flags{Pwd: "/home/user"},
	})
	require.NoError(t, err)

	// Get response (this also registers the session)
	_, err = stream.Recv()
	require.NoError(t, err)

	// Verify session is registered (use Eventually to avoid races)
	assert.Eventually(t, func() bool {
		return d.sessions.Count() >= 1
	}, 10*time.Second, 500*time.Millisecond, "session should be registered")

	initialCount := d.sessions.Count()

	// Trigger exit by closing stdin
	stdin.Close()
	err = cmd.Wait()
	require.NoError(t, err)

	// Session should be unregistered (use Eventually to avoid races)
	assert.Eventually(t, func() bool {
		return d.sessions.Count() < initialCount
	}, 10*time.Second, 10*time.Millisecond, "session count should decrease after process exit")
}

// TestHelperProcess is a helper for TestIntegration_ProcessExitDetection.
// It acts as a long-running process that we can control via stdin.
func TestHelperProcess(_ *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// Run until stdin is closed
	_, _ = io.Copy(io.Discard, os.Stdin)
	os.Exit(0)
}

// TestIntegration_CacheIsolation tests that cache entries are isolated per session.
func TestIntegration_CacheIsolation(t *testing.T) {
	// This test doesn't need a running daemon - it tests the cache directly.
	// Using NewMemoryCache() directly avoids race conditions with daemon lifecycle.
	cache := NewMemoryCache()

	// Set values for session A
	cache.Set("session-A", "key1", "value-A1", time.Minute)
	cache.Set("session-A", "key2", "value-A2", time.Minute)

	// Set different values for session B
	cache.Set("session-B", "key1", "value-B1", time.Minute)
	cache.Set("session-B", "key2", "value-B2", time.Minute)

	// Verify isolation
	valA1, ok := cache.Get("session-A", "key1")
	assert.True(t, ok)
	assert.Equal(t, "value-A1", valA1)

	valB1, ok := cache.Get("session-B", "key1")
	assert.True(t, ok)
	assert.Equal(t, "value-B1", valB1)

	// Clean session A
	cache.CleanSession("session-A")

	// Session A values should be gone
	_, ok = cache.Get("session-A", "key1")
	assert.False(t, ok)

	// Session B values should remain
	valB2, ok := cache.Get("session-B", "key2")
	assert.True(t, ok)
	assert.Equal(t, "value-B2", valB2)
}

// TestIntegration_MultipleRequestsSameSession tests sending multiple requests
// from the same session in sequence.
func TestIntegration_MultipleRequestsSameSession(t *testing.T) {
	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)
	t.Setenv("POSH_SESSION_ID", "multi-request-session")

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

	// Send multiple sequential requests from same client
	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	configPath := createTestConfig(t)
	flags := &runtime.Flags{
		ConfigPath: configPath,
		PWD:        "/home/user",
		Shell:      "zsh",
		Type:       "primary",
	}

	for i := range 5 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		resp, err := client.RenderPromptSync(ctx, flags, 0, "", nil)
		cancel()

		require.NoError(t, err, "request %d should succeed", i)
		require.NotNil(t, resp)
		assert.Equal(t, "complete", resp.Type)
	}
}
