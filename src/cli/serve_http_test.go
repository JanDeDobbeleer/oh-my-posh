package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// httpHarness wires a serveHTTPServer to a stream-reading client so each
// test only expresses protocol traffic - the HTTP sibling of serveHarness.
type httpHarness struct {
	t        *testing.T
	url      string
	token    string
	reader   *recordReader
	stream   *http.Response
	done     chan struct{}
	rendered bool
}

func startHTTPHarness(t *testing.T) *httpHarness {
	t.Helper()

	t.Setenv("OMP_CACHE_DIR", t.TempDir())

	portFilePath := filepath.Join(t.TempDir(), "omp-serve-test.port")

	server, err := startServeHTTP(portFilePath)
	require.NoError(t, err)

	contents, err := os.ReadFile(portFilePath)
	require.NoError(t, err, "the port file must exist as soon as startServeHTTP returns")

	fields := strings.Fields(string(contents))
	require.Len(t, fields, 2, "port file carries '<port> <token>'")

	h := &httpHarness{
		t:     t,
		url:   "http://127.0.0.1:" + fields[0],
		token: fields[1],
		done:  make(chan struct{}),
	}

	go func() {
		defer close(h.done)
		h.rendered = server.run()
	}()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, h.url+"/stream", nil)
	require.NoError(t, err)
	req.Header.Set(serveTokenHeader, h.token)

	h.stream, err = http.DefaultClient.Do(req) //nolint:bodyclose // closed in t.Cleanup below
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, h.stream.StatusCode)

	h.reader = newRecordReader(h.stream.Body)

	t.Cleanup(func() {
		_ = h.stream.Body.Close()

		select {
		case <-h.done:
		case <-time.After(2 * time.Second):
			t.Error("serve HTTP server did not exit on harness teardown")
		}
	})

	return h
}

// post sends one serveRequest JSON and returns the response status code.
func (h *httpHarness) post(v any) int {
	h.t.Helper()

	data, err := json.Marshal(v)
	require.NoError(h.t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, h.url+"/", strings.NewReader(string(data)))
	require.NoError(h.t, err)
	req.Header.Set(serveTokenHeader, h.token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(h.t, err)
	defer resp.Body.Close()

	return resp.StatusCode
}

func TestServeHTTP_RenderStreamsIDPrefixedRecords(t *testing.T) {
	h := startHTTPHarness(t)
	pwdOne := filepath.Join(t.TempDir(), "first-dir")
	pwdTwo := filepath.Join(t.TempDir(), "second-dir")
	require.NoError(t, os.Mkdir(pwdOne, 0o755))
	require.NoError(t, os.Mkdir(pwdTwo, 0o755))
	chdirBackToWD(t)

	status := h.post(map[string]any{"command": "render", "id": 1, "shell": "nu", "pwd": pwdOne})
	assert.Equal(t, http.StatusOK, status)

	records := h.reader.collect(500 * time.Millisecond)
	require.NotEmpty(t, records, "render over HTTP must produce records on /stream")

	for _, rec := range records {
		assert.Equal(t, "1", rec.id)
	}
	assert.Contains(t, records[0].payload, "first-dir")

	// A second render on a fresh POST must reach the same daemon and follow
	// the directory change.
	status = h.post(map[string]any{"command": "render", "id": 2, "shell": "nu", "pwd": pwdTwo})
	assert.Equal(t, http.StatusOK, status)

	records = h.reader.collect(500 * time.Millisecond)
	require.NotEmpty(t, records, "second render must still reach the daemon")
	assert.Contains(t, records[0].payload, "second-dir")

	status = h.post(map[string]any{"command": "quit"})
	assert.Equal(t, http.StatusOK, status)

	select {
	case <-h.done:
	case <-time.After(2 * time.Second):
		t.Fatal("serve HTTP server did not exit after quit")
	}

	assert.True(t, h.rendered, "at least one render occurred before quit")
}

func TestServeHTTP_RejectsBadToken(t *testing.T) {
	h := startHTTPHarness(t)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, h.url+"/", strings.NewReader(`{"command":"quit"}`))
	require.NoError(t, err)
	req.Header.Set(serveTokenHeader, "wrong-token")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "a wrong token must not drive the daemon")

	h.post(map[string]any{"command": "quit"})
}

// TestServeHTTP_StreamDisconnectShutsDown validates the lifecycle contract:
// the stream connection breaking (shell gone without a quit) must terminate
// the daemon, exactly like SIGPIPE does on the fifo transport.
func TestServeHTTP_StreamDisconnectShutsDown(t *testing.T) {
	h := startHTTPHarness(t)

	require.NoError(t, h.stream.Body.Close())

	select {
	case <-h.done:
	case <-time.After(2 * time.Second):
		t.Fatal("serve HTTP server did not exit when the stream client disconnected")
	}
}

func TestServeHTTP_PortFileRemovedOnExit(t *testing.T) {
	t.Setenv("OMP_CACHE_DIR", t.TempDir())

	portFilePath := filepath.Join(t.TempDir(), "omp-serve-test.port")

	server, err := startServeHTTP(portFilePath)
	require.NoError(t, err)

	contents, err := os.ReadFile(portFilePath)
	require.NoError(t, err)

	fields := strings.Fields(string(contents))
	require.Len(t, fields, 2)

	done := make(chan struct{})
	go func() {
		defer close(done)
		server.run()
	}()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, fmt.Sprintf("http://127.0.0.1:%s/", fields[0]), strings.NewReader(`{"command":"quit"}`))
	require.NoError(t, err)
	req.Header.Set(serveTokenHeader, fields[1])

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("serve HTTP server did not exit after quit")
	}

	assert.NoFileExists(t, portFilePath, "the daemon must clean up its own port file")
}
