package cli

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

// The HTTP transport exists for Windows nu: mkfifo doesn't exist there and
// nu's write primitives (save --append/--force, o> redirects) all fail
// against Windows named pipes, but its built-in http client both fires
// one-shot POSTs and consumes chunked responses incrementally. Requests
// arrive as single POSTs on "/", and every prompt record flows over one
// long-lived "/stream" response the shell's reader job holds open for the
// daemon's whole lifetime - the moral equivalent of the fifo pair, with the
// stream disconnect replacing SIGPIPE as the shell-disappeared signal.

// serveTokenHeader carries the shared secret published in the port file.
// The listener is loopback-only, but any local process can connect to it -
// the token keeps other users' processes (and browsers doing DNS-rebinding
// tricks) from driving the daemon.
const serveTokenHeader = "X-Omp-Token"

// serveStreamTimeout bounds how long the daemon waits for the shell's
// reader job to connect to /stream - at startup (a daemon nobody ever
// connects to must not linger as an orphan) and per render request.
const serveStreamTimeout = 30 * time.Second

type serveHTTPServer struct {
	listener net.Listener
	token    string
	portFile string

	// streamReady is closed once the stream client connected; stream is
	// written exactly once before that and never mutated after.
	streamReady chan struct{}
	stream      io.Writer

	done     chan struct{}
	doneOnce sync.Once

	mu       sync.Mutex
	active   *serveActiveCycle
	envKeys  map[string]struct{}
	rendered bool
}

// startServeHTTP binds a loopback listener on an ephemeral port and
// publishes "<port> <token>" to portFile so the shell can find it. The
// write is atomic (temp file + rename): the shell polls the file and must
// never observe a partial write.
func startServeHTTP(portFile string) (*serveHTTPServer, error) {
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	raw := make([]byte, 16)
	if _, err = rand.Read(raw); err != nil {
		listener.Close()
		return nil, err
	}

	server := &serveHTTPServer{
		listener:    listener,
		token:       hex.EncodeToString(raw),
		portFile:    portFile,
		streamReady: make(chan struct{}),
		done:        make(chan struct{}),
		envKeys:     map[string]struct{}{},
	}

	port := listener.Addr().(*net.TCPAddr).Port
	temp := portFile + ".tmp"

	if err = os.WriteFile(temp, fmt.Appendf(nil, "%d %s\n", port, server.token), 0o600); err != nil {
		listener.Close()
		return nil, err
	}

	if err = os.Rename(temp, portFile); err != nil {
		_ = os.Remove(temp)
		listener.Close()
		return nil, err
	}

	return server, nil
}

// run serves until quit or stream disconnect and reports whether at least
// one render request was handled (same contract as runServeLoop).
func (s *serveHTTPServer) run() bool {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", s.handleRequest)
	mux.HandleFunc("GET /stream", s.handleStream)

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		_ = server.Serve(s.listener)
	}()

	// A daemon whose stream client never shows up (the shell died between
	// spawning it and connecting the reader job) must not linger forever.
	go func() {
		select {
		case <-s.streamReady:
		case <-s.done:
		case <-time.After(serveStreamTimeout):
			s.shutdown()
		}
	}()

	<-s.done

	s.mu.Lock()
	s.stopActiveLocked()
	rendered := s.rendered
	s.mu.Unlock()

	_ = server.Close()
	_ = os.Remove(s.portFile)

	return rendered
}

func (s *serveHTTPServer) shutdown() {
	s.doneOnce.Do(func() { close(s.done) })
}

func (s *serveHTTPServer) authorized(r *http.Request) bool {
	token := r.Header.Get(serveTokenHeader)
	return subtle.ConstantTimeCompare([]byte(token), []byte(s.token)) == 1
}

// stopActiveLocked mirrors runServeLoop's stopActiveCycle; see the
// serveActiveCycle doc for why the previous cycle must be fully stopped
// (Abort returned, copier drained) before the next one may start.
func (s *serveHTTPServer) stopActiveLocked() {
	if s.active == nil {
		return
	}

	s.active.engine.Abort()
	<-s.active.copierDone
	s.active = nil
}

// handleStream is the single long-lived record sink: every cycle's records
// are written (and flushed) to this response. The shell's reader job holds
// it open for the daemon's lifetime, so the connection breaking means the
// shell is gone - the daemon's cue to exit, exactly like SIGPIPE on the
// fifo transport.
func (s *serveHTTPServer) handleStream(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	s.mu.Lock()
	if s.stream != nil {
		s.mu.Unlock()
		http.Error(w, "stream already connected", http.StatusConflict)
		return
	}
	s.stream = &flushWriter{writer: w, flusher: flusher}
	s.mu.Unlock()

	close(s.streamReady)

	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	select {
	case <-r.Context().Done():
		s.shutdown()
	case <-s.done:
	}
}

// handleRequest accepts one serveRequest JSON per POST body - the protocol
// of runServeLoop with the newline-delimited stdin stream replaced by
// one-shot requests. Responses carry no payload; render records flow over
// /stream.
func (s *serveHTTPServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1024*1024))
	if err != nil {
		http.Error(w, "unreadable body", http.StatusBadRequest)
		return
	}

	body = bytes.TrimPrefix(body, []byte{0xEF, 0xBB, 0xBF})

	var req serveRequest
	if err := json.Unmarshal(body, &req); err != nil {
		// Malformed request: ignore for forward/backward compatibility,
		// like runServeLoop does with malformed lines.
		w.WriteHeader(http.StatusOK)
		return
	}

	switch req.Command {
	case serveCommandRender:
		// Records have nowhere to go until the reader job is on /stream;
		// the connect happens concurrently with the first render request,
		// so wait for it instead of dropping the cycle.
		select {
		case <-s.streamReady:
		case <-s.done:
			http.Error(w, "shutting down", http.StatusServiceUnavailable)
			return
		case <-time.After(serveStreamTimeout):
			http.Error(w, "no stream client", http.StatusServiceUnavailable)
			return
		}

		s.mu.Lock()
		s.stopActiveLocked()
		if s.active = startRenderCycle(&req, s.stream, s.envKeys); s.active != nil {
			s.rendered = true
		}
		s.mu.Unlock()

		w.WriteHeader(http.StatusOK)
	case serveCommandAbort:
		s.mu.Lock()
		s.stopActiveLocked()
		s.mu.Unlock()

		w.WriteHeader(http.StatusOK)
	case serveCommandQuit:
		s.mu.Lock()
		s.stopActiveLocked()
		s.mu.Unlock()

		w.WriteHeader(http.StatusOK)
		s.shutdown()
	default:
		// Unknown command: ignore for forward compatibility.
		w.WriteHeader(http.StatusOK)
	}
}

// flushWriter flushes after every write so each record leaves the daemon
// immediately - nu's http client surfaces chunks as they arrive, which is
// what makes streamed segment updates paint live.
type flushWriter struct {
	writer  io.Writer
	flusher http.Flusher
}

func (fw *flushWriter) Write(p []byte) (int, error) {
	n, err := fw.writer.Write(p)
	fw.flusher.Flush()
	return n, err
}
