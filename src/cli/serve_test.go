package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// serveHarness wires a runServeLoop instance to pipes so each test only
// expresses protocol traffic. Construct with startServeHarness.
type serveHarness struct {
	t        *testing.T
	stdin    *os.File
	reader   *recordReader
	done     chan struct{}
	rendered bool
}

func startServeHarness(t *testing.T) *serveHarness {
	t.Helper()

	t.Setenv("OMP_CACHE_DIR", t.TempDir())

	stdinR, stdinW, err := os.Pipe()
	require.NoError(t, err)

	stdoutR, stdoutW, err := os.Pipe()
	require.NoError(t, err)

	h := &serveHarness{
		t:      t,
		stdin:  stdinW,
		reader: newRecordReader(stdoutR),
		done:   make(chan struct{}),
	}

	go func() {
		defer close(h.done)
		h.rendered = runServeLoop(stdinR, stdoutW)
	}()

	t.Cleanup(func() {
		_ = stdinW.Close()
		_ = stdoutW.Close()
		_ = stdoutR.Close()
	})

	return h
}

// send writes a single newline-terminated JSON request to the loop's stdin.
func (h *serveHarness) send(v any) {
	h.t.Helper()

	data, err := json.Marshal(v)
	require.NoError(h.t, err)

	_, err = h.stdin.Write(append(data, '\n'))
	require.NoError(h.t, err)
}

func (h *serveHarness) render(id int, pwd string) {
	h.send(map[string]any{"command": "render", "id": id, "shell": "pwsh", "pwd": pwd})
}

func (h *serveHarness) records(timeout time.Duration) []serveRecord {
	return h.reader.collect(timeout)
}

// quitAndWait sends the quit command and fails the test when the loop does
// not exit in time.
func (h *serveHarness) quitAndWait() {
	h.t.Helper()

	h.send(map[string]any{"command": "quit"})

	select {
	case <-h.done:
	case <-time.After(2 * time.Second):
		h.t.Fatal("serve loop did not exit after quit")
	}
}

// chdirBackToWD restores the process's current working directory once the
// test completes. startRenderCycle calls os.Chdir(pwd) for each render
// request (mirroring the real daemon), which - on Windows - keeps the
// directory handle open and blocks t.TempDir()'s cleanup if the process is
// still sitting inside it. Cleanup functions run in LIFO order, so this must
// be called AFTER every t.TempDir() call in the test (i.e. registered last),
// so it runs FIRST and moves the process out of any temp dir before
// t.TempDir()'s own removal cleanup runs.
func chdirBackToWD(t *testing.T) {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
}

// serveRecord is one parsed NUL-delimited protocol record from serve's stdout.
type serveRecord struct {
	id        string
	payload   string
	transient bool
}

// recordReader parses NUL-delimited protocol records off a pipe. One reader
// per pipe for its whole lifetime: a bufio.Scanner buffers past record
// boundaries, so a second scanner on the same pipe would lose data.
type recordReader struct {
	ch chan serveRecord
}

func newRecordReader(r *os.File) *recordReader {
	rr := &recordReader{ch: make(chan serveRecord, 64)}

	go func() {
		defer close(rr.ch)

		scanner := bufio.NewScanner(r)
		// Match the daemon's request scanner: a record carries a full prompt
		// payload, which can exceed bufio's 64 KB default token size.
		scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)
		scanner.Split(splitOnNull)

		for scanner.Scan() {
			raw := scanner.Text()

			idPart, payload, found := strings.Cut(raw, serveIDMarker)
			if !found {
				continue
			}

			rec := serveRecord{id: idPart, payload: payload}
			if strings.HasPrefix(payload, "\x1e") {
				rec.transient = true
				rec.payload = strings.TrimPrefix(payload, "\x1e")
			}

			rr.ch <- rec
		}
	}()

	return rr
}

// collect returns records until timeout elapses with no new record arriving,
// or the pipe closes.
func (rr *recordReader) collect(timeout time.Duration) []serveRecord {
	var records []serveRecord
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case rec, ok := <-rr.ch:
			if !ok {
				return records
			}
			records = append(records, rec)
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(timeout)
		case <-timer.C:
			return records
		}
	}
}

// splitOnNull is a bufio.SplitFunc that splits on the \x00 record delimiter
// used by the serve/stream protocols.
func splitOnNull(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, 0); i >= 0 {
		return i + 1, data[:i], nil
	}

	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}

func TestServeCommand_Creation(t *testing.T) {
	cmd := createServeCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "serve", cmd.Use)
	assert.True(t, cmd.Hidden, "serve command should be hidden from help")
}

func TestServeLoop_RenderProducesIDPrefixedRecords(t *testing.T) {
	h := startServeHarness(t)
	pwd := t.TempDir()
	chdirBackToWD(t)

	h.render(1, pwd)

	records := h.records(500 * time.Millisecond)
	require.NotEmpty(t, records, "expected at least one record for cycle 1")

	for _, rec := range records {
		assert.Equal(t, "1", rec.id, "every record in this cycle should carry the id from the request")
	}

	// The last record of a cycle with no pending segments should be the
	// transient record (refreshed once all segments resolved).
	assert.True(t, records[len(records)-1].transient, "final record of a completed cycle should be the transient record")

	h.quitAndWait()

	assert.True(t, h.rendered, "at least one render occurred before quit")
}

func TestServeLoop_AbortStopsRecordFlowThenNewRenderWorks(t *testing.T) {
	h := startServeHarness(t)
	pwd := t.TempDir()
	chdirBackToWD(t)

	h.render(1, pwd)

	// Let cycle 1 start producing, then abort it immediately.
	h.send(map[string]any{"command": "abort"})

	h.render(2, pwd)

	records := h.records(500 * time.Millisecond)
	require.NotEmpty(t, records, "expected records for cycle 2 after abort+re-render")

	// No record from cycle 1 should appear once cycle 2 begins - cycle 1 had
	// no pending segments so it may have fully completed before the abort
	// landed, but every record we DO see for id 2 must never be interleaved
	// with a id-1 record following it.
	seenTwo := false
	for _, rec := range records {
		if rec.id == "2" {
			seenTwo = true
		}
		if seenTwo {
			assert.Equal(t, "2", rec.id, "no cycle 1 record should arrive once cycle 2 records begin")
		}
	}
	assert.True(t, seenTwo, "expected to see cycle 2 records")

	h.quitAndWait()
}

func TestServeLoop_WaitRenderEmitsExactlyTwoRecords(t *testing.T) {
	h := startServeHarness(t)
	pwd := t.TempDir()
	chdirBackToWD(t)

	h.send(map[string]any{"command": "render", "id": 1, "shell": "bash", "pwd": pwd, "wait": true})

	records := h.records(500 * time.Millisecond)
	require.Len(t, records, 2, "a wait render emits exactly the final primary and the transient")
	assert.Equal(t, "1", records[0].id)
	assert.False(t, records[0].transient, "first record is the fully resolved primary")
	assert.True(t, records[1].transient, "second record is the transient")
	assert.NotEmpty(t, records[0].payload)
	assert.NotEmpty(t, records[1].payload)

	// A regular streaming render must still work after a wait render.
	h.render(2, pwd)
	records = h.records(500 * time.Millisecond)
	require.NotEmpty(t, records, "streaming render after a wait render must still produce records")
	assert.Equal(t, "2", records[0].id)

	h.quitAndWait()
}

// TestServeLoop_EnvOverlayUnsetsVanishedVariables guards against stale env
// pinning: a variable present in one request's overlay but absent from the
// next (e.g. VIRTUAL_ENV after `deactivate`) must be unset in the daemon,
// not keep its old value forever.
func TestServeLoop_EnvOverlayUnsetsVanishedVariables(t *testing.T) {
	h := startServeHarness(t)
	pwd := t.TempDir()
	chdirBackToWD(t)

	const name = "POSH_SERVE_ENV_TEST"
	t.Cleanup(func() { _ = os.Unsetenv(name) })

	h.send(map[string]any{
		"command": "render", "id": 1, "shell": "pwsh", "pwd": pwd,
		"env": map[string]string{name: "venv-active"},
	})
	records := h.records(500 * time.Millisecond)
	require.NotEmpty(t, records)
	assert.Equal(t, "venv-active", os.Getenv(name), "overlay variable applied")

	h.send(map[string]any{
		"command": "render", "id": 2, "shell": "pwsh", "pwd": pwd,
		"env": map[string]string{},
	})
	records = h.records(500 * time.Millisecond)
	require.NotEmpty(t, records)
	assert.Empty(t, os.Getenv(name), "vanished overlay variable must be unset")

	h.quitAndWait()
}

func TestServeLoop_QuitExitsCleanly(t *testing.T) {
	h := startServeHarness(t)

	h.quitAndWait()

	// Regression guard: quitting before any render must report "no render
	// happened" so the caller (createServeCmd) knows NOT to call
	// template.SaveCache(), which panics on template package state that's
	// only initialized by a render (template.Init runs inside prompt.New).
	assert.False(t, h.rendered, "no render occurred before quit")
}

func TestServeLoop_EOFExitsCleanly(t *testing.T) {
	h := startServeHarness(t)
	pwd := t.TempDir()
	chdirBackToWD(t)

	h.render(1, pwd)

	_ = h.records(300 * time.Millisecond)

	// Closing stdin (EOF) must make the loop exit even without an explicit quit.
	require.NoError(t, h.stdin.Close())

	select {
	case <-h.done:
	case <-time.After(2 * time.Second):
		t.Fatal("serve loop did not exit on stdin EOF")
	}
}

func TestServeLoop_UnknownCommandIgnored(t *testing.T) {
	h := startServeHarness(t)
	pwd := t.TempDir()
	chdirBackToWD(t)

	h.send(map[string]any{"command": "reload"})
	h.send(map[string]any{"unknown-field": "value"})

	h.render(1, pwd)

	records := h.records(500 * time.Millisecond)
	assert.NotEmpty(t, records, "serve should still work after unknown commands/fields")

	h.quitAndWait()
}

// TestServeLoop_UTF8BOMOnFirstLine validates that a UTF-8 BOM prefixing the
// very first request line (written by .NET's default UTF8 StreamWriter
// encoding on its first write) does not make the daemon drop the request.
func TestServeLoop_UTF8BOMOnFirstLine(t *testing.T) {
	h := startServeHarness(t)
	pwd := t.TempDir()
	chdirBackToWD(t)

	// Prefix the first line with a UTF-8 BOM, exactly like a .NET
	// StreamWriter with Encoding.UTF8 does on its first write.
	data, err := json.Marshal(map[string]any{
		"command": "render",
		"id":      1,
		"shell":   "pwsh",
		"pwd":     pwd,
	})
	require.NoError(t, err)

	payload := append([]byte{0xEF, 0xBB, 0xBF}, data...)
	payload = append(payload, '\n')
	_, err = h.stdin.Write(payload)
	require.NoError(t, err)

	records := h.records(500 * time.Millisecond)
	require.NotEmpty(t, records, "a BOM-prefixed first request must not be dropped")

	for _, rec := range records {
		assert.Equal(t, "1", rec.id)
	}

	h.quitAndWait()
}
