//go:build !windows

package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServeLoop_RequestPipe validates the fifo transport used by fish: the
// daemon opens the fifo O_RDWR (so it never sees EOF between writers) and
// clients do open-write-close per request.
func TestServeLoop_RequestPipe(t *testing.T) {
	t.Setenv("OMP_CACHE_DIR", t.TempDir())
	pwd := t.TempDir()
	chdirBackToWD(t)

	fifoPath := filepath.Join(t.TempDir(), "omp-serve-test.req")
	require.NoError(t, syscall.Mkfifo(fifoPath, 0o600))

	in, err := openServeInput(fifoPath)
	require.NoError(t, err, "O_RDWR open of a fifo must not fail or block without a writer")

	stdoutR, stdoutW, err := os.Pipe()
	require.NoError(t, err)

	reader := newRecordReader(stdoutR)
	done := make(chan struct{})

	go func() {
		defer close(done)
		runServeLoop(in, stdoutW)
	}()

	t.Cleanup(func() {
		_ = in.Close()
		_ = stdoutW.Close()
		_ = stdoutR.Close()
	})

	writeRequest := func(v any) {
		// open-write-close per request, exactly like fish's `echo $json > fifo`
		f, err := os.OpenFile(fifoPath, os.O_WRONLY, 0)
		require.NoError(t, err)

		data, err := json.Marshal(v)
		require.NoError(t, err)

		_, err = f.Write(append(data, '\n'))
		require.NoError(t, err)
		require.NoError(t, f.Close())
	}

	writeRequest(map[string]any{"command": "render", "id": 1, "shell": "fish", "pwd": pwd})

	records := reader.collect(500 * time.Millisecond)
	require.NotEmpty(t, records, "render request over the fifo must produce records")
	for _, rec := range records {
		assert.Equal(t, "1", rec.id)
	}

	// A second open-write-close must reach the daemon too: its own O_RDWR
	// handle prevents EOF between writers.
	writeRequest(map[string]any{"command": "render", "id": 2, "shell": "fish", "pwd": pwd})

	records = reader.collect(500 * time.Millisecond)
	require.NotEmpty(t, records, "second fifo writer must still reach the daemon")

	writeRequest(map[string]any{"command": "quit"})

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("serve loop did not exit after quit over the fifo")
	}
}
