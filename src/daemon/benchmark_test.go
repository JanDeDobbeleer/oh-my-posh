package daemon

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/daemon/ipc"
	"github.com/jandedobbeleer/oh-my-posh/src/prompt"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"

	"github.com/stretchr/testify/require"
)

// Sleep segment implementation
type Sleep struct {
	segments.Base
}

func (s *Sleep) Template() string {
	return "slow"
}

func (s *Sleep) Enabled() bool {
	time.Sleep(200 * time.Millisecond)
	return true
}

func init() {
	// Register the sleep segment
	config.Segments["sleep"] = func() config.SegmentWriter {
		return &Sleep{}
	}
}

// createSlowConfigDefault creates a config with a slow segment and NO explicit timeout
func createSlowConfigDefault(t testing.TB) string {
	t.Helper()
	content := `{
		"version": 4,
		"blocks": [
			{
				"type": "prompt",
				"alignment": "left",
				"segments": [
					{
						"type": "text",
						"style": "plain",
						"template": "fast "
					},
					{
						"type": "sleep",
						"style": "plain",
						"template": " {{ .Text }} "
					}
				]
			}
		]
	}`

	tmpFile, err := os.CreateTemp("", "omp-slow-default-*.json")
	require.NoError(t, err)
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })
	return tmpFile.Name()
}

// benchmarkDaemonWithConfig runs a daemon benchmark with the given config path.
func benchmarkDaemonWithConfig(b *testing.B, configPath string) {
	b.Helper()

	dir, err := os.MkdirTemp("", "omp-bench")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(dir)

	os.Setenv("XDG_STATE_HOME", dir)
	os.Setenv("XDG_RUNTIME_DIR", dir)

	d, err := New(configPath)
	if err != nil {
		b.Fatal(err)
	}

	go func() {
		_ = d.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	defer func() {
		d.shutdown()
		<-d.Done()
	}()

	client, err := NewClient()
	if err != nil {
		b.Fatal(err)
	}
	defer client.Close()

	flags := &runtime.Flags{
		ConfigPath: configPath,
		PWD:        "/tmp",
		Shell:      "bash",
		Type:       "primary",
		IsPrimary:  true,
	}

	for b.Loop() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		err := client.RenderPrompt(ctx, flags, 0, "", nil, func(_ *ipc.PromptResponse) bool {
			return false
		})
		cancel()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDaemonDefault measures the time to receive the FIRST response from the daemon using the DEFAULT 100ms timeout.
func BenchmarkDaemonDefault(b *testing.B) {
	benchmarkDaemonWithConfig(b, createSlowConfigDefault(b))
}

func createSlowConfig(t testing.TB) string {
	t.Helper()
	content := `{
		"version": 4,
		"daemon_timeout": 10,
		"blocks": [
			{
				"type": "prompt",
				"alignment": "left",
				"segments": [
					{
						"type": "text",
						"style": "plain",
						"template": "fast "
					},
					{
						"type": "sleep",
						"style": "plain",
						"template": " {{ .Text }} "
					}
				]
			}
		]
	}`

	tmpFile, err := os.CreateTemp("", "omp-slow-*.json")
	require.NoError(t, err)
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })
	return tmpFile.Name()
}

// BenchmarkDirect measures the time to render a slow prompt directly (blocking).
func BenchmarkDirect(b *testing.B) {
	configPath := createSlowConfig(b)
	flags := &runtime.Flags{
		ConfigPath: configPath,
		PWD:        "/tmp",
		Shell:      "bash",
		Type:       "primary",
		IsPrimary:  true,
	}

	for b.Loop() {
		eng := prompt.New(flags)
		_ = eng.Primary()
	}
}

// BenchmarkDaemon measures the time to receive the FIRST response from the daemon.
func BenchmarkDaemon(b *testing.B) {
	benchmarkDaemonWithConfig(b, createSlowConfig(b))
}
