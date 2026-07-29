package cli

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Does not run Execute/Enabled(), so the caller controls Enabled and the writer's
// fields directly, keeping the test hermetic (no real environment probing).
func newRecordedSessionSegment(t *testing.T, alias string) *config.Segment {
	t.Helper()

	env := new(mock.Environment)
	env.On("Getenv", "SSH_CONNECTION").Return("")
	env.On("Getenv", "SSH_CLIENT").Return("")

	segment := &config.Segment{Type: config.SESSION, Alias: alias}
	require.NoError(t, segment.MapSegmentWithWriter(env))
	segment.Enabled = true

	return segment
}

func TestBuildDataDocument_EnvSectionDropsInternalKeysKeepsRest(t *testing.T) {
	template.Cache = &cache.Template{
		Segments: maps.NewConcurrent[any](),
		SimpleTemplate: cache.SimpleTemplate{
			PWD:      "/home/jan",
			UserName: "jan",
			Var:      maps.Simple[any]{"foo": "bar"},
		},
	}

	cfg := &config.Config{}

	doc, err := buildDataDocument(cfg)
	require.NoError(t, err)

	var parsed map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(doc, &parsed))

	require.Contains(t, parsed, "env")

	var env map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(parsed["env"], &env))

	assert.NotContains(t, env, "SegmentsCache", "internal cache plumbing must not be recorded")
	assert.NotContains(t, env, "Var", "config vars are already covered by the config's own var section")
	assert.Contains(t, env, "PWD")
	assert.Contains(t, env, "UserName")
}

// unwrapRecorded decodes a segment's raw message as the RecordedSegment
// envelope {"enabled":...,"data":...} that buildDataDocument now always writes,
// and returns its enabled flag plus the inner writer data unmarshaled into out.
func unwrapRecorded(t *testing.T, raw json.RawMessage, out any) bool {
	t.Helper()

	var envelope config.RecordedSegment
	require.NoError(t, json.Unmarshal(raw, &envelope))
	require.NoError(t, json.Unmarshal(envelope.Data, out))

	return envelope.Enabled
}

func TestBuildDataDocument_RecordsEveryConfiguredSegmentWithItsEnabledState(t *testing.T) {
	template.Cache = &cache.Template{Segments: maps.NewConcurrent[any]()}

	enabled := newRecordedSessionSegment(t, "")
	enabled.Writer().(*segments.Session).SSHSession = true

	disabled := newRecordedSessionSegment(t, "disabled-alias")
	disabled.Enabled = false

	noWriter := &config.Segment{Type: config.TEXT, Alias: "no-writer", Enabled: true}

	cfg := &config.Config{
		Blocks: []*config.Block{
			{Segments: []*config.Segment{enabled, disabled, noWriter}},
		},
	}

	doc, err := buildDataDocument(cfg)
	require.NoError(t, err)

	var parsed map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(doc, &parsed))

	require.Contains(t, parsed, "version", "the recorder must stamp a version marker so replay is hermetic")

	var version int
	require.NoError(t, json.Unmarshal(parsed["version"], &version))
	assert.Equal(t, config.DataVersion, version)

	var segs map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(parsed["segments"], &segs))

	assert.Contains(t, segs, "session")
	assert.Contains(t, segs, "disabled-alias", "a segment that wasn't enabled at record time must still be recorded, "+
		"so replay can suppress it without a live probe instead of falling through to one")
	assert.NotContains(t, segs, "no-writer", "a segment with a nil writer must be skipped")

	var session map[string]json.RawMessage
	enabled1 := unwrapRecorded(t, segs["session"], &session)
	assert.True(t, enabled1)
	assert.JSONEq(t, `true`, string(session["SSHSession"]))

	var disabledSession map[string]json.RawMessage
	enabled2 := unwrapRecorded(t, segs["disabled-alias"], &disabledSession)
	assert.False(t, enabled2, "the recorded enabled flag must reflect the segment's own state, not force true")
}

func TestBuildDataDocument_CollisionWarnsAndLastWriterWins(t *testing.T) {
	template.Cache = &cache.Template{Segments: maps.NewConcurrent[any]()}

	first := newRecordedSessionSegment(t, "")
	first.Writer().(*segments.Session).SSHSession = false

	second := newRecordedSessionSegment(t, "")
	second.Writer().(*segments.Session).SSHSession = true

	cfg := &config.Config{
		Blocks: []*config.Block{
			{Segments: []*config.Segment{first, second}},
		},
	}

	stderrR, stderrW, err := os.Pipe()
	require.NoError(t, err)

	originalStderr := os.Stderr
	os.Stderr = stderrW

	doc, docErr := buildDataDocument(cfg)

	os.Stderr = originalStderr
	require.NoError(t, stderrW.Close())

	var warning []byte
	buf := make([]byte, 4096)
	n, _ := stderrR.Read(buf)
	warning = buf[:n]
	_ = stderrR.Close()

	require.NoError(t, docErr)
	assert.Contains(t, string(warning), "session", "collision warning should name the colliding key")

	var parsed map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(doc, &parsed))

	var segs map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(parsed["segments"], &segs))

	var session map[string]json.RawMessage
	unwrapRecorded(t, segs["session"], &session)
	assert.JSONEq(t, `true`, string(session["SSHSession"]), "the last segment sharing the key should win")
}

func TestDataCmd_Flags(t *testing.T) {
	flag := dataCmd.Flags().Lookup("output")
	require.NotNil(t, flag)
	assert.Equal(t, "o", flag.Shorthand)
}
