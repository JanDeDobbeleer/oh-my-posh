package config

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
)

// gateStubType is registered per test into the Segments registry so Execute
// constructs the exact writer instance the test wants to inspect.
const gateStubType SegmentType = "gatestub"

// gateStubWriter overrides Base's default Activation and records whether
// Enabled() ran, so tests can prove the gate skipped (or did not skip) the
// probe.
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

// plainStubWriter inherits the default Activation from Base.
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

func TestSegmentGate_FolderConditionMatches(t *testing.T) {
	writer := &gateStubWriter{
		activation: segments.Activation{Folders: []string{".venv"}},
	}
	registerStubWriter(t, writer)

	env := newDataReplayEnv(&runtime.Flags{})
	// the folder condition stats filepath.Join(Pwd, folder), which joins with
	// the platform separator - expect exactly what production produces on
	// either OS
	env.On("HasFolder", filepath.Join("/test", ".venv")).Return(true)

	segment := &Segment{Type: gateStubType}
	segment.Execute(env)

	assert.True(t, writer.enabledCalled, "a matching folder keeps the segment alive")
	assert.True(t, segment.Enabled)
}

func TestSegmentGate_ProjectFileConditionMatches(t *testing.T) {
	writer := &gateStubWriter{
		activation: segments.Activation{ProjectFiles: []string{".git"}},
	}
	registerStubWriter(t, writer)

	env := newDataReplayEnv(&runtime.Flags{})
	env.On("HasParentFilePath", ".git", false).Return(&runtime.FileInfo{Path: "/test/.git"}, nil)

	segment := &Segment{Type: gateStubType}
	segment.Execute(env)

	assert.True(t, writer.enabledCalled, "a found project file keeps the segment alive")
	assert.True(t, segment.Enabled)
}

func TestSegmentGate_ProjectFileConditionMisses(t *testing.T) {
	writer := &gateStubWriter{
		activation: segments.Activation{ProjectFiles: []string{".git"}},
	}
	registerStubWriter(t, writer)

	env := newDataReplayEnv(&runtime.Flags{})
	env.On("HasParentFilePath", ".git", testifymock.Anything).Return((*runtime.FileInfo)(nil), errors.New("no match at root level"))

	segment := &Segment{Type: gateStubType}
	segment.Execute(env)

	assert.False(t, writer.enabledCalled, "both search variants missed, the segment gates off")
	assert.False(t, segment.Enabled)
}

func TestSegmentGate_EnvVarConditionMatches(t *testing.T) {
	writer := &gateStubWriter{
		activation: segments.Activation{EnvVars: []string{"VIRTUAL_ENV"}},
	}
	registerStubWriter(t, writer)

	env := newDataReplayEnv(&runtime.Flags{})
	env.On("Getenv", "VIRTUAL_ENV").Return("/venv")

	segment := &Segment{Type: gateStubType}
	segment.Execute(env)

	assert.True(t, writer.enabledCalled, "a non-empty env var keeps the segment alive")
	assert.True(t, segment.Enabled)
}

func TestSegmentGate_ConditionsAreOrAcrossKinds(t *testing.T) {
	writer := &gateStubWriter{
		activation: segments.Activation{
			FileGlobs: []string{"*.uni"},
			EnvVars:   []string{"VIRTUAL_ENV"},
		},
	}
	registerStubWriter(t, writer)

	env := newDataReplayEnv(&runtime.Flags{})
	env.On("HasFiles", "*.uni").Return(false)
	env.On("Getenv", "VIRTUAL_ENV").Return("/venv")

	segment := &Segment{Type: gateStubType}
	segment.Execute(env)

	assert.True(t, writer.enabledCalled, "any matching condition of any kind keeps the segment alive")
	assert.True(t, segment.Enabled)
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

// Contract change (deliberate): a failing gate no longer forces the full
// Enabled() evaluation for a fallback template. The segment counts as
// evaluated so renderFallback runs, but against the zero-state writer.
func TestSegmentGate_FallbackTemplateRendersWithoutEnabled(t *testing.T) {
	writer := &gateStubWriter{
		activation: segments.Activation{FileGlobs: []string{"*.uni"}},
	}
	registerStubWriter(t, writer)

	env := newDataReplayEnv(&runtime.Flags{})
	env.On("HasFiles", "*.uni").Return(false)

	segment := &Segment{Type: gateStubType, FallbackTemplate: "fallback"}
	segment.Execute(env)

	assert.False(t, writer.enabledCalled, "the gate wins: Enabled() must not run for the fallback")
	assert.False(t, segment.Enabled)
	assert.True(t, segment.evaluated, "the fallback must still render, against the zero-state writer")
	assert.True(t, segment.Render(0, false), "renderFallback picks the segment up")
	assert.Equal(t, "fallback", segment.Text())
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

func TestSegmentGate_BaseDefaultIsUngated(t *testing.T) {
	writer := &plainStubWriter{}
	registerStubWriter(t, writer)

	env := newDataReplayEnv(&runtime.Flags{})

	segment := &Segment{Type: gateStubType}
	segment.Execute(env)

	assert.True(t, writer.enabledCalled, "writers inheriting Base's Activation are never gated")
	assert.True(t, segment.Enabled)
}
