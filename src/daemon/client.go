package daemon

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/daemon/ipc"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// DialTimeout is the maximum time to wait for daemon connection.
const DialTimeout = 2 * time.Second

// ResponseTypeComplete indicates the final response from the daemon.
const ResponseTypeComplete = "complete"

// ResponseCallback is called for each response from the daemon.
// Return false to stop receiving responses.
type ResponseCallback func(*ipc.PromptResponse) bool

// Client handles communication with the daemon.
type Client struct {
	conn   *grpc.ClientConn
	client ipc.DaemonServiceClient
}

// NewClient creates a new daemon client.
// Returns an error if the daemon is not running.
func NewClient() (*Client, error) {
	conn, err := ipc.Dial()
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	// gRPC uses lazy connection, so we need to explicitly connect and verify.
	// Connect() initiates connection, then we wait for Ready state.
	ctx, cancel := context.WithTimeout(context.Background(), DialTimeout)
	defer cancel()

	conn.Connect()

	// Wait for connection to become ready
	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			break
		}
		if state == connectivity.TransientFailure || state == connectivity.Shutdown {
			conn.Close()
			return nil, fmt.Errorf("daemon not available: connection state %s", state)
		}
		// Wait for state change or timeout
		if !conn.WaitForStateChange(ctx, state) {
			conn.Close()
			return nil, fmt.Errorf("connection timeout: daemon not responding")
		}
	}

	return &Client{
		conn:   conn,
		client: ipc.NewDaemonServiceClient(conn),
	}, nil
}

// ConnectOrStart attempts to connect to the daemon.
// If connection fails, it kills any stale daemon, calls startFunc to start a new one,
// waits briefly, and retries the connection once.
func ConnectOrStart(startFunc func() error) (*Client, error) {
	client, err := NewClient()
	if err == nil {
		return client, nil
	}

	// Connection failed.
	// 1. Force kill ANY existing daemon/lock (clean slate)
	_ = KillDaemon()

	// 2. Start a fresh daemon
	if err := startFunc(); err != nil {
		return nil, fmt.Errorf("failed to start daemon: %w", err)
	}

	// 3. Wait briefly for startup
	// TODO: Replace with a more robust readiness check if needed,
	// but NewClient already waits for connection readiness.
	// This sleep is just to allow the process to initialize the socket file.
	time.Sleep(50 * time.Millisecond)

	// Attempt 2: Connect again
	client, err = NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon after restart: %w", err)
	}

	return client, nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// RenderPrompt sends a prompt render request to the daemon and streams responses.
//
// The daemon returns responses incrementally:
//   - After async_timeout (100ms): partial prompt with fast segments + cached values for slow ones
//   - As slow segments complete: streamed updates for shell to repaint
//   - Final "complete" response when all segments are done
//
// The callback is invoked for each response. Return false from callback to stop receiving.
// The requestID is automatically generated and used to filter stale responses.
func (c *Client) RenderPrompt(ctx context.Context, flags *runtime.Flags, pid int, sessionID string, env map[string]string, callback ResponseCallback) error {
	requestID := uuid.NewString()

	// Use PID as SessionID if available to ensure accurate session tracking and cleanup in daemon
	if pid > 0 {
		sessionID = fmt.Sprint(pid)
	} else if sessionID == "" {
		sessionID = getSessionID()
	}

	req := &ipc.PromptRequest{
		Version:   ipc.ProtocolVersion,
		SessionId: sessionID,
		RequestId: requestID,
		Pid:       int32(pid),
		Env:       env,
		Flags:     ipc.FlagsToProto(flags),
	}

	log.Debugf("Sending prompt request: session=%s, request=%s, type=%s", sessionID, requestID, flags.Type)

	// NOTE: c.client.RenderPrompt is the gRPC-generated method on ipc.DaemonServiceClient,
	// not a recursive call to this method. The gRPC client is stored in c.client.
	stream, err := c.client.RenderPrompt(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send render request: %w", err)
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("stream error: %w", err)
		}

		// Filter stale responses from previous requests (different request ID)
		if resp.RequestId != requestID {
			log.Debugf("Ignoring stale response: got %s, expected %s", resp.RequestId, requestID)
			continue
		}

		if resp.Error != "" {
			return fmt.Errorf("daemon error: %s", resp.Error)
		}

		if !callback(resp) {
			return nil
		}

		if resp.Type == ResponseTypeComplete {
			return nil
		}
	}
}

// RenderPromptSync sends a prompt render request and waits for the complete response.
// This is a convenience wrapper for cases that don't need incremental updates.
func (c *Client) RenderPromptSync(ctx context.Context, flags *runtime.Flags, pid int, sessionID string, env map[string]string) (*ipc.PromptResponse, error) {
	var finalResponse *ipc.PromptResponse

	err := c.RenderPrompt(ctx, flags, pid, sessionID, env, func(resp *ipc.PromptResponse) bool {
		finalResponse = resp
		return resp.Type != ResponseTypeComplete
	})
	if err != nil {
		return nil, err
	}

	return finalResponse, nil
}

// ToggleSegment toggles segments in the daemon.
func (c *Client) ToggleSegment(ctx context.Context, pid int, segments []string) error {
	var sessionID string
	if pid > 0 {
		sessionID = fmt.Sprint(pid)
	} else {
		sessionID = getSessionID()
	}

	req := &ipc.ToggleSegmentRequest{
		SessionId: sessionID,
		Segments:  segments,
	}

	resp, err := c.client.ToggleSegment(ctx, req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}

	return nil
}

// CacheClear clears all daemon cache entries.
func (c *Client) CacheClear(ctx context.Context) error {
	resp, err := c.client.CacheClear(ctx, &ipc.CacheClearRequest{})
	if err != nil {
		return err
	}
	if !resp.Success && resp.Error != "" {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

// CacheSetTTL sets the default cache TTL (in days).
func (c *Client) CacheSetTTL(ctx context.Context, days int) error {
	_, err := c.client.CacheSetTTL(ctx, &ipc.CacheSetTTLRequest{Days: int32(days)})
	return err
}

// CacheGetTTL gets the current default cache TTL (in days).
func (c *Client) CacheGetTTL(ctx context.Context) (int, error) {
	resp, err := c.client.CacheGetTTL(ctx, &ipc.CacheGetTTLRequest{})
	if err != nil {
		return 0, err
	}
	return int(resp.Days), nil
}

// SetLogging enables or disables file logging on the running daemon.
func (c *Client) SetLogging(ctx context.Context, path string) error {
	resp, err := c.client.SetLogging(ctx, &ipc.SetLoggingRequest{Path: path})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

// getSessionID returns the session ID from environment or cache.
func getSessionID() string {
	if id := os.Getenv("POSH_SESSION_ID"); id != "" {
		return id
	}
	return cache.SessionID()
}

// IsRunning checks if the daemon is currently running.
func IsRunning() bool {
	client, err := NewClient()
	if err != nil {
		return false
	}
	client.Close()
	return true
}

// PromptResult contains the rendered prompts from a daemon response.
type PromptResult struct {
	Primary   string
	Right     string
	Secondary string
	Transient string
	Debug     string
	Tooltip   string
	Valid     string
	Error     string
}

// ExtractPrompts converts a PromptResponse into a PromptResult.
func ExtractPrompts(resp *ipc.PromptResponse) *PromptResult {
	result := &PromptResult{}
	if resp == nil || resp.Prompts == nil {
		return result
	}

	if p, ok := resp.Prompts["primary"]; ok {
		result.Primary = p.Text
	}
	if p, ok := resp.Prompts["right"]; ok {
		result.Right = p.Text
	}
	if p, ok := resp.Prompts["secondary"]; ok {
		result.Secondary = p.Text
	}
	if p, ok := resp.Prompts["transient"]; ok {
		result.Transient = p.Text
	}
	if p, ok := resp.Prompts["debug"]; ok {
		result.Debug = p.Text
	}
	if p, ok := resp.Prompts["tooltip"]; ok {
		result.Tooltip = p.Text
	}
	if p, ok := resp.Prompts["valid"]; ok {
		result.Valid = p.Text
	}
	if p, ok := resp.Prompts["error"]; ok {
		result.Error = p.Text
	}

	return result
}
