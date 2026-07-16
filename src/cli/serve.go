package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/prompt"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/spf13/cobra"
)

// requestPipe is the path to a named pipe (fifo) to read requests from
// instead of stdin. Unix only - used by shells that cannot hold a child's
// stdin open across prompts (fish).
var requestPipe string

// serveCmd represents the serve command
var serveCmd = createServeCmd()

func init() {
	RootCmd.AddCommand(serveCmd)
}

// serveRequest mirrors the request protocol documented in the implementation
// plan: one JSON object per line on stdin. Unknown fields are ignored by
// encoding/json by default, which gives us forward compatibility for free.
type serveRequest struct {
	Env           map[string]string `json:"env"`
	Command       string            `json:"command"`
	Shell         string            `json:"shell"`
	ShellVersion  string            `json:"shell-version"`
	PWD           string            `json:"pwd"`
	PSWD          string            `json:"pswd"`
	PipeStatus    string            `json:"pipestatus"`
	ID            int64             `json:"id"`
	Status        int               `json:"status"`
	StackCount    int               `json:"stack-count"`
	TerminalWidth int               `json:"terminal-width"`
	JobCount      int               `json:"job-count"`
	ExecutionTime float64           `json:"execution-time"`
	NoStatus      bool              `json:"no-status"`
	Cleared       bool              `json:"cleared"`
	// Wait makes the render synchronous: every segment resolves (bounded by
	// the regular per-segment timeouts, like print primary) and exactly two
	// records are emitted - the final primary and the transient. For shells
	// without an async record consumer (bash): incremental updates would pile
	// up unread in the pipe buffer, and a full pipe blocks the record copier,
	// which stopActiveCycle waits on.
	Wait bool `json:"wait"`
}

const (
	// serveCommandRender asks the daemon to render a new primary prompt cycle.
	serveCommandRender = "render"
	// serveCommandAbort asks the daemon to stop the active render cycle, if any.
	serveCommandAbort = "abort"
	// serveCommandQuit asks the daemon to flush caches and exit cleanly.
	serveCommandQuit = "quit"
)

// serveIDMarker separates the cycle id prefix from the record payload on
// stdout: "<id>\x1f<payload>\x00". \x1f is the ASCII unit separator.
const serveIDMarker = "\x1f"

func createServeCmd() *cobra.Command {
	serveCmd := &cobra.Command{
		Use:    "serve",
		Short:  "Start a persistent prompt server that streams prompt updates over stdio",
		Hidden: true,
		Args:   cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			if shellName == "" {
				shellName = shell.GENERIC
			}

			in, err := openServeInput(requestPipe)
			if err != nil {
				os.Exit(1)
			}

			options := []cache.Option{cache.Persist}

			cache.Init(shellName, options...)

			defer cache.Close()

			// template.SaveCache() requires template.Init() to have run at
			// least once (it reads package-level state set there); if the
			// daemon quits/hits EOF before ever handling a render request,
			// skip it instead of panicking on that unset state.
			if renderedAtLeastOnce := runServeLoop(in, os.Stdout); renderedAtLeastOnce {
				template.SaveCache()
			}
		},
	}

	serveCmd.Flags().StringVar(&shellName, "shell", "", "the shell to serve for")
	serveCmd.Flags().StringVar(&requestPipe, "request-pipe", "", "named pipe (fifo) to read requests from instead of stdin")

	// Hide flags that are for internal use only.
	_ = serveCmd.Flags().MarkHidden("request-pipe")

	return serveCmd
}

// openServeInput returns the request source: stdin by default, or the given
// named pipe (fifo) opened read-write. O_RDWR is the load-bearing detail: the
// daemon itself keeps a writer on the fifo, so a client doing
// open-write-close per request (the only write primitive fish has) never
// EOFs the read side. Unix only - the shell owns the fifo's lifecycle.
//
// Clients must write each request in a single write(2) call; requests from a
// single sequential writer (one shell session) never interleave regardless
// of size.
func openServeInput(pipePath string) (*os.File, error) {
	if pipePath == "" {
		return os.Stdin, nil
	}

	return os.OpenFile(pipePath, os.O_RDWR, 0)
}

// serveActiveCycle tracks the currently rendering cycle so a new render
// request (or an abort) can interrupt it before starting the next one. All
// engine rendering must stay serialized: the engine's prompt builder and the
// terminal package globals are not thread-safe, so the previous cycle's
// producer goroutine must be fully stopped (Abort blocks until it is) before
// a new *prompt.Engine starts rendering.
//
// The copier goroutine started in startRenderCycle is the sole reader of the
// engine's streamed-record channel; copierDone is closed once that goroutine
// has drained it to completion (channel closed), which only happens after
// engine.Abort() has returned. Nothing else may read from that channel.
type serveActiveCycle struct {
	engine     *prompt.Engine
	copierDone chan struct{}
}

// runServeLoop reads newline-delimited JSON requests from in and writes
// NUL-delimited, cycle-id-prefixed prompt records to out. It returns when it
// reads a quit command or hits EOF on stdin. The returned bool reports
// whether at least one render request was handled, so the caller knows
// whether template.SaveCache() (which requires template.Init() to have run)
// is safe to call.
//
// stdout carries ONLY protocol records - never log output. Panics during
// cycle setup and rendering are recovered (see startRenderCycle and
// StreamPrimary) so a broken render costs one prompt, not the daemon; the
// shell additionally redirects this process's stderr so anything unrecovered
// can never reach the user's terminal.
func runServeLoop(in, out *os.File) bool {
	scanner := bufio.NewScanner(in)
	// Env payloads (a POSH_* overlay plus PATH) can exceed the default 64 KB
	// scanner buffer, so grow it up front.
	scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)

	var active *serveActiveCycle
	renderedAtLeastOnce := false

	// envKeys tracks which variables the previous request's overlay set, so
	// a variable that disappears from a later request (e.g. VIRTUAL_ENV after
	// `deactivate`) gets unset instead of pinning its stale value for the rest
	// of the daemon's life. Scoped to the loop so repeated invocations in the
	// same process (tests) never inherit a previous loop's keys. The serve
	// loop is single-threaded, so no locking.
	envKeys := map[string]struct{}{}

	stopActiveCycle := func() {
		if active == nil {
			return
		}

		// Abort blocks until the previous cycle's producer goroutine has
		// fully exited, guaranteeing no two cycles ever render concurrently.
		// For a Wait cycle (renderComplete) Abort is a no-op - there the
		// wait on copierDone below provides the same guarantee, since the
		// record channel only closes when the render goroutine is done.
		active.engine.Abort()

		// The copier goroutine is the sole reader of the cycle's record
		// channel; wait for it to observe the channel close so we never have
		// two goroutines reading it and never race the next cycle's stdout
		// writes against this one's.
		<-active.copierDone

		active = nil
	}

	for scanner.Scan() {
		line := scanner.Bytes()

		// Strip a UTF-8 BOM: .NET's default UTF8 encoding writes one on the
		// StreamWriter's first write, which would otherwise make the first
		// request line of a session unparseable.
		line = bytes.TrimPrefix(line, []byte{0xEF, 0xBB, 0xBF})

		if len(line) == 0 {
			continue
		}

		var req serveRequest
		if err := json.Unmarshal(line, &req); err != nil {
			// Malformed line: ignore for forward/backward compatibility.
			continue
		}

		switch req.Command {
		case serveCommandRender:
			// A new render request implicitly aborts whatever is running.
			stopActiveCycle()
			// A nil cycle means setup panicked before prompt.New completed -
			// template.Init may never have run, in which case the shutdown
			// path must not call template.SaveCache (it dereferences state
			// only Init sets). A started cycle implies Init completed.
			if active = startRenderCycle(&req, out, envKeys); active != nil {
				renderedAtLeastOnce = true
			}
		case serveCommandAbort:
			stopActiveCycle()
		case serveCommandQuit:
			stopActiveCycle()
			return renderedAtLeastOnce
		default:
			// Unknown command: ignore for forward compatibility.
		}
	}

	// EOF (or a scanner error) on stdin: behave like an explicit quit so
	// caches are still flushed by the caller's deferred cleanup.
	stopActiveCycle()

	return renderedAtLeastOnce
}

func applyEnvOverlay(env map[string]string, keys map[string]struct{}) {
	for key := range keys {
		if _, ok := env[key]; ok {
			continue
		}

		_ = os.Unsetenv(key)
		delete(keys, key)
	}

	for key, value := range env {
		keys[key] = struct{}{}
		_ = os.Setenv(key, value)
	}
}

// startRenderCycle builds a fresh engine for the request (mirroring
// stream.go/print.go) and starts copying its streamed records to stdout,
// prefixed with the cycle id, in a background goroutine. It does not wait
// for the cycle to complete.
//
// A panic while setting up the cycle (e.g. in prompt.New) is recovered and
// reported as "no cycle": the daemon stays alive, the shell's waiter times
// out and falls back to the legacy path for that prompt.
func startRenderCycle(req *serveRequest, out *os.File, envKeys map[string]struct{}) (cycle *serveActiveCycle) {
	defer func() {
		if r := recover(); r != nil {
			cycle = nil
		}
	}()

	// Apply the env overlay BEFORE constructing the engine so segment
	// execution and config templates observe the calling shell's
	// environment. v1 accepts the theoretical race with a still-running
	// background segment from a previous (aborted) cycle reading the
	// overlay applied for this one - see the Engine-per-cycle discussion in
	// the implementation plan.
	applyEnvOverlay(req.Env, envKeys)

	if req.PWD != "" {
		if info, err := os.Stat(req.PWD); err == nil && info.IsDir() {
			_ = os.Chdir(req.PWD)
		}
	}

	// The template cache is per-prompt context (PWD, Folder, Code, Jobs, ...)
	// built once per PROCESS by template.Init - in a daemon that would pin
	// every render to the first request's context (e.g. cd never updating the
	// path segment). Drop it so prompt.New rebuilds it for this request.
	template.ResetCache()

	// prompt.New decodes a FRESH config graph from the session cache on every
	// cycle. Do not memoize or share it across cycles: segment structs carry
	// runtime state, and goroutines abandoned by an aborted cycle still hold
	// pointers into their own cycle's graph - a shared graph would let them
	// race the active render.
	shellName := req.Shell
	if shellName == "" {
		shellName = shell.GENERIC
	}

	flags := &runtime.Flags{
		ConfigPath:    configFlag,
		PWD:           req.PWD,
		PSWD:          req.PSWD,
		PipeStatus:    req.PipeStatus,
		ErrorCode:     req.Status,
		ExecutionTime: req.ExecutionTime,
		StackCount:    req.StackCount,
		TerminalWidth: req.TerminalWidth,
		Shell:         shellName,
		ShellVersion:  req.ShellVersion,
		Plain:         plain,
		Type:          prompt.PRIMARY,
		Cleared:       req.Cleared,
		NoExitCode:    req.NoStatus,
		JobCount:      req.JobCount,
		IsPrimary:     true,
		Escape:        true,
		Streaming:     !req.Wait,
	}

	eng := prompt.New(flags)

	var records <-chan string
	if req.Wait {
		records = renderComplete(eng)
	} else {
		records = eng.StreamPrimary()
	}

	return &serveActiveCycle{
		engine:     eng,
		copierDone: copyRecords(req.ID, records, out),
	}
}

// renderComplete produces the two records of a Wait render: the fully
// resolved primary prompt (Streaming is off, so segments block until done,
// bounded by their regular timeouts - print primary semantics) and the
// transient prompt. Rendered in a single goroutine for the same
// thread-safety reasons as StreamPrimary.
//
// Exactly two records are emitted even when the render panics: wait-mode
// clients (Clink) block-read both records with no timeout mechanism, so a
// short reply would leave them hung on a silent daemon. An empty primary
// tells the client to fall back to its one-shot path for this prompt.
func renderComplete(eng *prompt.Engine) <-chan string {
	records := make(chan string, 2)

	go func() {
		defer close(records)

		sent := 0
		defer func() {
			if recover() == nil {
				return
			}

			if sent == 0 {
				records <- ""
			}

			if sent <= 1 {
				records <- prompt.TransientMarker
			}
		}()

		records <- eng.Primary()
		sent++
		records <- prompt.TransientMarker + eng.ExtraPrompt(prompt.Transient)
		sent++
	}()

	return records
}

// copyRecords copies prompt records to out prefixed with the cycle id and
// closes the returned channel once the source channel is exhausted.
func copyRecords(id int64, records <-chan string, out *os.File) chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)

		for record := range records {
			// Single Fprintf call so each record is written to stdout
			// atomically with respect to other os.Stdout writers in this
			// process (there are none today, but keep the invariant).
			// The write error is deliberately unchecked: on Unix, a broken
			// stdout pipe raises SIGPIPE (default disposition on fd 1) and
			// terminates the daemon - the desired lifecycle when the shell
			// disappears without sending quit, and the only exit signal in
			// the request-pipe transport where stdin EOF never arrives.
			fmt.Fprintf(out, "%d%s%s\x00", id, serveIDMarker, record)
		}

		// Deliberately no cache persistence here: unlike stream/print
		// (--save-cache on every prompt), serve keeps the segment and
		// template caches in memory for the daemon's lifetime - that's the
		// whole point of a long-lived process. Caches are only flushed to
		// disk once, on clean shutdown (quit/EOF), via the cache.Close()/
		// template.SaveCache() defer in createServeCmd.
	}()

	return done
}
