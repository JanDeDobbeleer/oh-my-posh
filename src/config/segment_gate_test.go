package config

import (
	"encoding/json"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"

	"github.com/stretchr/testify/assert"
)

// gateStubType is registered per test into the Segments registry so Execute
// constructs the exact writer instance the test wants to inspect.
const gateStubType SegmentType = "gatestub"

// gateStubWriter implements segments.Activator and records whether Enabled()
// ran, so tests can prove the gate skipped (or did not skip) the probe.
type gateStubWriter struct {
	segments.Base
	activation    segments.Activation
	enabledCalled bool
}

func (w *gateStubWriter) Enabled() bool {
	w.enabledCalled = true
	return true
}

func (w *gateStubWriter) Template() string {
	return "stub"
}

func (w *gateStubWriter) Activation() segments.Activation {
	return w.activation
}

// plainStubWriter does NOT implement segments.Activator.
type plainStubWriter struct {
	segments.Base
	enabledCalled bool
}

func (w *plainStubWriter) Enabled() bool {
	w.enabledCalled = true
	return true
}

func (w *plainStubWriter) Template() string {
	return "stub"
}

func registerStubWriter(t *testing.T, writer SegmentWriter) {
	t.Helper()

	Segments[gateStubType] = func() SegmentWriter { return writer }
	t.Cleanup(func() { delete(Segments, gateStubType) })
}

func TestSegmentGate_NoMatchingFilesSkipsEnabled(t *testing.T) {
	writer := &gateStubWriter{
		activation: segments.Activation{FileGlobs: []string{"*.uni", "*.corn"}},
	}
	registerStubWriter(t, writer)

	env := newDataReplayEnv(&runtime.Flags{})
	env.On("HasFiles", "*.uni").Return(false)
	env.On("HasFiles", "*.corn").Return(false)

	segment := &Segment{Type: gateStubType}
	segment.Execute(env)

	assert.False(t, writer.enabledCalled, "gated segment must never reach Enabled()")
	assert.False(t, segment.Enabled, "gated segment stays disabled")
	assert.False(t, segment.evaluated, "gated segment must not count as evaluated")
	env.AssertCalled(t, "HasFiles", "*.uni")
}

func TestSegmentGate_MatchingFileRunsEnabled(t *testing.T) {
	writer := &gateStubWriter{
		activation: segments.Activation{FileGlobs: []string{"*.uni", "*.corn"}},
	}
	registerStubWriter(t, writer)

	env := newDataReplayEnv(&runtime.Flags{})
	env.On("HasFiles", "*.uni").Return(false)
	env.On("HasFiles", "*.corn").Return(true)

	segment := &Segment{Type: gateStubType}
	segment.Execute(env)

	assert.True(t, writer.enabledCalled, "a single matching glob keeps the segment alive")
	assert.True(t, segment.Enabled)
	assert.True(t, segment.evaluated)
}

// No "HasFiles" stub: a bypass that still evaluated the globs would panic the
// mock, so the stub's absence is part of the assertion.
func TestSegmentGate_ForceBypassesGate(t *testing.T) {
	writer := &gateStubWriter{
		activation: segments.Activation{FileGlobs: []string{"*.uni"}},
	}
	registerStubWriter(t, writer)

	env := newDataReplayEnv(&runtime.Flags{})

	segment := &Segment{Type: gateStubType, Force: true}
	segment.Execute(env)

	assert.True(t, writer.enabledCalled, "forced segments must run the full path")
	assert.True(t, segment.Enabled)
}

// Fallback templates render against the evaluated-but-disabled writer, so a
// gated skip would silently drop them; they bypass the gate entirely. No
// "HasFiles" stub, same reasoning as the Force test.
func TestSegmentGate_FallbackTemplateBypassesGate(t *testing.T) {
	writer := &gateStubWriter{
		activation: segments.Activation{FileGlobs: []string{"*.uni"}},
	}
	registerStubWriter(t, writer)

	env := newDataReplayEnv(&runtime.Flags{})

	segment := &Segment{Type: gateStubType, FallbackTemplate: "fallback"}
	segment.Execute(env)

	assert.True(t, writer.enabledCalled, "fallback-template segments must run the full path")
	assert.True(t, segment.evaluated, "evaluated semantics must stay identical for renderFallback")
}

// A hand-written (flat) data entry pins the segment on via the pendingData
// overlay, which runs after writer.Enabled(); gating it would drop the pinned
// data. No "HasFiles" stub, same reasoning as the Force test.
func TestSegmentGate_PendingDataBypassesGate(t *testing.T) {
	writer := &gateStubWriter{
		activation: segments.Activation{FileGlobs: []string{"*.uni"}},
	}
	registerStubWriter(t, writer)

	flags := &runtime.Flags{
		SegmentData: map[string]json.RawMessage{
			string(gateStubType): json.RawMessage(`{}`),
		},
	}
	env := newDataReplayEnv(flags)

	segment := &Segment{Type: gateStubType}
	segment.Execute(env)

	assert.True(t, writer.enabledCalled, "pinned segments must run the full path")
	assert.True(t, segment.Enabled)
}

func TestSegmentGate_AlwaysActivationRunsEnabled(t *testing.T) {
	writer := &gateStubWriter{
		activation: segments.Activation{Always: true, FileGlobs: []string{"*.uni"}},
	}
	registerStubWriter(t, writer)

	env := newDataReplayEnv(&runtime.Flags{})

	segment := &Segment{Type: gateStubType}
	segment.Execute(env)

	assert.True(t, writer.enabledCalled, "Always means no gate")
	assert.True(t, segment.Enabled)
}

func TestSegmentGate_ZeroActivationRunsEnabled(t *testing.T) {
	writer := &gateStubWriter{}
	registerStubWriter(t, writer)

	env := newDataReplayEnv(&runtime.Flags{})

	segment := &Segment{Type: gateStubType}
	segment.Execute(env)

	assert.True(t, writer.enabledCalled, "the zero value Activation means no gate")
	assert.True(t, segment.Enabled)
}

func TestSegmentGate_NonActivatorWriterUnchanged(t *testing.T) {
	writer := &plainStubWriter{}
	registerStubWriter(t, writer)

	env := newDataReplayEnv(&runtime.Flags{})

	segment := &Segment{Type: gateStubType}
	segment.Execute(env)

	assert.True(t, writer.enabledCalled, "writers without Activator are never gated")
	assert.True(t, segment.Enabled)
}
