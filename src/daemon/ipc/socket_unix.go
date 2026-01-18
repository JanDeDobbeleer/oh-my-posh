//go:build !windows

package ipc

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Dial connects to the daemon via gRPC over Unix socket.
func Dial(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	target := dialTarget()
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	opts = append(defaultOpts, opts...)
	return grpc.NewClient(target, opts...)
}

func socketPath() string {
	// Try XDG_RUNTIME_DIR first (for systemd-based systems)
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		return filepath.Join(runtimeDir, fmt.Sprintf("oh-my-posh-%d.sock", os.Getuid()))
	}

	// Fall back to state directory
	statePath := os.Getenv("XDG_STATE_HOME")
	if statePath == "" {
		statePath = filepath.Join(path.Home(), ".local", "state")
	}

	return filepath.Join(statePath, "oh-my-posh", fmt.Sprintf("oh-my-posh-%d.sock", os.Getuid()))
}

// dialTarget returns the gRPC dial target for Unix sockets.
func dialTarget() string {
	return "unix://" + socketPath()
}

func listen() (net.Listener, error) {
	sockPath := socketPath()

	// Remove stale socket if exists
	_ = os.Remove(sockPath)

	lc := net.ListenConfig{}
	listener, err := lc.Listen(context.Background(), "unix", sockPath)
	if err != nil {
		return nil, err
	}

	// Set socket permissions to 0600 (owner read/write only)
	if err := os.Chmod(sockPath, 0o600); err != nil {
		_ = listener.Close()
		return nil, err
	}

	return listener, nil
}

func cleanupSocket() error {
	return os.Remove(socketPath())
}

func socketExists() bool {
	_, err := os.Stat(socketPath())
	return err == nil
}
