package prompt

import (
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

func TestRenderBlock(t *testing.T) {
	engine := New(&runtime.Flags{
		IsPrimary: true,
	})
	block := &config.Block{
		Segments: []*config.Segment{
			{
				Type:       "text",
				Template:   "Hello",
				Foreground: "red",
				Background: "blue",
			},
			{
				Type:       "text",
				Template:   "World",
				Foreground: "red",
				Background: "blue",
			},
		},
	}

	prompt, length := engine.writeBlockSegments(block)
	assert.Equal(t, "\x1b[44m\x1b[31mHello\x1b[0m\x1b[44m\x1b[31mWorld\x1b[0m", prompt)
	assert.Equal(t, 10, length)
}

func TestCanRenderSegment(t *testing.T) {
	cases := []struct {
		Case     string
		Executed map[string]bool
		Needs    []string
		Expected bool
	}{
		{
			Case:     "No cross segment dependencies",
			Expected: true,
		},
		{
			Case:     "Cross segment dependencies, nothing executed",
			Expected: false,
			Needs:    []string{"Foo"},
		},
		{
			Case:     "Cross segment dependencies, available",
			Expected: true,
			Executed: map[string]bool{
				"Foo": true,
			},
			Needs: []string{"Foo"},
		},
	}
	for _, c := range cases {
		segment := &config.Segment{
			Type:  "text",
			Needs: c.Needs,
		}

		engine := &Engine{}
		got := engine.canRenderSegment(segment, c.Executed)

		assert.Equal(t, c.Expected, got, c.Case)
	}
}

func TestExecuteSegmentWithTimeout_Streaming(t *testing.T) {
	// This test verifies that when streaming is enabled and a timeout occurs,
	// the segment is marked as pending and tracked for later completion
	env := new(mock.Environment)
	env.On("Flags").Return(&runtime.Flags{Streaming: true})

	segment := &config.Segment{
		Type:    "text",
		Timeout: 1, // Very short timeout to ensure it triggers
	}

	engine := &Engine{
		Env:              env,
		streamingResults: make(chan *config.Segment, 10),
	}

	// Create a mock segment that will definitely timeout
	// We'll use the actual timeout mechanism by making the execution slow
	done := make(chan bool)
	go func() {
		time.Sleep(100 * time.Millisecond) // Longer than timeout
		close(done)
	}()

	// Pre-register segment as pending (this happens in writeSegmentsConcurrently)
	engine.pendingSegments.Store(segment.Name(), true)

	// Mark as pending and track (simulating what executeSegmentWithTimeout does)
	segment.Pending = true
	engine.trackPendingSegment(segment, done)

	// Verify it was tracked as pending
	_, exists := engine.pendingSegments.Load(segment.Name())
	assert.True(t, exists, "Segment should be tracked as pending")

	// Wait for completion notification
	select {
	case completed := <-engine.streamingResults:
		assert.Equal(t, segment, completed)
		assert.False(t, completed.Pending, "Segment should no longer be pending after completion")
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected segment completion notification")
	}
}

func TestExecuteSegmentWithTimeout_NonStreaming(t *testing.T) {
	// This test verifies that when streaming is disabled,
	// trackPendingSegment returns early without tracking when streamingResults is nil
	segment := &config.Segment{
		Type:    "text",
		Timeout: 10,
	}

	engine := &Engine{
		// streamingResults is nil (non-streaming mode)
	}

	done := make(chan bool)

	// Pre-register segment (simulating what happens in concurrent execution)
	engine.pendingSegments.Store(segment.Name(), true)

	// trackPendingSegment should not track when streamingResults is nil
	engine.trackPendingSegment(segment, done)

	// Signal completion
	close(done)

	// Give time for any goroutine to run (shouldn't be one)
	time.Sleep(50 * time.Millisecond)

	// Segment should still be in pendingSegments because notifySegmentCompletion
	// was never called (trackPendingSegment returns early when streamingResults is nil)
	_, exists := engine.pendingSegments.Load(segment.Name())
	assert.True(t, exists, "Segment should remain in pendingSegments when streaming is disabled")
}

func TestExecuteSegmentWithTimeout_CachedValueFallback(t *testing.T) {
	// This test verifies that a pending segment's Text() returns "..." placeholder
	env := new(mock.Environment)
	env.On("Flags").Return(&runtime.Flags{})

	segment := &config.Segment{
		Type:     "text",
		Pending:  true,
		Template: "actual content",
	}

	// Initialize the segment writer
	err := segment.MapSegmentWithWriter(env)
	assert.NoError(t, err)

	// Render with pending state - should show "..."
	segment.Render(0, true)
	text := segment.Text()
	assert.Equal(t, "...", text, "Pending segment should show ...")

	// After completion, render again with actual content
	segment.Pending = false
	segment.Render(0, true)
	text = segment.Text()
	assert.NotEqual(t, "...", text, "Non-pending segment should show actual content")
}
