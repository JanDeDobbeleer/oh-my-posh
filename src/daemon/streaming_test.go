package daemon

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/daemon/ipc"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Segment that sleeps for 2 seconds
type slowSegment2s struct {
	segments.Base
}

func (s *slowSegment2s) Template() string { return "slow2s" }
func (s *slowSegment2s) Enabled() bool {
	time.Sleep(2 * time.Second)
	return true
}

// Segment that sleeps for 5 seconds
type slowSegment5s struct {
	segments.Base
}

func (s *slowSegment5s) Template() string { return "slow5s" }
func (s *slowSegment5s) Enabled() bool {
	time.Sleep(5 * time.Second)
	return true
}

func init() {
	config.Segments["slow2s"] = func() config.SegmentWriter { return &slowSegment2s{} }
	config.Segments["slow5s"] = func() config.SegmentWriter { return &slowSegment5s{} }
}

func createStreamingTestConfig(t *testing.T) string {
	t.Helper()
	content := `{
		"version": 4,
		"daemon_timeout": 200,
		"blocks": [
			{
				"type": "prompt",
				"alignment": "left",
				"segments": [
					{
						"type": "text",
						"style": "plain",
						"template": "fast1 "
					},
					{
						"type": "slow2s",
						"style": "plain",
						"template": " {{ .Text }} "
					},
					{
						"type": "text",
						"style": "plain",
						"template": "fast2 "
					},
					{
						"type": "slow5s",
						"style": "plain",
						"template": " {{ .Text }} "
					}
				]
			}
		]
	}`
	tmpFile, err := os.CreateTemp("", "omp-streaming-*.json")
	require.NoError(t, err)
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })
	return tmpFile.Name()
}

// TestStreaming_PartialThenUpdates verifies the daemon streams partial results:
//
//	Response 1 (< 500ms): initial render with pending segments (type=update)
//	Response 2 (~2s):     slow2s completes (type=update)
//	Response 3 (~5s):     slow5s completes (type=update, from callback)
//	Response 4 (~5s):     final re-render  (type=complete)
func TestStreaming_PartialThenUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping streaming test in short mode")
	}

	tmpDir := testSocketDir(t)
	setTestEnv(t, tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	configPath := createStreamingTestConfig(t)
	d, err := New(configPath)
	require.NoError(t, err)

	go func() { _ = d.Start() }()
	time.Sleep(100 * time.Millisecond)

	defer func() {
		d.shutdown()
		<-d.Done()
	}()

	conn, err := ipc.Dial()
	require.NoError(t, err)
	defer conn.Close()

	client := ipc.NewDaemonServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	flags := &runtime.Flags{
		ConfigPath: configPath,
		PWD:        "/tmp",
		Shell:      "bash",
		Type:       "primary",
		IsPrimary:  true,
	}

	start := time.Now()

	stream, err := client.RenderPrompt(ctx, &ipc.PromptRequest{
		Version:   ipc.ProtocolVersion,
		SessionId: "streaming-test",
		RequestId: "req-1",
		Flags:     ipc.FlagsToProto(flags),
	})
	require.NoError(t, err)

	type response struct {
		resp *ipc.PromptResponse
		at   time.Duration
	}

	var responses []response
	for {
		resp, err := stream.Recv()
		if err != nil {
			break
		}
		responses = append(responses, response{resp: resp, at: time.Since(start)})
		if resp.Type == "complete" {
			break
		}
	}

	// We expect 4 responses:
	//   1. Initial partial render (type=update, <500ms)
	//   2. slow2s completes (type=update, ~2s)
	//   3. slow5s completes (type=update, ~5s)  — from updateCallback
	//   4. Final complete (type=complete, ~5s)   — from RenderPrompt's pendingDone path
	require.Len(t, responses, 4, "expected 4 streaming responses")

	// Response 1: initial partial render, must arrive before 500ms
	r1 := responses[0]
	assert.Equal(t, "update", r1.resp.Type, "first response should be an update (pending segments)")
	assert.Less(t, r1.at, 500*time.Millisecond, "initial response must arrive within 500ms")

	// Response 2: slow2s completed, should arrive around 2s (allow 1s–3.5s)
	r2 := responses[1]
	assert.Equal(t, "update", r2.resp.Type, "second response should be an update")
	assert.Greater(t, r2.at, 1*time.Second, "second response should not arrive before 1s")
	assert.Less(t, r2.at, 3500*time.Millisecond, "second response should arrive before 3.5s")

	// Response 3: slow5s completes via callback (~5s)
	r3 := responses[2]
	assert.Equal(t, "update", r3.resp.Type, "third response should be an update")
	assert.Greater(t, r3.at, 4*time.Second, "third response should not arrive before 4s")
	assert.Less(t, r3.at, 7*time.Second, "third response should arrive before 7s")

	// Response 4: final complete, immediately after slow5s
	r4 := responses[3]
	assert.Equal(t, "complete", r4.resp.Type, "final response should be complete")
	assert.Greater(t, r4.at, 4*time.Second, "final response should not arrive before 4s")
	assert.Less(t, r4.at, 7*time.Second, "final response should arrive before 7s")

	// All responses should contain a primary prompt
	for i, r := range responses {
		assert.Contains(t, r.resp.Prompts, "primary", "response %d should contain primary prompt", i)
	}
}
