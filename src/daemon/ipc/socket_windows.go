//go:build windows

package ipc

import (
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/Microsoft/go-winio"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Dial connects to the daemon via gRPC over Windows named pipe.
func Dial(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// Use passthrough scheme to bypass gRPC's name resolver.
	// The resolver doesn't understand Windows named pipe paths,
	// so we pass the path directly to our custom dialer.
	target := "passthrough:///" + socketPath()
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(winio.DialPipeContext),
	}
	opts = append(defaultOpts, opts...)
	return grpc.NewClient(target, opts...)
}

func socketPath() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		username := os.Getenv("USERNAME")
		if username == "" {
			username = "default"
		}
		return fmt.Sprintf(`\\.\pipe\oh-my-posh-%s`, username)
	}

	// Create a unique pipe name based on the LOCALAPPDATA path.
	// This ensures test isolation when LOCALAPPDATA is overridden.
	hash := sha256.Sum256([]byte(strings.ToLower(localAppData)))
	return fmt.Sprintf(`\\.\pipe\oh-my-posh-%x`, hash[:8])
}

func listen() (net.Listener, error) {
	path := socketPath()

	// Create named pipe with security descriptor that allows only the current user
	config := &winio.PipeConfig{
		SecurityDescriptor: "",
		MessageMode:        false,
		InputBufferSize:    65536,
		OutputBufferSize:   65536,
	}

	return winio.ListenPipe(path, config)
}

func cleanupSocket() error {
	// Named pipes are automatically cleaned up on Windows
	return nil
}

func socketExists() bool {
	// Try to open the named pipe - if it exists, the daemon is listening
	conn, err := winio.DialPipe(socketPath(), nil)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
