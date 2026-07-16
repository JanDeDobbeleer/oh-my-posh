package harness

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aymanbagabas/go-pty"
	"github.com/hinshun/vt10x"
)

// promptRegexp matches the base config's prompt template "E2E:<exit code>>", used by
// WaitForPrompt to detect that a shell session is ready for input.
var promptRegexp = regexp.MustCompile(`E2E:\d+>`)

const (
	ptyCols = 120
	ptyRows = 30

	waitForPollInterval = 50 * time.Millisecond
	waitForDeadline     = 30 * time.Second
	exitDeadline        = 15 * time.Second
)

// Session drives one interactive shell process in a pseudo-terminal. A single goroutine
// (started by Start) reads pty output and feeds it to both a vt10x terminal, for rendered-
// screen assertions, and a raw strings.Builder, for escape-sequence assertions; both are
// guarded by mu.
type Session struct {
	t    *testing.T
	pty  pty.Pty
	cmd  *pty.Cmd
	term vt10x.Terminal

	mu  sync.Mutex
	raw strings.Builder

	writeMu sync.Mutex

	waitMu      sync.Mutex
	waitStarted bool
}

// Start launches sh interactively with cfgPath as its oh-my-posh config, in a 120x30 pty.
// It skips the test if sh's binary is missing or unsupported on the current platform,
// otherwise it fails the test on any setup error - unless sh.Name is listed in
// OMP_E2E_REQUIRE, in which case an unavailable shell fails the test instead (see
// SkipUnavailable). The returned Session's process and pty are killed and closed, and the
// process reaped, via t.Cleanup.
func Start(t *testing.T, sh ShellDef, cfgPath string) *Session {
	t.Helper()

	if !sh.SupportedOnHost() {
		SkipUnavailable(t, sh.Name, fmt.Sprintf("%s is not supported on this platform, skipping", sh.Name))
	}

	binPath, err := LookupShellBinary(sh.Binary)
	if err != nil {
		SkipUnavailable(t, sh.Name, fmt.Sprintf("%s not found on PATH, skipping", sh.Binary))
	}

	workDir := t.TempDir()
	script := InitScript(t, sh.Name, cfgPath)
	scriptPath := WriteScript(t, sh.Name, script)

	_, args, env := sh.Launch(t, scriptPath, workDir)
	env = append(env, "TERM=xterm-256color", "OMP_CACHE_DIR="+t.TempDir())

	p, err := pty.New()
	if err != nil {
		t.Fatalf("creating pty: %v", err)
	}

	if err := p.Resize(ptyCols, ptyRows); err != nil {
		_ = p.Close()
		t.Fatalf("resizing pty: %v", err)
	}

	cmd := p.Command(binPath, args...)
	cmd.Dir = workDir
	cmd.Env = env

	if err := cmd.Start(); err != nil {
		_ = p.Close()
		t.Fatalf("starting %s: %v", sh.Name, err)
	}

	s := &Session{
		t:    t,
		pty:  p,
		cmd:  cmd,
		term: vt10x.New(vt10x.WithSize(ptyCols, ptyRows)),
	}

	go s.read()

	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = p.Close()

		// Only reap here if ExpectExit never claimed that responsibility: on its timeout
		// path ExpectExit's own pending cmd.Wait() goroutine is released by the Kill above
		// and reaps the process, so waiting again here would race a second Wait call
		// against it.
		if s.takeWaitOwnership() {
			_ = cmd.Wait()
		}
	})

	return s
}

// takeWaitOwnership reports whether the caller is the first to claim responsibility for
// reaping the process via cmd.Wait(), so exactly one of ExpectExit or the Cleanup
// registered by Start ever calls it. Later callers get false and must not call Wait.
func (s *Session) takeWaitOwnership() bool {
	s.waitMu.Lock()
	defer s.waitMu.Unlock()

	if s.waitStarted {
		return false
	}

	s.waitStarted = true

	return true
}

// dsrRequest is the Device Status Report cursor-position query (CSI 6n). PSReadLine on
// Linux issues it repeatedly and blocks on the reply; ConPTY answers it internally on
// Windows, but on a raw Unix pty the "terminal" — this harness — must respond itself.
var dsrRequest = []byte("\x1b[6n")

// read copies pty output into both the vt10x terminal and the raw buffer until the pty is
// closed, answering DSR cursor-position queries along the way. It is the session's single
// reader goroutine.
func (s *Session) read() {
	buf := make([]byte, 4096)

	// carry holds a tail of the previous chunk that might be the start of a query split
	// across two reads; see splitDSRQueries.
	var carry []byte

	for {
		n, err := s.pty.Read(buf)

		if n == 0 && err == nil {
			// io.Reader permits returning (0, nil); without this backstop such a reader
			// would spin this loop at 100% CPU. This is a defensive backstop, not a
			// pacing mechanism - real ptys never take this path.
			time.Sleep(time.Millisecond)
			continue
		}

		if n > 0 {
			window := make([]byte, 0, len(carry)+n)
			window = append(window, carry...)
			window = append(window, buf[:n]...)

			queries, rest, newCarry := splitDSRQueries(window)

			for _, query := range queries {
				s.mu.Lock()
				s.raw.Write(query)
				_, _ = s.term.Write(query)
				cursor := s.term.Cursor()
				s.mu.Unlock()

				s.writeMu.Lock()
				_, _ = fmt.Fprintf(s.pty, "\x1b[%d;%dR", cursor.Y+1, cursor.X+1)
				s.writeMu.Unlock()
			}

			if len(rest) > 0 {
				s.mu.Lock()
				s.raw.Write(rest)
				_, _ = s.term.Write(rest)
				s.mu.Unlock()
			}

			carry = append([]byte(nil), newCarry...)
		}

		if err != nil {
			return
		}
	}
}

// splitDSRQueries scans data - the bytes carried over from the previous read plus the
// newly read chunk - for CSI 6n cursor-position queries (dsrRequest). It returns, in
// order, one entry per complete query found (each holding every byte since the end of the
// previous query through the end of this one), the non-query bytes following the last
// query, and a carry: a suffix of that remainder which is a proper prefix of dsrRequest,
// held back because it might complete a query split across the next read. A remainder that
// is not a genuine prefix of dsrRequest is never held back, so bytes that could not
// possibly become a query are never stalled.
func splitDSRQueries(data []byte) (queries [][]byte, rest, carry []byte) {
	start := 0

	for {
		idx := bytes.Index(data[start:], dsrRequest)
		if idx < 0 {
			break
		}

		end := start + idx + len(dsrRequest)
		queries = append(queries, data[start:end])
		start = end
	}

	remainder := data[start:]
	carryLen := dsrPrefixLen(remainder)

	return queries, remainder[:len(remainder)-carryLen], remainder[len(remainder)-carryLen:]
}

// dsrPrefixLen returns the length of the longest suffix of data that is also a proper
// (non-full) prefix of dsrRequest, or 0 if data has no such suffix.
func dsrPrefixLen(data []byte) int {
	maxLen := min(len(data), len(dsrRequest)-1)

	for l := maxLen; l > 0; l-- {
		if bytes.Equal(data[len(data)-l:], dsrRequest[:l]) {
			return l
		}
	}

	return 0
}

// Screen returns the current rendered terminal screen as text.
func (s *Session) Screen() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.term.String()
}

// ScreenLines returns the current rendered terminal screen split into its individual
// rows, each padded with trailing spaces to the pty's column width. Use this over Screen
// when an assertion needs to reason about a single row (e.g. what shares a line with the
// prompt, or how far right some text is rendered).
func (s *Session) ScreenLines() []string {
	return strings.Split(s.Screen(), "\n")
}

// Raw returns the raw bytes captured from the pty so far, as a string.
func (s *Session) Raw() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.raw.String()
}

// WaitFor polls Screen every 50ms until it matches re, up to a 30s deadline. It returns
// the matching screen text, or fails the test with the full rendered screen and the tail
// of the raw output if the deadline elapses first.
func (s *Session) WaitFor(re *regexp.Regexp) string {
	s.t.Helper()

	deadline := time.Now().Add(waitForDeadline)

	for time.Now().Before(deadline) {
		screen := s.Screen()
		if re.MatchString(screen) {
			return screen
		}

		time.Sleep(waitForPollInterval)
	}

	s.t.Fatalf("timed out waiting for %q\n--- screen ---\n%s\n--- raw tail ---\n%s",
		re.String(), s.Screen(), tail(s.Raw(), 2000))

	return ""
}

// tail returns the last n bytes of s, or s unchanged if it is shorter than n.
func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}

	return s[len(s)-n:]
}

// WaitForPrompt waits for the "E2E:<code>>" prompt marker to appear on screen.
func (s *Session) WaitForPrompt() string {
	return s.WaitFor(promptRegexp)
}

// MarkerColor locates the first occurrence of marker on the rendered screen and returns
// the vt10x foreground/background color of its first cell (found via the terminal's
// per-cell Cell(x, y) accessor, x=column, y=row). found is false if marker is not
// present anywhere on screen. Callers needing a marker that hasn't rendered yet should
// WaitFor it first; byte and rune offsets coincide here because every marker and
// everything rendered ahead of it in these tests is ASCII.
func (s *Session) MarkerColor(marker string) (fg, bg vt10x.Color, found bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for y, line := range strings.Split(s.term.String(), "\n") {
		x := strings.Index(line, marker)
		if x < 0 {
			continue
		}

		glyph := s.term.Cell(x, y)
		return glyph.FG, glyph.BG, true
	}

	return 0, 0, false
}

// SendLine writes cmd followed by a carriage return to the pty, as if typed interactively.
func (s *Session) SendLine(cmd string) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	_, _ = s.pty.Write([]byte(cmd + "\r"))
}

// ExpectExit sends "exit\r" and waits up to 15s for the process to terminate. It fails the
// test (killing the process first) if the process does not exit in time. Call it at most
// once per session: it claims ownership of reaping the process via cmd.Wait(), so the
// Cleanup registered by Start knows not to wait a second time. On the timeout path, the
// Kill releases the pending Wait goroutine started below, which still reaps the process;
// Cleanup only kills and closes the pty at that point.
func (s *Session) ExpectExit() {
	s.t.Helper()

	s.SendLine("exit")

	done := make(chan error, 1)

	if s.takeWaitOwnership() {
		go func() { done <- s.cmd.Wait() }()
	}

	select {
	case <-done:
		return
	case <-time.After(exitDeadline):
		_ = s.cmd.Process.Kill()
		s.t.Fatalf("shell did not exit within %s\n--- screen ---\n%s", exitDeadline, s.Screen())
	}
}
