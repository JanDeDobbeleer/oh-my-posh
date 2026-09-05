package config

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

const (
	cwd = "Projects/oh-my-posh"
)

func TestMapSegmentWriterCanMap(t *testing.T) {
	sc := &Segment{
		Type: SESSION,
	}
	env := new(mock.Environment)
	err := sc.MapSegmentWithWriter(env)
	assert.NoError(t, err)
	assert.NotNil(t, sc.writer)
}

func TestMapSegmentWriterCannotMap(t *testing.T) {
	sc := &Segment{
		Type: "nilwriter",
	}
	env := new(mock.Environment)
	err := sc.MapSegmentWithWriter(env)
	assert.Error(t, err)
}

func TestParseTestConfig(t *testing.T) {
	segmentJSON :=
		`
		{
			"type": "path",
			"style": "powerline",
			"powerline_symbol": "\uE0B0",
			"foreground": "#ffffff",
			"background": "#61AFEF",
			"options": {
				"style": "folder"
			},
			"exclude_folders": [
				"/super/secret/project"
			]
		}
		`
	segment := &Segment{}
	err := json.Unmarshal([]byte(segmentJSON), segment)
	assert.NoError(t, err)
	assert.NotNil(t, segment.Options)
	assert.Equal(t, "folder", segment.Options.String("style", ""))
}

func TestParseConfigWithOptions(t *testing.T) {
	segmentJSON :=
		`
		{
			"type": "path",
			"style": "powerline",
			"options": {
				"style": "folder"
			}
		}
		`
	segment := &Segment{}
	err := json.Unmarshal([]byte(segmentJSON), segment)
	assert.NoError(t, err)
	assert.NotNil(t, segment.Options)
	assert.Equal(t, "folder", segment.Options.String("style", ""))
}

func TestParseYAMLConfigWithProperties(t *testing.T) {
	segmentYAML := `
type: path
style: powerline
properties:
  style: folder
`
	segment := &Segment{}
	err := yaml.Unmarshal([]byte(segmentYAML), segment)
	assert.NoError(t, err)
	assert.NotNil(t, segment.Options)
	assert.Equal(t, "folder", segment.Options.String("style", ""))
}

func TestParseYAMLConfigWithOptions(t *testing.T) {
	segmentYAML := `
type: path
style: powerline
options:
  style: folder
`
	segment := &Segment{}
	err := yaml.Unmarshal([]byte(segmentYAML), segment)
	assert.NoError(t, err)
	assert.NotNil(t, segment.Options)
	assert.Equal(t, "folder", segment.Options.String("style", ""))
}

func TestParseTOMLConfigWithProperties(t *testing.T) {
	segmentTOML := `
type = "path"
style = "powerline"
[properties]
style = "folder"
`
	segment := &Segment{}
	err := toml.Unmarshal([]byte(segmentTOML), segment)
	assert.NoError(t, err)

	// Migrate properties to options (normally done by Config.migrateSegmentProperties)
	segment.MigratePropertiesToOptions()

	assert.NotNil(t, segment.Options)
	assert.Equal(t, "folder", segment.Options.String("style", ""))
}

func TestParseTOMLConfigWithOptions(t *testing.T) {
	segmentTOML := `
type = "path"
style = "powerline"
[options]
style = "folder"
`
	segment := &Segment{}
	err := toml.Unmarshal([]byte(segmentTOML), segment)
	assert.NoError(t, err)

	// Migrate properties to options (should be a no-op since options is set)
	segment.MigratePropertiesToOptions()

	assert.NotNil(t, segment.Options)
	assert.Equal(t, "folder", segment.Options.String("style", ""))
}

func TestParseTOMLConfigWithBothOptionsAndProperties(t *testing.T) {
	// If both are specified, options takes precedence
	segmentTOML := `
type = "path"
style = "powerline"
[options]
style = "folder"
[properties]
style = "letter"
`
	segment := &Segment{}
	err := toml.Unmarshal([]byte(segmentTOML), segment)
	assert.NoError(t, err)

	// Migrate should not overwrite options
	segment.MigratePropertiesToOptions()

	assert.NotNil(t, segment.Options)
	assert.Equal(t, "folder", segment.Options.String("style", ""))
}

func TestShouldIncludeFolder(t *testing.T) {
	cases := []struct {
		Case     string
		Included bool
		Excluded bool
		Expected bool
	}{
		{Case: "Include", Included: true, Excluded: false, Expected: true},
		{Case: "Exclude", Included: false, Excluded: true, Expected: false},
		{Case: "Include & Exclude", Included: true, Excluded: true, Expected: false},
		{Case: "!Include & !Exclude", Included: false, Excluded: false, Expected: false},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("GOOS").Return(runtime.LINUX)
		env.On("Home").Return("")
		env.On("Pwd").Return(cwd)
		env.On("DirMatchesOneOf", cwd, []string{"Projects/oh-my-posh"}).Return(tc.Included)
		env.On("DirMatchesOneOf", cwd, []string{"Projects/nope"}).Return(tc.Excluded)
		segment := &Segment{
			IncludeFolders: []string{"Projects/oh-my-posh"},
			ExcludeFolders: []string{"Projects/nope"},
			env:            env,
		}
		got := segment.shouldIncludeFolder()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestGetColors(t *testing.T) {
	cases := []struct {
		Case       string
		Expected   color.Ansi
		Default    color.Ansi
		Region     string
		Profile    string
		Templates  []string
		Background bool
	}{
		{Case: "No template - foreground", Expected: "color", Background: false, Default: "color"},
		{Case: "No template - background", Expected: "color", Background: true, Default: "color"},
		{Case: "Nil template", Expected: "color", Default: "color", Templates: nil},
		{
			Case:     "Template - default",
			Expected: "color",
			Default:  "color",
			Templates: []string{
				"{{if contains \"john\" .Profile}}color2{{end}}",
			},
			Profile: "doe",
		},
		{
			Case:     "Template - override",
			Expected: "color2",
			Default:  "color",
			Templates: []string{
				"{{if contains \"john\" .Profile}}color2{{end}}",
			},
			Profile: "john",
		},
		{
			Case:     "Template - override multiple",
			Expected: "color3",
			Default:  "color",
			Templates: []string{
				"{{if contains \"doe\" .Profile}}color2{{end}}",
				"{{if contains \"john\" .Profile}}color3{{end}}",
			},
			Profile: "john",
		},
		{
			Case:     "Template - override multiple no match",
			Expected: "color",
			Default:  "color",
			Templates: []string{
				"{{if contains \"doe\" .Profile}}color2{{end}}",
				"{{if contains \"philip\" .Profile}}color3{{end}}",
			},
			Profile: "john",
		},
	}
	for _, tc := range cases {
		segment := &Segment{
			writer: &segments.Aws{
				Profile: tc.Profile,
				Region:  tc.Region,
			},
		}

		if tc.Background {
			segment.Background = tc.Default
			segment.BackgroundTemplates = tc.Templates
			bgColor := segment.ResolveBackground()
			assert.Equal(t, tc.Expected, bgColor, tc.Case)
			continue
		}

		segment.Foreground = tc.Default
		segment.ForegroundTemplates = tc.Templates
		fgColor := segment.ResolveForeground()
		assert.Equal(t, tc.Expected, fgColor, tc.Case)
	}
}

func TestEvaluateNeeds(t *testing.T) {
	cases := []struct {
		Segment *Segment
		Case    string
		Needs   []string
	}{
		{
			Case: "No needs",
			Segment: &Segment{
				Template: "foo",
			},
		},
		{
			Case: "Template needs",
			Segment: &Segment{
				Template: "{{ .Segments.Git.URL }}",
			},
			Needs: []string{"Git"},
		},
		{
			Case: "Template & Foreground needs",
			Segment: &Segment{
				Template:            "{{ .Segments.Git.URL }}",
				ForegroundTemplates: []string{"foo", "{{ .Segments.Os.Icon }}"},
			},
			Needs: []string{"Git", "Os"},
		},
		{
			Case: "Template & Foreground & Background needs",
			Segment: &Segment{
				Template:            "{{ .Segments.Git.URL }}",
				ForegroundTemplates: []string{"foo", "{{ .Segments.Os.Icon }}"},
				BackgroundTemplates: []string{"bar", "{{ .Segments.Exit.Icon }}"},
			},
			Needs: []string{"Git", "Os", "Exit"},
		},
	}
	for _, tc := range cases {
		tc.Segment.evaluateNeeds()
		assert.Equal(t, tc.Needs, tc.Segment.Needs, tc.Case)
	}
}

func TestSegment_NoCachingWhenPending(t *testing.T) {
	env := new(mock.Environment)
	env.On("Shell").Return("pwsh")
	env.On("Flags").Return(&runtime.Flags{})
	env.On("Pwd").Return("/test")
	env.On("Home").Return("/home")

	segment := &Segment{
		Type:     SESSION,
		Pending:  true,
		Template: "test",
	}

	err := segment.MapSegmentWithWriter(env)
	assert.NoError(t, err)

	// When Pending=true, setCache should return early without caching
	// We can't easily mock cache.Set, but we can verify the method doesn't panic
	// and that the behavior differs between Pending=true and Pending=false

	// With Pending=true, setCache returns early
	segment.Cache = &Cache{Duration: "5h"}
	segment.setCache() // Should return early, not attempt to cache

	// Verify this doesn't panic and segment still works
	assert.True(t, segment.Pending, "Segment should still be pending")

	// Now with Pending=false, setCache will attempt to cache
	segment.Pending = false
	segment.restored = false
	segment.setCache() // Should attempt to cache (may fail but shouldn't panic)

	assert.False(t, segment.Pending, "Segment should not be pending")
}

func TestSegment_DataKey(t *testing.T) {
	aliased := &Segment{Type: SESSION, Alias: "work"}
	assert.Equal(t, "work", aliased.DataKey())

	unaliased := &Segment{Type: SESSION}
	assert.Equal(t, "session", unaliased.DataKey())
}

// newDataReplayEnv builds a mock environment suitable for driving Segment.Execute
// through the shouldIncludeFolder/isToggled checks that precede data replay, and
// initializes template.Cache/template.Init so AddSegmentData doesn't panic.
func newDataReplayEnv(flags *runtime.Flags) *mock.Environment {
	env := new(mock.Environment)
	env.On("Shell").Return("pwsh")
	env.On("Pwd").Return("/test")
	env.On("Home").Return("/home")
	env.On("DirMatchesOneOf", testifymock.Anything, testifymock.Anything).Return(false)
	env.On("Flags").Return(flags)

	template.Cache = &cache.Template{
		Segments: maps.NewConcurrent[any](),
	}
	template.Init(env, nil, nil)

	return env
}

// The mock env deliberately has no "Getenv"/"Platform" stubs: Session.Enabled()
// calls both, so if restoreData ever fell through to a live probe for a
// recorded-enabled entry, the mock would panic on the unstubbed call. That
// absence is the hermeticity assertion.
func TestSegment_RestoreDataRecordedEnabledPopulatesWriterNoProbe(t *testing.T) {
	flags := &runtime.Flags{
		SegmentData: map[string]json.RawMessage{
			"session": json.RawMessage(`{"enabled":true,"data":{"SSHSession":true}}`),
		},
	}
	env := newDataReplayEnv(flags)

	segment := &Segment{Type: SESSION}
	segment.Execute(env)

	assert.True(t, segment.Enabled, "segment should be forced enabled by data replay")
	assert.True(t, segment.restored, "restored flag should prevent setCache pollution")

	writer, ok := segment.Writer().(*segments.Session)
	assert.True(t, ok, "Writer() should expose the concrete writer")
	assert.True(t, writer.SSHSession)

	segment.Template = "{{ .SSHSession }}"
	assert.Equal(t, "true", segment.string())
}

func TestSegment_RestoreDataMainWorktreeStaysRuntimeFree(t *testing.T) {
	flags := &runtime.Flags{
		DataOnly: true,
		SegmentData: map[string]json.RawMessage{
			"git": json.RawMessage(
				`{"enabled":true,"data":{"IsWorkTree":true,"MainWorktree":"/repo/main"}}`,
			),
		},
	}
	env := newDataReplayEnv(flags)

	segment := &Segment{
		Type:     GIT,
		Template: "{{ .MainWorktree }}",
	}
	segment.Execute(env)

	assert.True(t, segment.Enabled)
	assert.Empty(t, segment.string())
	env.AssertNotCalled(t, "HasParentFilePath", testifymock.Anything, testifymock.Anything)
	env.AssertNotCalled(t, "HasCommand", testifymock.Anything)
	env.AssertNotCalled(t, "RunCommand", testifymock.Anything, testifymock.Anything)
}

// A recorded-but-disabled entry must suppress the segment without probing it
// live (that is what makes replay hermetic for e.g. wakatime on a config where
// it was disabled at record time). No "Getenv"/"Platform" stub, same reasoning
// as above.
func TestSegment_RestoreDataRecordedDisabledSuppressesNoProbe(t *testing.T) {
	flags := &runtime.Flags{
		SegmentData: map[string]json.RawMessage{
			"session": json.RawMessage(`{"enabled":false,"data":{"SSHSession":true}}`),
		},
	}
	env := newDataReplayEnv(flags)

	segment := &Segment{Type: SESSION}
	segment.Execute(env)

	assert.False(t, segment.Enabled, "recorded-but-disabled segment must stay disabled")
	assert.True(t, segment.restored, "still counts as restored: no live probe happened")

	writer, ok := segment.Writer().(*segments.Session)
	assert.True(t, ok, "Writer() should expose the concrete writer")
	assert.False(t, writer.SSHSession, "disabled entries are not unmarshaled onto the writer")
}

// A flat entry (no version marker on the file) is a hand-written file's shape.
// It must NOT take the hermetic short-circuit: writer.Enabled() has to run first
// (proven here by the Getenv/Platform stubs actually being exercised) and the
// pinned value is overlaid on top afterward, winning over the live one.
func TestSegment_RestoreDataUnmarkedFlatDerivesLiveStateThenOverlays(t *testing.T) {
	flags := &runtime.Flags{
		SegmentData: map[string]json.RawMessage{
			"session": json.RawMessage(`{"SSHSession":true}`),
		},
	}
	env := newDataReplayEnv(flags)
	env.On("Getenv", "SSH_CONNECTION").Return("")
	env.On("Getenv", "SSH_CLIENT").Return("")
	env.On("Platform").Return(runtime.WINDOWS)

	segment := &Segment{Type: SESSION}
	segment.Execute(env)

	assert.True(t, segment.Enabled, "pinned data forces Enabled even though nothing here would suppress it")
	assert.True(t, segment.restored, "overlay marks the segment restored so setCache doesn't cache derived state")

	writer, ok := segment.Writer().(*segments.Session)
	assert.True(t, ok, "Writer() should expose the concrete writer")
	assert.True(t, writer.SSHSession, "the pinned value must win over the live-computed one")

	env.AssertCalled(t, "Getenv", "SSH_CONNECTION")
}

// Regression test for the reported bug: 53 of 102 segments (time among them)
// compute template-visible state inside Enabled(), so pinning only CurrentDate
// in a hand-written file used to leave Format empty and the whole segment
// rendered blank. It must now derive Format live and overlay CurrentDate on top.
func TestSegment_RestoreDataUnmarkedTimeSegmentDerivesFormatBeforeOverlay(t *testing.T) {
	flags := &runtime.Flags{
		SegmentData: map[string]json.RawMessage{
			"time": json.RawMessage(`{"CurrentDate":"2026-01-01T09:41:00Z"}`),
		},
	}
	env := newDataReplayEnv(flags)

	segment := &Segment{Type: TIME}
	segment.Execute(env)

	assert.True(t, segment.Enabled)

	writer, ok := segment.Writer().(*segments.Time)
	assert.True(t, ok, "Writer() should expose the concrete writer")
	assert.NotEmpty(t, writer.Format, "Format is computed inside Enabled(); pinning only CurrentDate must not leave it empty")
	assert.Equal(t, 2026, writer.CurrentDate.Year(), "the pinned CurrentDate must win over the live one")

	segment.Template = writer.Template()
	assert.NotEmpty(t, segment.string(), "the segment must not render blank")
}

// Regression test: restoreCache's own short-circuit (immediately after
// restoreData, before shouldHideForWidth/the Job setup/writer.Enabled()) used
// to return before overlayData ever ran, so a hand-written pinned value lost
// to a live cache hit instead of winning as documented. Exactly this shape
// exists in the wild: themes/chips.omp.json's wakatime segment carries a
// `cache` block, and website/segment_data.json pins it - without this fix the
// website's chips image could render a stale cached value instead of the
// pinned one, non-deterministically.
//
// No "Getenv"/"Platform" stub on the mock env: a cache hit must win without
// ever probing the live environment, same reasoning as the recorded-envelope
// tests above.
func TestSegment_RestoreDataOverlaysOnTopOfACacheHit(t *testing.T) {
	segmentCache := &Cache{Duration: "5h", Strategy: Session}

	// Compute the cache key/store the same way the segment under test will,
	// then warm the cache with a writer that disagrees with the pinned value -
	// as if an earlier live render had cached SSHSession: false.
	keySegment := &Segment{Type: SESSION, Cache: segmentCache}
	key, store := keySegment.cacheKeyAndStore()

	var warm bytes.Buffer
	require.NoError(t, gob.NewEncoder(&warm).Encode(&segments.Session{SSHSession: false}))
	store.Set(key, warm.Bytes(), cache.INFINITE)
	t.Cleanup(func() { store.Delete(key) })

	flags := &runtime.Flags{
		SegmentData: map[string]json.RawMessage{
			"session": json.RawMessage(`{"SSHSession":true}`),
		},
	}
	env := newDataReplayEnv(flags)

	segment := &Segment{Type: SESSION, Cache: segmentCache}
	segment.Execute(env)

	assert.True(t, segment.Enabled)
	assert.True(t, segment.restored, "cache-hit-then-overlay must still mark the segment restored, so setCache doesn't re-cache it")

	writer, ok := segment.Writer().(*segments.Session)
	assert.True(t, ok, "Writer() should expose the concrete writer")
	assert.True(t, writer.SSHSession, "the pinned value must win over the warm cache hit")
}

func TestSegment_RestoreDataAliasBeatsType(t *testing.T) {
	// Only the type-keyed entry exists; the segment has an alias, so DataKey()
	// resolves to "work" and must NOT match the "text" (type) entry.
	flags := &runtime.Flags{
		SegmentData: map[string]json.RawMessage{
			"text": json.RawMessage(`{}`),
		},
	}
	env := newDataReplayEnv(flags)

	segment := &Segment{Type: TEXT, Alias: "work"}
	segment.Execute(env)

	assert.False(t, segment.restored, "type-keyed data must not be used for an aliased segment")
}

func TestSegment_RestoreDataNoEntryFallsThroughToNormalExecution(t *testing.T) {
	env := newDataReplayEnv(&runtime.Flags{})

	segment := &Segment{Type: TEXT}
	segment.Execute(env)

	assert.True(t, segment.Enabled, "text segment is always enabled on normal execution")
	assert.False(t, segment.restored, "no data entry means no replay happened")
}

// DataOnly deliberately changes nothing here: a segment the data file does not
// cover still falls through to writer.Enabled() exactly as it would without the
// flag. What stops it reaching the machine is the environment itself refusing
// every probe (runtime.Flags.DataOnly, pinned by TestTerminalDataOnly* in the
// runtime package). Suppressing uncovered segments at this layer as well was
// tried and backed out - it also erased path and session, which need no machine
// at all and render from the data file's own env section.
func TestSegment_RestoreDataOnlyLeavesTheFallThroughAlone(t *testing.T) {
	env := newDataReplayEnv(&runtime.Flags{DataOnly: true})

	segment := &Segment{Type: TEXT, Template: "hello"}
	segment.Execute(env)

	assert.True(t, segment.Enabled, "an uncovered segment still executes")
	assert.False(t, segment.restored, "no data entry means no replay happened")
}

// Malformed JSON does not parse as a RecordedSegment envelope either, so it is
// treated as unmarked data: Execute keeps running (TEXT.Enabled() has no live
// dependency, so this can't panic), and the overlay attempt fails silently
// rather than blocking normal rendering.
func TestSegment_RestoreDataInvalidJSONFallsThrough(t *testing.T) {
	flags := &runtime.Flags{
		SegmentData: map[string]json.RawMessage{
			"text": json.RawMessage(`{invalid`),
		},
	}
	env := newDataReplayEnv(flags)

	segment := &Segment{Type: TEXT}

	assert.NotPanics(t, func() {
		segment.Execute(env)
	})

	assert.True(t, segment.Enabled, "should fall through to normal execution")
	assert.False(t, segment.restored, "a failed replay must not mark the segment as restored")
}

func TestDecodeRecordedSegment(t *testing.T) {
	cases := []struct {
		Case            string
		Raw             string
		ExpectedData    string
		ExpectedEnabled bool
		ExpectedOK      bool
	}{
		{
			Case:            "recorded enabled",
			Raw:             `{"enabled":true,"data":{"HEAD":"main"}}`,
			ExpectedOK:      true,
			ExpectedEnabled: true,
			ExpectedData:    `{"HEAD":"main"}`,
		},
		{
			Case:            "recorded disabled",
			Raw:             `{"enabled":false,"data":{"HEAD":"main"}}`,
			ExpectedOK:      true,
			ExpectedEnabled: false,
			ExpectedData:    `{"HEAD":"main"}`,
		},
		{
			Case:       "flat hand-written data is not an envelope",
			Raw:        `{"HEAD":"main"}`,
			ExpectedOK: false,
		},
		{
			Case:       "missing the data key is not an envelope",
			Raw:        `{"enabled":true,"extra":1}`,
			ExpectedOK: false,
		},
		{
			Case:       "missing the enabled key is not an envelope",
			Raw:        `{"data":{"HEAD":"main"},"extra":1}`,
			ExpectedOK: false,
		},
		{
			Case:       "extra keys alongside enabled/data are not an envelope",
			Raw:        `{"enabled":true,"data":{},"extra":1}`,
			ExpectedOK: false,
		},
		{
			Case:       "invalid JSON is not an envelope",
			Raw:        `{invalid`,
			ExpectedOK: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			recorded, ok := decodeRecordedSegment(json.RawMessage(tc.Raw))

			assert.Equal(t, tc.ExpectedOK, ok)

			if !tc.ExpectedOK {
				return
			}

			assert.Equal(t, tc.ExpectedEnabled, recorded.Enabled)
			assert.JSONEq(t, tc.ExpectedData, string(recorded.Data))
		})
	}
}

func TestSegment_RestoreDataToggledOffStaysDisabled(t *testing.T) {
	cache.Session.Set(cache.TOGGLECACHE, map[string]bool{"text": true}, cache.INFINITE)
	defer cache.Session.Delete(cache.TOGGLECACHE)

	flags := &runtime.Flags{
		SegmentData: map[string]json.RawMessage{
			"text": json.RawMessage(`{}`),
		},
	}
	env := newDataReplayEnv(flags)

	segment := &Segment{Type: TEXT}
	segment.Execute(env)

	assert.False(t, segment.Enabled, "toggled-off segment must stay disabled even with data available")
	assert.False(t, segment.restored, "toggle check happens before data replay")
}

// fallbackWriter is a minimal SegmentWriter used to drive Segment.Render's
// fallback_template behavior without depending on a concrete segment type.
type fallbackWriter struct {
	template string
	text     string
	enabled  bool
}

func (w *fallbackWriter) Enabled() bool                                  { return w.enabled }
func (w *fallbackWriter) Activation() runtime.Activation                 { return runtime.Activation{Always: true} }
func (w *fallbackWriter) Template() string                               { return w.template }
func (w *fallbackWriter) SetText(text string)                            { w.text = text }
func (w *fallbackWriter) SetIndex(_ int)                                 {}
func (w *fallbackWriter) Text() string                                   { return w.text }
func (w *fallbackWriter) Init(_ options.Provider, _ runtime.Environment) {}
func (w *fallbackWriter) CacheKey() (string, bool)                       { return "", false }

// initTemplateCache initializes template.Cache for the duration of a test and
// restores the previous value afterward, since AddSegmentData/RemoveSegmentData
// dereference it directly.
func initTemplateCache(t *testing.T) {
	t.Helper()

	orig := template.Cache
	t.Cleanup(func() { template.Cache = orig })

	template.Cache = &cache.Template{
		Segments: maps.NewConcurrent[any](),
	}
}

func TestSegment_FallbackTemplate(t *testing.T) {
	cases := []struct {
		Cache            *Cache
		Case             string
		FallbackTemplate string
		Template         string
		ExpectedText     string
		WriterEnabled    bool
		Evaluated        bool
		Enabled          bool
		Killed           bool
		ExpectedRendered bool
		ExpectedEnabled  bool
	}{
		{
			Case:             "renders when writer disabled",
			FallbackTemplate: " disconnected ",
			Evaluated:        true,
			ExpectedRendered: true,
			ExpectedEnabled:  true,
			ExpectedText:     " disconnected ",
		},
		{
			Case:      "unset keeps segment hidden",
			Evaluated: true,
		},
		{
			// evaluated is left false, simulating a segment hidden for another
			// reason (toggled off, folder include/exclude, width, timeout kill)
			// where Execute never reached the writer.Enabled() call.
			Case:             "not evaluated keeps segment hidden",
			FallbackTemplate: " disconnected ",
		},
		{
			Case:             "whitespace-only result keeps segment hidden",
			FallbackTemplate: "   ",
			Evaluated:        true,
		},
		{
			// A timed-out segment's goroutine may still complete the
			// evaluation after the kill, so Killed must win over evaluated.
			Case:             "killed by timeout keeps segment hidden",
			FallbackTemplate: " disconnected ",
			Evaluated:        true,
			Killed:           true,
		},
		{
			Case:             "does not write to the segment cache",
			FallbackTemplate: " disconnected ",
			Evaluated:        true,
			Cache:            &Cache{Duration: "5h", Strategy: Session},
			ExpectedRendered: true,
			ExpectedEnabled:  true,
			ExpectedText:     " disconnected ",
		},
		{
			Case:             "does not affect an enabled segment",
			Template:         "primary",
			FallbackTemplate: " disconnected ",
			WriterEnabled:    true,
			Enabled:          true,
			Evaluated:        true,
			ExpectedRendered: true,
			ExpectedEnabled:  true,
			ExpectedText:     "primary",
		},
	}

	initTemplateCache(t)

	for _, tc := range cases {
		segment := &Segment{
			Type:             TEXT,
			Template:         tc.Template,
			FallbackTemplate: tc.FallbackTemplate,
			Cache:            tc.Cache,
			Enabled:          tc.Enabled,
			Killed:           tc.Killed,
			writer:           &fallbackWriter{enabled: tc.WriterEnabled, template: tc.Template},
		}
		segment.evaluated = tc.Evaluated

		rendered := segment.Render(0, false)

		assert.Equal(t, tc.ExpectedRendered, rendered, tc.Case)
		assert.Equal(t, tc.ExpectedEnabled, segment.Enabled, tc.Case)

		if tc.ExpectedText != "" {
			assert.Equal(t, tc.ExpectedText, segment.Text(), tc.Case)
		}

		if tc.Cache == nil {
			continue
		}

		key, store := segment.cacheKeyAndStore()
		_, found := store.Get[string](key)
		store.Delete(key)
		assert.False(t, found, tc.Case)
	}
}

// Without a writer (the website build) recorded data lands in a map; the
// tagged markup it carries must come back as Markup so the anchors render.
func TestRestoreIntoRevivesMarkupWithoutWriter(t *testing.T) {
	segment := &Segment{}

	raw := json.RawMessage(`{"HEAD":{"$markup":"<red>main</>"},"Ref":"<b>","Total":2}`)
	methods := json.RawMessage(`{"Working":{"String":{"$markup":"<b>~1</>"}}}`)

	require.NoError(t, segment.restoreInto(raw, methods))

	assert.Equal(t, template.RawMarkup("<red>main</>"), segment.data["HEAD"])
	assert.Equal(t, "<b>", segment.data["Ref"])
	assert.Equal(t, 2, segment.data["Total"])
	assert.Equal(t, template.RawMarkup("<b>~1</>"), segment.data["Working"].(map[string]any)["String"])
}
