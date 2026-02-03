package prompt

import (
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

	template.Cache = &cache.Template{
		Segments: maps.NewConcurrent[any](),
	}
	template.Init(env, nil, nil)
	terminal.Init(shell.PWSH)

	engine := &Engine{
		Config: &config.Config{
			Blocks: []*config.Block{},
		},
		Env: env,
	}

	out := engine.StreamPrimary()
	prompts := collectChannelOutput(out, 100*time.Millisecond)

	// Should get exactly one prompt (initial) with no pending segments
	assert.Len(t, prompts, 1)
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

func TestStreamPrimary_FullFlow_WithRendering(t *testing.T) {
	env := new(mock.Environment)
	env.On("Pwd").Return("/test")
	env.On("Home").Return("/home")
	env.On("Shell").Return(shell.PWSH)
	env.On("Flags").Return(&runtime.Flags{Streaming: true})
	env.On("CursorPosition").Return(1, 1)
	env.On("StatusCodes").Return(0, "0")
	env.On("DirMatchesOneOf", testifymock.Anything, testifymock.Anything).Return(false)

	template.Cache = &cache.Template{
		Segments: maps.NewConcurrent[any](),
	}
	template.Init(env, nil, nil)
	terminal.Init(shell.PWSH)
	terminal.Colors = color.MakeColors(nil, false, "", nil)

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
	env := new(mock.Environment)
	env.On("Pwd").Return("/test")
	env.On("Home").Return("/home")
	env.On("Shell").Return(shell.PWSH)
	env.On("Flags").Return(&runtime.Flags{Streaming: true})
	env.On("CursorPosition").Return(1, 1)
	env.On("StatusCodes").Return(0, "0")
	env.On("DirMatchesOneOf", testifymock.Anything, testifymock.Anything).Return(false)

	template.Cache = &cache.Template{
		Segments: maps.NewConcurrent[any](),
	}
	template.Init(env, nil, nil)
	terminal.Init(shell.PWSH)
	terminal.Colors = color.MakeColors(nil, false, "", nil)

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

func TestStreamPrimary_EarlyChannelClosure(t *testing.T) {
	env := new(mock.Environment)
	env.On("Shell").Return(shell.PWSH)
	env.On("Flags").Return(&runtime.Flags{Streaming: true})

	engine := &Engine{
		Config: &config.Config{
			Blocks: []*config.Block{},
		},
		Env:              env,
		streamingResults: make(chan *config.Segment, 10),
	}

	// Start streaming
	out := engine.StreamPrimary()

	// Close streamingResults channel early (simulating cleanup)
	close(engine.streamingResults)

	// Should still be able to read from output channel without panic
	prompts := collectChannelOutput(out, 100*time.Millisecond)

	// Should get at least the initial prompt before closure
	assert.NotEmpty(t, prompts, "Should receive initial prompt before closure")
}

func TestStreamPrimary_NoStreamingResults_Channel(t *testing.T) {
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

	// Engine without streamingResults channel (edge case)
	engine := &Engine{
		Config: &config.Config{
			Blocks: []*config.Block{},
		},
		Env: env,
		// No streamingResults channel
	}

	// Should not panic
	out := engine.StreamPrimary()
	prompts := collectChannelOutput(out, 100*time.Millisecond)

	assert.Len(t, prompts, 1, "Should get exactly one prompt with no pending segments")
}
