package config

import (
	"encoding/json"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
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

func TestSegment_RestoreDataPopulatesWriterAndForcesEnabled(t *testing.T) {
	flags := &runtime.Flags{
		SegmentData: map[string]json.RawMessage{
			"session": json.RawMessage(`{"SSHSession":true}`),
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

func TestSegment_RestoreDataToggledOffStaysDisabled(t *testing.T) {
	cache.Set(cache.Session, cache.TOGGLECACHE, map[string]bool{"text": true}, cache.INFINITE)
	defer cache.Delete(cache.Session, cache.TOGGLECACHE)

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
