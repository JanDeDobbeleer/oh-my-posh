package prompt

import (
	"fmt"
	"strings"
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

func TestStreamPrimary_TransientRecord_RefreshedAfterCompletion(t *testing.T) {
	env := setupStreamingTestEnv()

	slowSegment := &config.Segment{
		Type:       "text",
		Template:   "SLOW",
		Pending:    true,
		Foreground: "#ffffff",
		Background: "#000000",
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
		Env:              env,
		streamingResults: make(chan *config.Segment, 10),
	}

	err := slowSegment.MapSegmentWithWriter(env)
	require.NoError(t, err)

	engine.pendingSegments.Store(slowSegment.Name(), true)

	out := engine.StreamPrimary()

	go func() {
		time.Sleep(50 * time.Millisecond)
		slowSegment.Pending = false
		engine.notifySegmentCompletion(slowSegment)
	}()

	prompts := collectChannelOutput(out, 200*time.Millisecond)

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

// TestStreamPrimary_AbortUnblocksSaturatedProducer guards against the abort
// deadlock: when the record channel fills against a stalled consumer, the
// producer blocks in a send; Abort() must still unblock it (via the
// abort-aware send) instead of waiting forever for the goroutine to exit.
func TestStreamPrimary_AbortUnblocksSaturatedProducer(t *testing.T) {
	env := setupStreamingTestEnv()

	engine := &Engine{
		Config: &config.Config{
			Blocks: []*config.Block{},
		},
		Env:              env,
		streamingResults: make(chan *config.Segment, 100),
	}

	// Pre-register pending segments that never get removed, so the producer
	// keeps looping over completions instead of finishing the cycle.
	segments := make([]*config.Segment, 20)
	for i := range segments {
		segments[i] = &config.Segment{Type: config.SegmentType(fmt.Sprintf("test-%d", i)), Pending: true}
		engine.pendingSegments.Store(segments[i].Name(), true)
	}

	// Nothing reads from the channel: the producer will saturate the record
	// buffer and block in a send while processing the completions. Feed them
	// directly (bypassing notifySegmentCompletion, which would drain
	// pendingSegments and end the cycle early).
	_ = engine.StreamPrimary()

	for _, segment := range segments {
		segment.Pending = false
		engine.streamingResults <- segment
	}

	// Give the producer time to fill the record buffer and block.
	time.Sleep(100 * time.Millisecond)

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

func TestStreamPrimary_WithPendingSegments(t *testing.T) {
	engine := &Engine{
		streamingResults: make(chan *config.Segment, 10),
	}

	segment := &config.Segment{
		Type:    "text",
		Pending: true,
	}

	// Track as pending
	engine.pendingSegments.Store(segment.Name(), true)

	// Simulate segment completion in background
	go func() {
		time.Sleep(50 * time.Millisecond)
		segment.Pending = false
		engine.notifySegmentCompletion(segment)
	}()

	// Verify notification is received
	select {
	case completed := <-engine.streamingResults:
		assert.Equal(t, segment, completed)
		assert.False(t, completed.Pending)
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected segment completion notification")
	}
}

func TestCountPendingSegments(t *testing.T) {
	cases := []struct {
		Case     string
		Segments []string
		Count    int
	}{
		{Case: "No pending segments", Count: 0, Segments: []string{}},
		{Case: "One pending segment", Count: 1, Segments: []string{"segment1"}},
		{Case: "Multiple pending segments", Count: 3, Segments: []string{"segment1", "segment2", "segment3"}},
	}

	for _, tc := range cases {
		engine := &Engine{}

		for _, name := range tc.Segments {
			engine.pendingSegments.Store(name, true)
		}

		count := engine.countPendingSegments()
		assert.Equal(t, tc.Count, count, tc.Case)
	}
}

func TestNotifySegmentCompletion(t *testing.T) {
	cases := []struct {
		Case           string
		StreamingSetup bool
		SegmentPending bool
		ExpectNotify   bool
	}{
		{Case: "No streaming channel", StreamingSetup: false, SegmentPending: true, ExpectNotify: false},
		{Case: "Segment not pending", StreamingSetup: true, SegmentPending: false, ExpectNotify: false},
		{Case: "Valid notification", StreamingSetup: true, SegmentPending: true, ExpectNotify: true},
	}

	for _, tc := range cases {
		engine := &Engine{}
		segment := &config.Segment{Type: "test"}

		if tc.StreamingSetup {
			engine.streamingResults = make(chan *config.Segment, 10)
		}

		if tc.SegmentPending {
			engine.pendingSegments.Store(segment.Name(), true)
		}

		engine.notifySegmentCompletion(segment)

		if tc.ExpectNotify {
			select {
			case received := <-engine.streamingResults:
				assert.Equal(t, segment, received, tc.Case)
			case <-time.After(100 * time.Millisecond):
				t.Errorf("%s: Expected notification but got timeout", tc.Case)
			}
		} else if tc.StreamingSetup {
			select {
			case <-engine.streamingResults:
				t.Errorf("%s: Unexpected notification received", tc.Case)
			case <-time.After(50 * time.Millisecond):
				// Expected - no notification
			}
		}
	}
}

func TestTrackPendingSegment(t *testing.T) {
	engine := &Engine{
		streamingResults: make(chan *config.Segment, 10),
	}

	segment := &config.Segment{
		Type:    "test",
		Pending: true,
	}

	done := make(chan bool)

	// Pre-register segment as pending (this happens in writeSegmentsConcurrently in real code)
	engine.pendingSegments.Store(segment.Name(), true)

	// Start tracking
	engine.trackPendingSegment(segment, done)

	// Verify segment is tracked
	_, ok := engine.pendingSegments.Load(segment.Name())
	assert.True(t, ok, "Segment should be tracked")

	// Simulate completion
	close(done)

	// Wait for goroutine to process
	select {
	case completed := <-engine.streamingResults:
		assert.Equal(t, segment, completed)
		assert.False(t, segment.Pending, "Segment should no longer be pending")
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected segment completion notification")
	}

	// Verify segment is no longer tracked
	_, ok = engine.pendingSegments.Load(segment.Name())
	assert.False(t, ok, "Segment should no longer be tracked")
}

func TestRenderFromBlocks(_ *testing.T) {
	env := new(mock.Environment)
	env.On("Shell").Return(shell.PWSH)
	env.On("Flags").Return(&runtime.Flags{})

	// This test validates that renderFromBlocks properly delegates to primaryInternal
	engine := &Engine{
		Config: &config.Config{
			Blocks: []*config.Block{},
		},
		Env:       env,
		allBlocks: []*config.Block{},
	}

	// Just verify it doesn't panic - full integration tested elsewhere
	_ = engine.renderFromBlocks()
}

func TestPrimaryInternal_FromCache(_ *testing.T) {
	env := new(mock.Environment)
	env.On("Shell").Return(shell.PWSH)
	env.On("Flags").Return(&runtime.Flags{})

	// This test validates the fromCache parameter is handled correctly
	engine := &Engine{
		Config: &config.Config{
			Blocks: []*config.Block{},
		},
		Env:       env,
		allBlocks: []*config.Block{},
	}

	// Just verify it doesn't panic - full integration tested elsewhere
	_ = engine.primaryInternal(true)
}

func TestRenderBlockFromCache(t *testing.T) {
	// This test validates renderBlockFromCache handles segments correctly
	segment := &config.Segment{
		Type:    "text",
		Enabled: false,
	}

	block := &config.Block{
		Type:      config.Prompt,
		Alignment: config.Left,
		Segments:  []*config.Segment{segment},
	}

	engine := &Engine{
		Config: &config.Config{},
	}

	terminal.Init(shell.PWSH)

	// Should not render when segment is disabled and not forced
	result := engine.renderBlockFromCache(block, false)
	assert.False(t, result, "Block should not render with disabled segment")
}

func TestSegmentPendingState(t *testing.T) {
	env := new(mock.Environment)
	env.On("Shell").Return(shell.PWSH)
	env.On("Flags").Return(&runtime.Flags{})

	template.Cache = &cache.Template{
		Segments: maps.NewConcurrent[any](),
	}
	template.Init(env, nil, nil)

	segment := &config.Segment{
		Type:     "text",
		Pending:  true,
		Template: "test template",
	}
	err := segment.MapSegmentWithWriter(env)
	require.NoError(t, err)

	// Render with pending state - should show "..."
	segment.Render(0, true)
	text := segment.Text()
	assert.Equal(t, "...", text, "Pending segment should show ...")

	// After completion
	segment.Pending = false
	segment.Render(0, true)
	text = segment.Text()
	assert.NotEqual(t, "...", text, "Non-pending segment should show actual content")
}

// Helper function to collect all output from a channel with timeout
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

func TestStreamingWithTimeout(t *testing.T) {
	engine := &Engine{
		streamingResults: make(chan *config.Segment, 10),
	}

	segment := &config.Segment{
		Type:    "test",
		Timeout: 10,
	}

	// Pre-register segment as pending (this happens in writeSegmentsConcurrently in real code)
	engine.pendingSegments.Store(segment.Name(), true)

	// Test that timeout with streaming enabled marks segment as pending
	done := make(chan bool)

	go func() {
		time.Sleep(50 * time.Millisecond)
		close(done)
	}()

	engine.trackPendingSegment(segment, done)

	// Verify pending state
	_, isPending := engine.pendingSegments.Load(segment.Name())
	require.True(t, isPending, "Segment should be pending")

	// Wait for completion
	select {
	case <-engine.streamingResults:
		// Success
	case <-time.After(200 * time.Millisecond):
		t.Error("Timeout waiting for segment completion")
	}

	// Verify no longer pending
	_, isPending = engine.pendingSegments.Load(segment.Name())
	assert.False(t, isPending, "Segment should no longer be pending")
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

func TestStreamPrimary_FullFlow_WithRendering(t *testing.T) {
	env := setupStreamingTestEnv()

	// Create segments with different speeds
	fastSegment := &config.Segment{
		Type:       "text",
		Template:   "FAST",
		Foreground: "#ffffff",
		Background: "#000000",
	}

	slowSegment := &config.Segment{
		Type:       "text",
		Template:   "SLOW",
		Pending:    true, // Initially pending
		Foreground: "#ffffff",
		Background: "#000000",
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
		Env:              env,
		streamingResults: make(chan *config.Segment, 10),
	}

	// Map segment writers
	err := fastSegment.MapSegmentWithWriter(env)
	require.NoError(t, err)
	err = slowSegment.MapSegmentWithWriter(env)
	require.NoError(t, err)

	// Track slow segment as pending
	engine.pendingSegments.Store(slowSegment.Name(), true)

	// Start streaming
	out := engine.StreamPrimary()

	// Simulate slow segment completion after delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		slowSegment.Pending = false
		engine.notifySegmentCompletion(slowSegment)
	}()

	// Collect all prompts
	prompts := collectChannelOutput(out, 200*time.Millisecond)

	// Should have at least 2 prompts: initial (with "...") and final (with "SLOW")
	assert.GreaterOrEqual(t, len(prompts), 1, "Should have at least initial prompt")

	// First prompt should contain "..." for pending segment
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

	// Block 1: Fast segment
	fast1 := &config.Segment{
		Type:     "text",
		Template: "FAST1",
	}

	// Block 2: Slow segment
	slow1 := &config.Segment{
		Type:     "text",
		Template: "SLOW1",
		Pending:  true,
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
		Env:              env,
		streamingResults: make(chan *config.Segment, 10),
	}

	// Map segments
	require.NoError(t, fast1.MapSegmentWithWriter(env))
	require.NoError(t, slow1.MapSegmentWithWriter(env))
	require.NoError(t, fast2.MapSegmentWithWriter(env))

	// Track slow segment
	engine.pendingSegments.Store(slow1.Name(), true)

	// Start streaming
	out := engine.StreamPrimary()

	// Simulate completion
	go func() {
		time.Sleep(50 * time.Millisecond)
		slow1.Pending = false
		engine.notifySegmentCompletion(slow1)
	}()

	prompts := collectChannelOutput(out, 200*time.Millisecond)

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

func TestStreamPrimary_NoStreamingResults_Channel(t *testing.T) {
	engine := setupBasicStreamingTestEnv()

	// Engine without streamingResults channel (edge case)
	// No streamingResults channel set

	// Should not panic
	out := engine.StreamPrimary()
	prompts := collectChannelOutput(out, 100*time.Millisecond)

	assert.Len(t, prompts, 2, "Should get the initial prompt and transient record with no pending segments")
}

// TestStreamPrimary_RaceConditionFix validates that the streaming loop
// correctly handles segments that complete after Primary() but before/during
// the counting phase. This tests the fix for the race where pendingCount
// could get out of sync with actual pending segments.
func TestStreamPrimary_RaceConditionFix(t *testing.T) {
	env := new(mock.Environment)
	env.On("Pwd").Return("/test")
	env.On("Home").Return("/home")
	env.On("Shell").Return(shell.PWSH)
	env.On("Flags").Return(&runtime.Flags{Streaming: true})
	env.On("CursorPosition").Return(1, 1)
	env.On("StatusCodes").Return(0, "0")

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
		Env:              env,
		streamingResults: make(chan *config.Segment, 10),
	}

	// Create three segments, simulating the race scenario:
	// - segmentA: Completes quickly after Primary()
	// - segmentB: Completes during loop
	// - segmentC: Completes last
	segmentA := &config.Segment{Type: "test-a", Pending: true}
	segmentB := &config.Segment{Type: "test-b", Pending: true}
	segmentC := &config.Segment{Type: "test-c", Pending: true}

	// Pre-register all three as pending (simulates timeout during Primary())
	engine.pendingSegments.Store(segmentA.Name(), true)
	engine.pendingSegments.Store(segmentB.Name(), true)
	engine.pendingSegments.Store(segmentC.Name(), true)

	// Simulate segmentA completing immediately after Primary() but before countPendingSegments()
	// This is the race condition - notification sent but segment removed from map
	go func() {
		// Small delay to ensure StreamPrimary has been called but before counting
		time.Sleep(5 * time.Millisecond)
		segmentA.Pending = false
		engine.notifySegmentCompletion(segmentA)
	}()

	// Simulate segmentB and segmentC completing during the loop
	go func() {
		time.Sleep(30 * time.Millisecond)
		segmentB.Pending = false
		engine.notifySegmentCompletion(segmentB)
	}()

	go func() {
		time.Sleep(50 * time.Millisecond)
		segmentC.Pending = false
		engine.notifySegmentCompletion(segmentC)
	}()

	// Start streaming
	out := engine.StreamPrimary()

	// Collect all prompts with sufficient timeout
	prompts := collectChannelOutput(out, 200*time.Millisecond)

	// Only count primary prompt records, transient records don't reflect segment updates
	var primaryPrompts []string
	for _, record := range prompts {
		if !strings.HasPrefix(record, TransientMarker) {
			primaryPrompts = append(primaryPrompts, record)
		}
	}

	// With the fix, we should receive updates for all three segments
	// Initial prompt + 3 updates (A, B, C) = 4 total
	// Without the fix, we might only get Initial + 2 updates and exit early
	assert.GreaterOrEqual(t, len(primaryPrompts), 3, "Should receive updates for all pending segments")

	// Verify all segments were properly cleaned up
	count := engine.countPendingSegments()
	assert.Equal(t, 0, count, "All pending segments should be cleared")
}

// TestStreamPrimary_Abort_StopsRenderingAndDrains validates that Abort() stops
// the producer from emitting further records and blocks until the producer
// goroutine has fully exited, even when a segment completes (and tries to
// notify) after the abort was issued - that late notification must not panic
// or deadlock.
func TestStreamPrimary_Abort_StopsRenderingAndDrains(t *testing.T) {
	env := setupStreamingTestEnv()

	slowSegment := &config.Segment{
		Type:       "text",
		Template:   "SLOW",
		Pending:    true,
		Foreground: "#ffffff",
		Background: "#000000",
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

	require.NoError(t, slowSegment.MapSegmentWithWriter(env))
	engine.pendingSegments.Store(slowSegment.Name(), true)

	out := engine.StreamPrimary()

	// Consume the initial prompt + transient record before aborting, mirroring
	// how a server would start draining a cycle immediately.
	<-out
	<-out

	// Complete the segment AFTER abort is issued: the drain loop must consume
	// this notification without rendering (and without blocking the sender).
	go func() {
		time.Sleep(20 * time.Millisecond)
		slowSegment.Pending = false
		engine.notifySegmentCompletion(slowSegment)
	}()

	engine.Abort()

	// Abort must not return until the producer has exited, which implies out
	// and streamingResults are both closed.
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
	time.Sleep(40 * time.Millisecond)
}

// TestStreamPrimary_NoStreamingTimeout_ChannelCloses guards against the
// pending-segment leak: with the Streaming flag set but no top-level
// "streaming" timeout in the config (Config.Streaming == 0), every segment
// used to be pre-registered in pendingSegments and never cleaned up (the
// cleanup in writeSegmentsConcurrently only ran for Timeout > 0), so
// countPendingSegments never reached 0, the producer goroutine waited on
// streamingResults forever, and the stream CLI command never exited.
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
			// Streaming (the global segment timeout) deliberately left at 0
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

	require.NoError(t, segment.MapSegmentWithWriter(env))

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
				assert.Equal(t, 0, engine.countPendingSegments(), "no segment may stay pending without a streaming timeout")
				return
			}

			prompts = append(prompts, record)
		case <-deadline:
			t.Fatal("output channel never closed: segments leaked into pendingSegments without a streaming timeout")
		}
	}
}

// TestStreamPrimary_Abort_BeforeAnyRender validates that Abort can be called
// (and is a safe no-op) when no cycle has ever been started.
func TestStreamPrimary_Abort_NoCycleStarted(t *testing.T) {
	engine := &Engine{}
	assert.NotPanics(t, func() { engine.Abort() })
}

// TestStreamPrimary_Abort_ThenNewCycleWorks validates that a fresh
// StreamPrimary cycle works correctly after a previous cycle was aborted -
// this is the serialization guarantee the serve command depends on.
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
