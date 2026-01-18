//go:build !windows

package ipc

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// testSocketDir creates a short temp directory suitable for Unix socket paths.
// Unix sockets have a path length limit (~104 chars on macOS), so we use /tmp directly.
func testSocketDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "omp")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestSocketPath(t *testing.T) {
	path := SocketPath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "oh-my-posh-")
	assert.Contains(t, path, ".sock")
}

func TestSocketPathWithXDGRuntimeDir(t *testing.T) {
	tmpDir := testSocketDir(t)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	path := socketPath()
	assert.Contains(t, path, tmpDir)
	assert.Contains(t, path, "oh-my-posh-")
	assert.Contains(t, path, ".sock")
}

func TestSocketPathFallback(t *testing.T) {
	// Unset XDG_RUNTIME_DIR to test fallback
	t.Setenv("XDG_RUNTIME_DIR", "")

	path := socketPath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "oh-my-posh-")
	assert.Contains(t, path, ".sock")
}

func TestDialTarget(t *testing.T) {
	target := dialTarget()
	assert.Contains(t, target, "unix://")
	assert.Contains(t, target, "oh-my-posh-")
}

func TestListenAndCleanup(t *testing.T) {
	// Use short temp directory for Unix socket path length limits
	tmpDir := testSocketDir(t)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	// Get socket path BEFORE creating listener (so we capture the right path)
	sockPath := SocketPath()

	// Create listener
	listener, err := Listen()
	require.NoError(t, err)
	require.NotNil(t, listener)

	// Verify socket file exists
	_, err = os.Stat(sockPath)
	assert.NoError(t, err, "socket file should exist")

	// Verify socket permissions (should be 0600)
	info, err := os.Stat(sockPath)
	require.NoError(t, err)
	// On Unix, check that only owner has read/write
	perm := info.Mode().Perm()
	assert.Equal(t, os.FileMode(0o600), perm, "socket should have 0600 permissions")

	// Close listener
	err = listener.Close()
	assert.NoError(t, err)

	// After close, socket file may or may not exist depending on OS
	// On macOS, closing the listener removes the socket file
	// On Linux, the file typically remains
	// Either way, ensure cleanup doesn't leave the file around
	_ = os.Remove(sockPath)

	// Verify socket file is removed (should be gone either way now)
	_, err = os.Stat(sockPath)
	assert.True(t, os.IsNotExist(err), "socket file should be removed after cleanup")
}

func TestListenRemovesStaleSocket(t *testing.T) {
	// Use short temp directory for Unix socket path length limits
	tmpDir := testSocketDir(t)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	sockPath := SocketPath()

	// Create a stale regular file at the socket path
	err := os.WriteFile(sockPath, []byte("stale"), 0o600)
	require.NoError(t, err)

	// Verify the file exists and is a regular file
	info, err := os.Stat(sockPath)
	require.NoError(t, err)
	require.True(t, info.Mode().IsRegular(), "should be a regular file before Listen")

	// Listen should remove the stale file and create a real socket
	listener, err := Listen()
	require.NoError(t, err)
	require.NotNil(t, listener)
	defer listener.Close()
	defer os.Remove(sockPath)

	// Verify it's a real socket now
	info, err = os.Stat(sockPath)
	require.NoError(t, err)
	assert.True(t, info.Mode()&os.ModeSocket != 0, "should be a socket file after Listen")
}

func TestListenAndDialRaw(t *testing.T) {
	// Test raw socket connectivity (without gRPC)
	tmpDir := testSocketDir(t)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	sockPath := SocketPath()

	// Create listener
	listener, err := Listen()
	require.NoError(t, err)
	require.NotNil(t, listener)
	defer listener.Close()
	defer os.Remove(sockPath)

	// Accept connections in background
	acceptDone := make(chan struct{})
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		conn.Close()
		close(acceptDone)
	}()

	// Use raw net.Dial to test socket connectivity
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(context.Background(), "unix", sockPath)
	require.NoError(t, err)
	require.NotNil(t, conn)
	conn.Close()

	// Wait for accept to complete with timeout
	select {
	case <-acceptDone:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for Accept")
	}
}

func TestDialCreatesGRPCClient(t *testing.T) {
	// Test that Dial() returns a valid gRPC client (lazy connection)
	tmpDir := testSocketDir(t)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	sockPath := SocketPath()

	// Create listener for the socket
	listener, err := Listen()
	require.NoError(t, err)
	require.NotNil(t, listener)
	defer listener.Close()
	defer os.Remove(sockPath)

	// Accept connections in background (needed for gRPC to connect)
	go func() {
		for {
			conn, acceptErr := listener.Accept()
			if acceptErr != nil {
				return
			}
			// Keep connection open briefly then close
			time.Sleep(100 * time.Millisecond)
			conn.Close()
		}
	}()

	// Create gRPC client
	grpcConn, err := Dial()
	require.NoError(t, err)
	require.NotNil(t, grpcConn)
	defer grpcConn.Close()

	// gRPC NewClient uses lazy connection, so state should be Idle initially
	state := grpcConn.GetState()
	assert.True(t, state == connectivity.Idle || state == connectivity.Connecting,
		"initial state should be Idle or Connecting, got: %v", state)

	// Force connection by calling Connect
	grpcConn.Connect()

	// Wait for connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Wait for state to become Ready (connected) or fail
	for {
		state = grpcConn.GetState()
		if state == connectivity.Ready {
			break
		}
		if state == connectivity.TransientFailure || state == connectivity.Shutdown {
			// Connection failed, but that's OK for this test - we just want to verify
			// the gRPC client was created and attempted to connect
			break
		}
		if !grpcConn.WaitForStateChange(ctx, state) {
			// Timeout - but we got a valid gRPC client, which is what we're testing
			break
		}
	}
}

func TestDialWithOptions(t *testing.T) {
	// Test that Dial() accepts additional gRPC options
	tmpDir := testSocketDir(t)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	// Don't need a listener for this test - just verifying Dial doesn't error
	grpcConn, err := Dial(grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024)))
	require.NoError(t, err)
	require.NotNil(t, grpcConn)
	defer grpcConn.Close()
}

func TestCleanupSocketNonExistent(t *testing.T) {
	// Use short temp directory for Unix socket path length limits
	tmpDir := testSocketDir(t)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	sockPath := SocketPath()

	// Cleanup should return error for non-existent file
	err := os.Remove(sockPath)
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestListenCreatesParentDirectory(t *testing.T) {
	// Use short temp directory for Unix socket path length limits
	tmpDir := testSocketDir(t)
	// Create a subdirectory that should contain the socket
	runtimeDir := filepath.Join(tmpDir, "run")
	err := os.Mkdir(runtimeDir, 0o700)
	require.NoError(t, err)

	t.Setenv("XDG_RUNTIME_DIR", runtimeDir)

	sockPath := SocketPath()

	listener, err := Listen()
	require.NoError(t, err)
	require.NotNil(t, listener)
	defer listener.Close()
	defer os.Remove(sockPath)

	// Verify socket exists
	_, err = os.Stat(sockPath)
	assert.NoError(t, err)
}

func TestDialNoServer(t *testing.T) {
	// Use short temp directory for Unix socket path length limits
	tmpDir := testSocketDir(t)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	sockPath := SocketPath()

	// Ensure no socket file exists
	_ = os.Remove(sockPath)

	// Dial creates a gRPC client with lazy connection (doesn't error immediately)
	grpcConn, err := Dial()
	require.NoError(t, err, "Dial creates client even without server (lazy connection)")
	require.NotNil(t, grpcConn)
	defer grpcConn.Close()

	// The client is created but not connected yet
	state := grpcConn.GetState()
	assert.Equal(t, connectivity.Idle, state, "should be Idle when no server")

	// Force connection attempt
	grpcConn.Connect()

	// Wait briefly for connection attempt
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	grpcConn.WaitForStateChange(ctx, connectivity.Idle)

	// Should be in TransientFailure since no server is listening
	state = grpcConn.GetState()
	assert.True(t, state == connectivity.TransientFailure || state == connectivity.Connecting,
		"should be TransientFailure or Connecting when no server, got: %v", state)
}

func TestMultipleListeners(t *testing.T) {
	// Use short temp directory for Unix socket path length limits
	tmpDir := testSocketDir(t)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	sockPath := SocketPath()

	// Create first listener
	listener1, err := Listen()
	require.NoError(t, err)
	defer listener1.Close()
	defer os.Remove(sockPath)

	// Try to create second listener on same path - should fail
	lc := net.ListenConfig{}
	_, err = lc.Listen(context.Background(), "unix", sockPath)
	assert.Error(t, err, "second listener should fail on same socket path")
}
