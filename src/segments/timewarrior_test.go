package segments

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimewarrior(t *testing.T) {
	start := time.Now().UTC().Add(-90 * time.Minute).Truncate(time.Second)
	activeOutput := fmt.Sprintf(`{"id":1,"start":%q,"tags":["work","project"],"annotation":"Writing code"}`, start.Format("20060102T150405Z"))
	const stoppedOutput = `{"id":1,"start":"20261012T090000Z","end":"20261012T103000Z","tags":["work"]}`
	statusError := errors.New("status failed")
	lookupError := errors.New("active interval lookup failed")
	cases := []struct {
		Error           error
		StatusError     error
		Case            string
		Command         string
		Status          string
		Output          string
		Annotation      string
		ExpectedTags    string
		HasCommand      bool
		AlwaysEnabled   bool
		ExpectedEnabled bool
		Active          bool
	}{
		{Case: "missing executable"},
		{Case: "invalid status", HasCommand: true, Status: "unexpected"},
		{Case: "inactive", HasCommand: true, Status: "0"},
		{Case: "inactive indicator", HasCommand: true, Status: "0", AlwaysEnabled: true, ExpectedEnabled: true},
		{Case: "status command error", HasCommand: true, Status: "1", StatusError: statusError},
		{Case: "status error with idle indicator", HasCommand: true, Status: "0", StatusError: statusError, AlwaysEnabled: true},
		{Case: "lookup error", HasCommand: true, Error: lookupError},
		{Case: "lookup error with idle indicator", HasCommand: true, Error: lookupError, AlwaysEnabled: true},
		{Case: "invalid JSON", HasCommand: true, Output: "invalid"},
		{Case: "invalid start", HasCommand: true, Output: `{"start":"invalid"}`},
		{Case: "missing start", HasCommand: true, Output: `{}`},
		{Case: "invalid end", HasCommand: true, Output: `{"start":"20261012T090000Z","end":"invalid"}`},
		{Case: "unexpected array", HasCommand: true, Output: "[]"},
		{Case: "null interval with idle indicator", HasCommand: true, Output: "null", AlwaysEnabled: true},
		{Case: "stopped interval hidden", HasCommand: true, Output: stoppedOutput},
		{
			Case: "stopped interval discarded for idle indicator", HasCommand: true, Output: stoppedOutput,
			AlwaysEnabled: true, ExpectedEnabled: true,
		},
		{
			Case: "stopped interval with invalid timestamps discarded", HasCommand: true,
			Output:        `{"id":1,"start":"invalid","end":"invalid","tags":["work"],"annotation":"Finished"}`,
			AlwaysEnabled: true, ExpectedEnabled: true,
		},
		{
			Case: "active interval", HasCommand: true, Output: activeOutput,
			ExpectedEnabled: true, Active: true, ExpectedTags: "#work #project", Annotation: "Writing code",
		},
		{
			Case: "custom executable", Command: "timew-custom", HasCommand: true, Output: activeOutput,
			ExpectedEnabled: true, Active: true, ExpectedTags: "#work #project", Annotation: "Writing code",
		},
		{
			Case: "active without tags", HasCommand: true,
			Output: fmt.Sprintf(`{"id":1,"start":%q}`, start.Format("20060102T150405Z")), ExpectedEnabled: true, Active: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			env := new(mock.Environment)
			command := tc.Command
			if command == "" {
				command = "timew"
			}
			env.On("HasCommand", command).Return(tc.HasCommand).Once()
			if tc.HasCommand {
				status := tc.Status
				if status == "" {
					status = "1"
				}
				env.On("RunCommand", command, []string{"get", "dom.active"}).Return(status, tc.StatusError).Once()
				if status == "1" && tc.StatusError == nil {
					env.On("RunCommand", command, []string{"get", "dom.active.json"}).Return(tc.Output, tc.Error).Once()
				}
			}
			props := options.Map{
				TimewarriorCommand:    command,
				options.AlwaysEnabled: tc.AlwaysEnabled,
			}
			segment := &Timewarrior{}
			segment.Init(props, env)
			require.Equal(t, tc.ExpectedEnabled, segment.Enabled())
			env.AssertExpectations(t)
			if !tc.ExpectedEnabled {
				return
			}
			assert.Equal(t, tc.Active, segment.Active)
			assert.Equal(t, tc.ExpectedTags, segment.Tags)
			assert.Equal(t, tc.Annotation, segment.Annotation)
			if !tc.Active {
				assert.True(t, segment.Start.IsZero())
				assert.Empty(t, segment.Duration)
				return
			}
			assert.Equal(t, start, segment.Start)
			duration, err := time.ParseDuration(segment.Duration)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, duration, 90*time.Minute)
			assert.LessOrEqual(t, duration, time.Since(start))
		})
	}
}
func TestTimewarriorTemplate(t *testing.T) {
	cases := []struct {
		Options            options.Map
		Case               string
		Annotation         string
		ExpectedAnnotation string
		ExpectedTags       string
		Rendered           string
		Tags               []string
		Active             bool
	}{
		{Case: "idle"},
		{Case: "idle indicator", Options: options.Map{options.AlwaysEnabled: true}, Rendered: "Not tracking"},
		{
			Case: "active with always enabled", Active: true, Tags: []string{"work"},
			Options: options.Map{options.AlwaysEnabled: true}, ExpectedTags: "#work", Rendered: "\uf252 #work",
		},
		{Case: "tags", Active: true, Tags: []string{"work", "project"}, ExpectedTags: "#work #project", Rendered: "\uf252 #work #project"},
		{
			Case: "tags and annotation", Active: true, Tags: []string{"work"}, Annotation: "Writing code",
			ExpectedAnnotation: "Writing code", ExpectedTags: "#work", Rendered: "\uf252 #work | Writing code",
		},
		{Case: "untagged", Active: true, Rendered: "\uf252 \u2205"},
		{
			Case: "untagged annotation", Active: true, Annotation: "Writing code",
			ExpectedAnnotation: "Writing code", Rendered: "\uf252 \u2205 | Writing code",
		},
		{
			Case: "real tag beginning with marker", Active: true, Tags: []string{"\u2205work"},
			ExpectedTags: "#\u2205work", Rendered: "\uf252 #\u2205work",
		},
		{
			Case: "long tags stay intact", Active: true, Tags: []string{strings.Repeat("x", 60)},
			ExpectedTags: "#" + strings.Repeat("x", 60), Rendered: "\uf252 #" + strings.Repeat("x", 60),
		},
		{
			Case: "exact annotation limit", Active: true, Tags: []string{"work"}, Annotation: strings.Repeat("x", 48),
			ExpectedAnnotation: strings.Repeat("x", 48), ExpectedTags: "#work", Rendered: "\uf252 #work | " + strings.Repeat("x", 48),
		},
		{
			Case: "unicode annotation truncation", Active: true, Tags: []string{"work"}, Annotation: strings.Repeat("\u754c", 49),
			ExpectedAnnotation: strings.Repeat("\u754c", 47) + "\u2026", ExpectedTags: "#work",
			Rendered: "\uf252 #work | " + strings.Repeat("\u754c", 47) + "\u2026",
		},
		{
			Case: "custom tag format and separator", Active: true, Tags: []string{"work", "project"}, Annotation: "Writing",
			Options:            options.Map{TimewarriorTagFormat: "[%s]", TimewarriorTagSeparator: ","},
			ExpectedAnnotation: "Writing", ExpectedTags: "[work],[project]", Rendered: "\uf252 [work],[project] | Writing",
		},
		{
			Case: "custom truncation", Active: true, Tags: []string{"work"}, Annotation: "abcdefgh",
			Options:            options.Map{TimewarriorAnnotationMaxLength: 6, TimewarriorEllipsis: ".."},
			ExpectedAnnotation: "abcd..", ExpectedTags: "#work", Rendered: "\uf252 #work | abcd..",
		},
		{
			Case: "ellipsis longer than limit", Active: true, Annotation: "abcdefgh",
			Options:            options.Map{TimewarriorAnnotationMaxLength: 2, TimewarriorEllipsis: "..."},
			ExpectedAnnotation: "..", Rendered: "\uf252 \u2205 | ..",
		},
		{
			Case: "unlimited", Active: true, Annotation: strings.Repeat("x", 60), Options: options.Map{TimewarriorAnnotationMaxLength: 0},
			ExpectedAnnotation: strings.Repeat("x", 60), Rendered: "\uf252 \u2205 | " + strings.Repeat("x", 60),
		},
		{
			Case: "empty ellipsis", Active: true, Annotation: "abcdefgh",
			Options:            options.Map{TimewarriorAnnotationMaxLength: 3, TimewarriorEllipsis: ""},
			ExpectedAnnotation: "abc", Rendered: "\uf252 \u2205 | abc",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			env := new(mock.Environment)
			env.On("HasCommand", "timew").Return(true).Once()
			status := "0"
			if tc.Active {
				status = "1"
				output, err := json.Marshal(map[string]any{
					"id": 1, "start": time.Now().UTC().Format("20060102T150405Z"), "tags": tc.Tags, "annotation": tc.Annotation,
				})
				require.NoError(t, err)
				env.On("RunCommand", "timew", []string{"get", "dom.active.json"}).Return(string(output), nil).Once()
			}
			env.On("RunCommand", "timew", []string{"get", "dom.active"}).Return(status, nil).Once()
			segment := &Timewarrior{}
			segment.Init(tc.Options, env)
			enabled := segment.Enabled()
			assert.Equal(t, tc.Active || tc.Options.Bool(options.AlwaysEnabled, false), enabled)
			assert.Equal(t, tc.ExpectedAnnotation, segment.Annotation)
			assert.Equal(t, tc.ExpectedTags, segment.Tags)
			rendered := ""
			if enabled {
				rendered = renderTemplate(env, segment.Template(), segment)
			}
			assert.Equal(t, tc.Rendered, rendered)
			env.AssertExpectations(t)
		})
	}
}
