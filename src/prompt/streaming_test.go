package prompt

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStreamPrimary_NoSegments(t *testing.T) {
	env := new(mock.Environment)
	env.On("Pwd").Return("/test")
	env.On("Home").Return("/home")
	env.On("Shell").Return(shell.PWSH)
	env.On("Flags").Return(&runtime.Flags{Streaming: true})
	env.On("CursorPosition").Return(1, 1)
	env.On("StatusCodes").Return(0, "0")
	env.On("DirMatchesOneOf", testifymock.Anything, testifymock.Anything).Return(false)
	// Mock accent color retrieval for both Windows and macOS. The mock
	// forwards RunCommand as Called(command, args), so the expectation
	// takes two arguments: the command and the args slice.
	env.On("RunCommand", testifymock.Anything, testifymock.Anything).Return("4", nil)
	env.On("WindowsRegistryKeyValue", testifymock.Anything).Return(&runtime.WindowsRegistryValue{ValueType: runtime.DWORD, DWord: 0xFF0078D7}, nil)

	template.Cache = &cache.Template{
		Segments: maps.NewConcurrent[any](),
	}
	template.Init(env, nil, nil)
	terminal.Init(shell.PWSH)
	terminal.Colors = color.MakeColors(nil, false, "", env)

	engine := &Engine{
		Config: &config.Config{
			Blocks: []*config.Block{},
		},
		Env: env,
	}

	out := engine.StreamPrimary()
	prompts := collectChannelOutput(out, 100*time.Millisecond)

	// Initial prompt and the transient record with no pending segments
	assert.Len(t, prompts, 2)
}

func TestStreamPrimary_TransientRecord(t *testing.T) {
	env := setupStreamingTestEnv()

	engine := &Engine{
		Config: &config.Config{
			Blocks: []*config.Block{},
		},
		Env: env,
	}

	out := engine.StreamPrimary()
	prompts := collectChannelOutput(out, 100*time.Millisecond)

	// Initial prompt followed by a marker-prefixed transient record
	assert.Len(t, prompts, 2)
	assert.False(t, strings.HasPrefix(prompts[0], TransientMarker), "Initial prompt should not carry the transient marker")
	assert.True(t, strings.HasPrefix(prompts[1], TransientMarker), "Second record should carry the transient marker")
	assert.NotEmpty(t, strings.TrimPrefix(prompts[1], TransientMarker), "Transient record should contain the rendered prompt")
}

// mockSlowSession configures env's Session-segment dependencies (SSH_CONNECTION,
// SSH_CLIENT, Platform) so a config.SESSION segment's real Execute genuinely blocks for
// delay before completing - driving the real timeout/completion machinery instead of
// poking streamCycle internals directly.
func mockSlowSession(env *mock.Environment, delay time.Duration) {
	env.On("Getenv", "SSH_CONNECTION").Return("").After(delay)
	env.On("Getenv", "SSH_CLIENT").Return("")
	env.On("Platform").Return(runtime.WINDOWS)
}

func TestStreamPrimary_TransientRecord_RefreshedAfterCompletion(t *testing.T) {
	env := setupStreamingTestEnv()
	mockSlowSession(env, 50*time.Millisecond)

	slowSegment := &config.Segment{
		Type:    config.SESSION,
		Timeout: 5,
	}

	engine := &Engine{
		Config: &config.Config{
			Blocks: []*config.Block{
				{
					Type:      config.Prompt,
					Alignment: config.Left,
					Segments:  []*config.Segment{slowSegment},
				},
			},
		},
		Env: env,
	}

	out := engine.StreamPrimary()
	prompts := collectChannelOutput(out, 300*time.Millisecond)

	// initial prompt, initial transient, primary update, refreshed transient
	assert.Len(t, prompts, 4)
	assert.True(t, strings.HasPrefix(prompts[1], TransientMarker), "Second record should be the initial transient prompt")
	assert.False(t, strings.HasPrefix(prompts[2], TransientMarker), "Third record should be the primary prompt update")
	assert.True(t, strings.HasPrefix(prompts[len(prompts)-1], TransientMarker), "Last record should be the refreshed transient prompt")
}

func TestStreamPrimary_RecoversFromRenderPanic(t *testing.T) {
	// No Env: e.Primary() nil-dereferences, which must be recovered by the
	// producer goroutine - a panicking render costs one cycle, not the
	// process (the serve daemon relies on this).
	engine := &Engine{
		Config: &config.Config{
			Blocks: []*config.Block{},
		},
	}

	out := engine.StreamPrimary()
	prompts := collectChannelOutput(out, 100*time.Millisecond)

	// The panic aborts the cycle before any record is produced, and the
	// channel still closes cleanly.
	assert.Empty(t, prompts)

	// Abort must not hang after a panicked cycle.
	done := make(chan struct{})
	go func() {
		engine.Abort()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Abort should return after a panicked cycle")
	}
}

// Guards against the abort deadlock: when the record channel fills against a
// stalled consumer, the producer blocks in a send; Abort() must still unblock
// it (via the abort-aware send) instead of waiting forever for the goroutine
// to exit. Fifteen segments complete at staggered real times (an incrementing
// sleep per call to the mocked Getenv), so their completions cannot all
// coalesce into a single render; with nothing draining out (buffer 10), the
// producer is driven into a blocked send well before all of them land.
func TestStreamPrimary_AbortUnblocksSaturatedProducer(t *testing.T) {
	env := setupStreamingTestEnv()
	env.On("Getenv", "SSH_CLIENT").Return("")
	env.On("Platform").Return(runtime.WINDOWS)

	var calls atomic.Int64
	env.On("Getenv", "SSH_CONNECTION").Return("").Run(func(_ testifymock.Arguments) {
		n := calls.Add(1)
		time.Sleep(time.Duration(n) * 8 * time.Millisecond)
	})

	const segmentCount = 15

	segments := make([]*config.Segment, segmentCount)
	for i := range segments {
		segments[i] = &config.Segment{Type: config.SESSION, Timeout: 1}
	}

	engine := &Engine{
		Config: &config.Config{
			Blocks: []*config.Block{
				{Type: config.Prompt, Alignment: config.Left, Segments: segments},
			},
		},
		Env: env,
	}

	// Nothing reads from the output channel.
	_ = engine.StreamPrimary()

	// Give every staggered completion time to fire for real before checking Abort.
	time.Sleep(time.Duration(segmentCount+2) * 8 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		engine.Abort()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Abort must unblock a producer stuck in a record send")
	}
}

func setupStreamingTestEnv() *mock.Environment {
	env := new(mock.Environment)
	env.On("Pwd").Return("/test")
	env.On("Home").Return("/home")
	env.On("Shell").Return(shell.PWSH)
	env.On("Flags").Return(&runtime.Flags{Streaming: true})
	env.On("CursorPosition").Return(1, 1)
	env.On("StatusCodes").Return(0, "0")
	env.On("DirMatchesOneOf", testifymock.Anything, testifymock.Anything).Return(false)
	// Mock accent color retrieval for both Windows and macOS. The mock
	// forwards RunCommand as Called(command, args), so the expectation
	// takes two arguments: the command and the args slice.
	env.On("RunCommand", testifymock.Anything, testifymock.Anything).Return("4", nil)
	env.On("WindowsRegistryKeyValue", testifymock.Anything).Return(&runtime.WindowsRegistryValue{ValueType: runtime.DWORD, DWord: 0xFF0078D7}, nil)

	template.Cache = &cache.Template{
		Segments: maps.NewConcurrent[any](),
	}
	template.Init(env, nil, nil)
	terminal.Init(shell.PWSH)
	terminal.Colors = color.MakeColors(nil, false, "", env)

	return env
}

func collectChannelOutput(ch <-chan string, timeout time.Duration) []string {
	var results []string
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case result, ok := <-ch:
			if !ok {
				return results
			}
			results = append(results, result)
		case <-timer.C:
			return results
		}
	}
}

func TestStreamPrimary_FullFlow_WithRendering(t *testing.T) {
	env := setupStreamingTestEnv()
	mockSlowSession(env, 50*time.Millisecond)

	// Create segments with different speeds
	fastSegment := &config.Segment{
		Type:       "text",
		Template:   "FAST",
		Foreground: "#ffffff",
		Background: "#000000",
	}

	slowSegment := &config.Segment{
		Type:    config.SESSION,
		Timeout: 5,
	}

	engine := &Engine{
		Config: &config.Config{
			Blocks: []*config.Block{
				{
					Type:      config.Prompt,
					Alignment: config.Left,
					Segments:  []*config.Segment{fastSegment, slowSegment},
				},
			},
		},
		Env: env,
	}

	// Start streaming
	out := engine.StreamPrimary()

	// Collect all prompts
	prompts := collectChannelOutput(out, 300*time.Millisecond)

	// Should have at least 2 prompts: initial (with "...") and final (with the resolved segment)
	assert.GreaterOrEqual(t, len(prompts), 1, "Should have at least initial prompt")

	// First prompt should contain "..." for the pending segment
	if len(prompts) > 0 {
		assert.Contains(t, prompts[0], "...", "Initial prompt should show pending text")
	}

	// If we got multiple prompts, last one should not have "..."
	if len(prompts) > 1 {
		assert.NotContains(t, prompts[len(prompts)-1], "...", "Final prompt should not show pending text")
	}
}

func TestStreamPrimary_MultipleBlocks_MixedSpeed(t *testing.T) {
	env := setupStreamingTestEnv()
	mockSlowSession(env, 50*time.Millisecond)

	// Block 1: Fast segment
	fast1 := &config.Segment{
		Type:     "text",
		Template: "FAST1",
	}

	// Block 2: Slow segment
	slow1 := &config.Segment{
		Type:    config.SESSION,
		Timeout: 5,
	}

	// Block 3: Another fast segment
	fast2 := &config.Segment{
		Type:     "text",
		Template: "FAST2",
	}

	engine := &Engine{
		Config: &config.Config{
			Blocks: []*config.Block{
				{Type: config.Prompt, Alignment: config.Left, Segments: []*config.Segment{fast1}},
				{Type: config.Prompt, Alignment: config.Left, Segments: []*config.Segment{slow1}},
				{Type: config.Prompt, Alignment: config.Left, Segments: []*config.Segment{fast2}},
			},
		},
		Env: env,
	}

	// Start streaming
	out := engine.StreamPrimary()

	prompts := collectChannelOutput(out, 300*time.Millisecond)

	// Should receive prompts
	assert.NotEmpty(t, prompts, "Should receive streaming prompts")
}

func setupBasicStreamingTestEnv() *Engine {
	env := new(mock.Environment)
	env.On("Pwd").Return("/test")
	env.On("Home").Return("/home")
	env.On("Shell").Return(shell.PWSH)
	env.On("Flags").Return(&runtime.Flags{Streaming: true})
	env.On("CursorPosition").Return(1, 1)
	env.On("StatusCodes").Return(0, "0")
	env.On("DirMatchesOneOf", testifymock.Anything, testifymock.Anything).Return(false)
	// MakeColors resolves the accent color, which reads the registry on Windows
	// and shells out on macOS. Without these the helper panics on those hosts.
	env.On("RunCommand", testifymock.Anything, testifymock.Anything).Return("4", nil)
	env.On("WindowsRegistryKeyValue", testifymock.Anything).Return(&runtime.WindowsRegistryValue{ValueType: runtime.DWORD, DWord: 0xFF0078D7}, nil)

	template.Cache = &cache.Template{
		Segments: maps.NewConcurrent[any](),
	}
	template.Init(env, nil, nil)
	terminal.Init(shell.PWSH)
	terminal.Colors = color.MakeColors(nil, false, "", env)

	engine := &Engine{
		Config: &config.Config{
			Blocks: []*config.Block{},
		},
		Env: env,
	}

	return engine
}

func TestStreamPrimary_EarlyChannelClosure(t *testing.T) {
	engine := setupBasicStreamingTestEnv()

	// Start streaming with no pending segments
	// The goroutine should complete quickly and close channels properly
	out := engine.StreamPrimary()

	// Should be able to read from output channel without panic
	prompts := collectChannelOutput(out, 100*time.Millisecond)

	// Initial prompt and the transient record with no pending segments
	assert.Len(t, prompts, 2, "Should receive initial prompt and transient record")
}

// Validates that Abort() stops the producer from emitting further records and
// blocks until the producer goroutine has fully exited, even when a segment
// completes (and tries to notify) after the abort was issued - that late
// notification must not panic or deadlock.
func TestStreamPrimary_Abort_StopsRenderingAndDrains(t *testing.T) {
	env := setupStreamingTestEnv()
	mockSlowSession(env, 40*time.Millisecond)

	slowSegment := &config.Segment{
		Type:    config.SESSION,
		Timeout: 5,
	}

	engine := &Engine{
		Config: &config.Config{
			Blocks: []*config.Block{
				{
					Type:      config.Prompt,
					Alignment: config.Left,
					Segments:  []*config.Segment{slowSegment},
				},
			},
		},
		Env: env,
	}

	out := engine.StreamPrimary()

	// Consume the initial prompt + transient record before aborting, mirroring
	// how a server would start draining a cycle immediately. By this point the
	// segment's real 5ms timeout has already fired, so it is pending.
	<-out
	<-out

	// The segment's real completion (its mocked Getenv unblocks after 40ms) lands
	// AFTER Abort is issued below; the producer must consume that late event
	// without rendering it (and without the sender blocking).
	engine.Abort()

	// Abort must not return until the producer has exited, which implies out
	// and the stream's internal channels are both closed.
	select {
	case _, ok := <-out:
		assert.False(t, ok, "out channel should be closed after Abort returns")
	default:
		t.Error("out channel should be closed (readable as closed) after Abort returns")
	}

	// A second Abort() call must be a safe no-op.
	assert.NotPanics(t, func() { engine.Abort() })

	// Give the delayed segment-completion goroutine time to fire its late,
	// post-abort notification; it must not panic (send on closed channel)
	// or otherwise crash the test binary.
	time.Sleep(80 * time.Millisecond)
}

// Guards a class of pending-segment leak: a segment with no explicit Timeout, under a
// config with no global per-segment timeout either (Config.Streaming == 0), must never
// enter the pending path at all - it runs (and reports) through the plain executeSegment
// path - so the producer's pending set never grows and the cycle always closes on its own.
func TestStreamPrimary_NoStreamingTimeout_ChannelCloses(t *testing.T) {
	env := setupStreamingTestEnv()

	segment := &config.Segment{
		Type:       "text",
		Template:   "TEXT",
		Foreground: "#ffffff",
		Background: "#000000",
	}

	engine := &Engine{
		Config: &config.Config{
			// Streaming (the global per-segment timeout) deliberately left at 0
			Blocks: []*config.Block{
				{
					Type:      config.Prompt,
					Alignment: config.Left,
					Segments:  []*config.Segment{segment},
				},
			},
		},
		Env: env,
	}

	out := engine.StreamPrimary()

	// collectChannelOutput cannot distinguish a closed channel from a timeout,
	// so assert closure explicitly - that is the whole point of this test.
	var prompts []string
	deadline := time.After(2 * time.Second)

	for {
		select {
		case record, ok := <-out:
			if !ok {
				// Initial prompt + transient record, then a clean close.
				assert.Len(t, prompts, 2)
				return
			}

			prompts = append(prompts, record)
		case <-deadline:
			t.Fatal("output channel never closed: a segment without a timeout must never become pending")
		}
	}
}

// Validates that Abort can be called (and is a safe no-op) when no cycle has
// ever been started.
func TestStreamPrimary_Abort_NoCycleStarted(t *testing.T) {
	engine := &Engine{}
	assert.NotPanics(t, func() { engine.Abort() })
}

// Validates that a fresh StreamPrimary cycle works correctly after a previous
// cycle was aborted - this is the serialization guarantee the serve command
// depends on.
func TestStreamPrimary_Abort_ThenNewCycleWorks(t *testing.T) {
	engine := setupBasicStreamingTestEnv()

	firstOut := engine.StreamPrimary()
	engine.Abort()
	// Drain any buffered records from the aborted cycle.
	for range firstOut {
	}

	secondOut := engine.StreamPrimary()
	prompts := collectChannelOutput(secondOut, 100*time.Millisecond)

	assert.Len(t, prompts, 2, "A new cycle after abort should render normally")
}

// collectRecords reads every record from ch until it closes. bound is a hang guard only: reaching
// it before the channel closes is always a test failure, never a signal that the cycle is done -
// a producer that is merely slow must still be given the chance to finish and close the channel.
func collectRecords(t *testing.T, ch <-chan string, bound time.Duration) []string {
	t.Helper()

	var records []string
	timer := time.NewTimer(bound)
	defer timer.Stop()

	for {
		select {
		case record, ok := <-ch:
			if !ok {
				return records
			}

			records = append(records, record)
		case <-timer.C:
			require.Fail(t, "streaming channel did not close within the hang bound",
				"bound: %s, records collected so far: %d", bound, len(records))
			return records
		}
	}
}

// splitRecords separates primary prompt records from transient records, which are identified by
// the TransientMarker prefix.
func splitRecords(records []string) (primaries, transients []string) {
	for _, record := range records {
		if strings.HasPrefix(record, TransientMarker) {
			transients = append(transients, record)
			continue
		}

		primaries = append(primaries, record)
	}

	return primaries, transients
}

// streamTestCase describes one segment inside a TestStreamPrimary_Lifecycle case: how its fake
// writer behaves, and whether the test needs to release it explicitly after the first record.
type streamTestCase struct {
	ID                      string
	Delay                   time.Duration
	ReleaseAfterFirstRecord bool
	Panics                  bool
	NeverCompletes          bool
}

// lifecycleCase is one row of the TestStreamPrimary_Lifecycle table.
type lifecycleCase struct {
	ExpectCountInLast  map[string]int
	Case               string
	Segments           []streamTestCase
	ExpectInLast       []string
	ExpectNotInLast    []string
	PendingLimit       time.Duration
	Streaming          int
	ExpectTransients   int
	ExpectPrimariesMin int
	ExpectPrimariesMax int
	ReleaseTogether    bool
}

// TestStreamPrimary_Lifecycle drives the streaming producer through the timedOut/completed/
// abandoned event machinery via one or more fake segments, and checks the record sequence it
// produces. The core invariant every case shares: once any primary record still shows a
// PENDING- placeholder, the very last primary record produced for that cycle must show none.
func TestStreamPrimary_Lifecycle(t *testing.T) {
	cases := []lifecycleCase{
		{
			Case: "resolves before timeout",
			Segments: []streamTestCase{
				{ID: "resolves-before-timeout"},
			},
			// A larger window than the spec's 5ms: nothing here is meant to race the timeout at
			// all, and 5ms is tight enough to flake under a loaded test binary (see the Streaming
			// field doc comment).
			Streaming:          250,
			ExpectTransients:   1,
			ExpectPrimariesMin: 1,
			ExpectPrimariesMax: 1,
			ExpectInLast:       []string{"resolved/42"},
			ExpectNotInLast:    []string{"PENDING"},
		},
		{
			Case: "resolves after timeout",
			Segments: []streamTestCase{
				{ID: "resolves-after-timeout", ReleaseAfterFirstRecord: true},
			},
			ExpectTransients:   2,
			ExpectPrimariesMin: 2,
			ExpectPrimariesMax: 2,
			ExpectInLast:       []string{"resolved/42"},
		},
		{
			Case: "two segments released one after another",
			Segments: []streamTestCase{
				{ID: "one-after-another-a", ReleaseAfterFirstRecord: true},
				{ID: "one-after-another-b", ReleaseAfterFirstRecord: true},
			},
			ExpectTransients:   2,
			ExpectPrimariesMin: 2,
			ExpectPrimariesMax: 3,
			ExpectCountInLast:  map[string]int{"resolved/42": 2},
			ExpectNotInLast:    []string{"PENDING"},
		},
		{
			Case: "two segments released together",
			Segments: []streamTestCase{
				{ID: "together-a", ReleaseAfterFirstRecord: true},
				{ID: "together-b", ReleaseAfterFirstRecord: true},
			},
			ReleaseTogether:    true,
			ExpectTransients:   2,
			ExpectPrimariesMin: 2,
			ExpectPrimariesMax: 3,
			ExpectCountInLast:  map[string]int{"resolved/42": 2},
			ExpectNotInLast:    []string{"PENDING"},
		},
		{
			Case: "mixed: one fast one slow",
			Segments: []streamTestCase{
				{ID: "mixed-fast"},
				{ID: "mixed-slow", ReleaseAfterFirstRecord: true},
			},
			// The fast segment must not itself race the timeout; see the Streaming field doc
			// comment. The slow segment is gated on its release channel regardless of this value.
			Streaming:          250,
			ExpectTransients:   2,
			ExpectPrimariesMin: 2,
			ExpectPrimariesMax: 2,
			ExpectCountInLast:  map[string]int{"resolved/42": 2},
			ExpectNotInLast:    []string{"PENDING"},
		},
		{
			Case: "panic before timeout",
			Segments: []streamTestCase{
				{ID: "panic-before-timeout", Panics: true},
			},
			// See the Streaming field doc comment: this panic must not itself race the timeout.
			Streaming:          250,
			ExpectTransients:   1,
			ExpectPrimariesMin: 1,
			ExpectPrimariesMax: 1,
			ExpectNotInLast:    []string{"PENDING", "resolved"},
		},
		{
			Case: "panic after timeout",
			Segments: []streamTestCase{
				{ID: "panic-after-timeout", ReleaseAfterFirstRecord: true, Panics: true},
			},
			ExpectTransients:   2,
			ExpectPrimariesMin: 2,
			ExpectPrimariesMax: 2,
			ExpectNotInLast:    []string{"PENDING", "resolved"},
		},
		{
			Case: "never completes is abandoned",
			Segments: []streamTestCase{
				{ID: "never-completes", NeverCompletes: true},
			},
			PendingLimit:       50 * time.Millisecond,
			ExpectTransients:   2,
			ExpectPrimariesMin: 2,
			ExpectPrimariesMax: 2,
			ExpectNotInLast:    []string{"PENDING", "resolved"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			runLifecycleCase(t, &tc)
		})
	}
}

// runLifecycleCase installs one fake writer control per segment, starts the cycle, drives any
// release channels the case calls for, then collects and asserts on the resulting record stream.
func runLifecycleCase(t *testing.T, tc *lifecycleCase) {
	t.Helper()

	if tc.PendingLimit != 0 {
		original := pendingSegmentLimit
		pendingSegmentLimit = tc.PendingLimit
		t.Cleanup(func() { pendingSegmentLimit = original })
	}

	env := setupStreamingTestEnv()

	streaming := tc.Streaming
	if streaming == 0 {
		streaming = 5
	}

	segs := make([]*config.Segment, len(tc.Segments))
	var releaseControls []*streamTestControl

	for i, sc := range tc.Segments {
		control := &streamTestControl{delay: sc.Delay, panics: sc.Panics}
		if sc.ReleaseAfterFirstRecord || sc.NeverCompletes {
			control.release = make(chan struct{})
		}

		streamTestControls.Store(sc.ID, control)
		t.Cleanup(func() { streamTestControls.Delete(sc.ID) })

		segs[i] = streamTestSegment(sc.ID)

		if sc.ReleaseAfterFirstRecord {
			releaseControls = append(releaseControls, control)
		}
	}

	engine := &Engine{
		Config: &config.Config{
			Streaming: streaming,
			Blocks: []*config.Block{
				{Type: config.Prompt, Alignment: config.Left, Segments: segs},
			},
		},
		Env: env,
	}

	out := engine.StreamPrimary()

	var records []string

	if len(releaseControls) > 0 {
		first, ok := <-out
		require.True(t, ok, "case %s: streaming channel closed before the first record", tc.Case)
		records = append(records, first)

		for i, control := range releaseControls {
			close(control.release)

			last := i == len(releaseControls)-1
			if tc.ReleaseTogether || last {
				continue
			}

			// Wait for the record reflecting this segment's own completion before releasing the
			// next one, so the two completions land as separate, serialized events rather than
			// racing to be absorbed together.
			record, ok := <-out
			require.True(t, ok, "case %s: streaming channel closed while awaiting an intermediate record", tc.Case)
			records = append(records, record)
		}
	}

	records = append(records, collectRecords(t, out, 5*time.Second)...)

	primaries, transients := splitRecords(records)

	assert.Len(t, transients, tc.ExpectTransients, "case %s: transient record count", tc.Case)
	assert.GreaterOrEqual(t, len(primaries), tc.ExpectPrimariesMin, "case %s: primary record count (min)", tc.Case)
	assert.LessOrEqual(t, len(primaries), tc.ExpectPrimariesMax, "case %s: primary record count (max)", tc.Case)

	require.NotEmpty(t, primaries, "case %s: expected at least one primary record", tc.Case)
	last := primaries[len(primaries)-1]

	for _, substr := range tc.ExpectInLast {
		assert.Contains(t, last, substr, "case %s: last primary missing %q", tc.Case, substr)
	}

	for _, substr := range tc.ExpectNotInLast {
		assert.NotContains(t, last, substr, "case %s: last primary should not contain %q", tc.Case, substr)
	}

	for substr, count := range tc.ExpectCountInLast {
		assert.Equal(t, count, strings.Count(last, substr), "case %s: last primary %q occurrence count", tc.Case, substr)
	}

	hadPending := false
	for _, primary := range primaries {
		if strings.Contains(primary, "PENDING-") {
			hadPending = true
			break
		}
	}

	if !hadPending {
		return
	}

	assert.NotContains(t, last, "PENDING-", "case %s: the last primary must never still show a placeholder", tc.Case)
}

// TestStreamPrimary_StraddlingTimeoutNeverSticks is a regression fuzz test for the bug this
// rewrite fixes: a segment resolving within a few milliseconds of the streaming timeout used to
// race the render against the writer, and could leave a placeholder in the last primary record.
// It runs many iterations with a randomized delay straddling the 5ms timeout so the race window
// gets exercised at many different offsets instead of just one fixed one.
func TestStreamPrimary_StraddlingTimeoutNeverSticks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping the straddling-timeout fuzz loop in short mode")
	}

	const iterations = 200

	for i := range iterations {
		delay := time.Duration(2+rand.IntN(7)) * time.Millisecond
		id := fmt.Sprintf("straddle-%d", i)

		streamTestControls.Store(id, &streamTestControl{delay: delay})

		runStraddlingIteration(t, i, id)

		streamTestControls.Delete(id)
	}
}

// runStraddlingIteration runs one iteration of the straddling-timeout fuzz loop against the
// segment already installed under id.
func runStraddlingIteration(t *testing.T, iteration int, id string) {
	t.Helper()

	env := setupStreamingTestEnv()

	engine := &Engine{
		Config: &config.Config{
			Streaming: 5,
			Blocks: []*config.Block{
				{Type: config.Prompt, Alignment: config.Left, Segments: []*config.Segment{streamTestSegment(id)}},
			},
		},
		Env: env,
	}

	out := engine.StreamPrimary()
	records := collectRecords(t, out, 5*time.Second)
	primaries, transients := splitRecords(records)

	require.NotEmpty(t, primaries, "iteration %d: no primary records produced; records: %v", iteration, records)

	last := primaries[len(primaries)-1]
	assert.Contains(t, last, "resolved/42", "iteration %d: last primary missing resolved/42; records: %v", iteration, records)
	assert.NotContains(t, last, "PENDING-", "iteration %d: last primary still shows a placeholder; records: %v", iteration, records)

	transientCount := len(transients)
	assert.True(t, transientCount == 1 || transientCount == 2,
		"iteration %d: unexpected transient count %d; records: %v", iteration, transientCount, records)

	hadPending := false
	for _, primary := range primaries {
		if strings.Contains(primary, "PENDING-") {
			hadPending = true
			break
		}
	}

	if !hadPending {
		return
	}

	assert.Equal(t, 2, transientCount,
		"iteration %d: a primary that was ever pending must produce a refreshed transient; records: %v", iteration, records)
}

// TestStreamPrimary_PlaceholderUsesConfiguredColors checks that a pending segment's placeholder
// renders with the segment's raw configured colors, never with a color that would require reading
// the writer - which is off limits while the segment is still pending (see RenderPlaceholder).
func TestStreamPrimary_PlaceholderUsesConfiguredColors(t *testing.T) {
	env := setupStreamingTestEnv()

	id := "placeholder-colors"
	control := &streamTestControl{release: make(chan struct{})}
	streamTestControls.Store(id, control)
	t.Cleanup(func() { streamTestControls.Delete(id) })

	engine := &Engine{
		Config: &config.Config{
			Streaming: 5,
			Blocks: []*config.Block{
				{Type: config.Prompt, Alignment: config.Left, Segments: []*config.Segment{streamTestSegment(id)}},
			},
		},
		Env: env,
	}

	out := engine.StreamPrimary()

	first, ok := <-out
	require.True(t, ok, "streaming channel closed before the first record")

	close(control.release)

	rest := collectRecords(t, out, 2*time.Second)
	primaries, _ := splitRecords(append([]string{first}, rest...))

	require.Len(t, primaries, 2, "expected an initial pending primary and a resolved primary")

	defaults := &color.Defaults{}
	whiteForeground := "\x1b[" + defaults.ToAnsi(color.Ansi("#ffffff"), false).String() + "m"
	greenForeground := "\x1b[" + defaults.ToAnsi(color.Ansi("#00ff00"), false).String() + "m"

	assert.Contains(t, primaries[0], whiteForeground, "the placeholder should use the segment's raw configured foreground")
	assert.NotContains(t, primaries[0], greenForeground, "the placeholder must not resolve a foreground template against the writer")

	assert.Contains(t, primaries[1], greenForeground, "the resolved segment should use its matched foreground template")
}
