package engine

import (
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	Foo    = "foo"
	Bar    = "bar"
	FooBar = "foobar"
)

func TestHasProperty(t *testing.T) {
	cases := []struct {
		Case     string
		Expected bool
		Property properties.Property
		Props    properties.Map
	}{
		{Case: "Match", Expected: true, Property: Foo, Props: properties.Map{Foo: "bar"}},
		{Case: "No Match", Expected: false, Property: Foo, Props: properties.Map{Bar: "foo"}},
		{Case: "No properties", Expected: false},
	}
	for _, tc := range cases {
		segment := &Segment{
			Properties: tc.Props,
		}
		got := segment.hasProperty(tc.Property)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestMigratePropertyValue(t *testing.T) {
	cases := []struct {
		Case     string
		Expected interface{}
		Property properties.Property
		Props    properties.Map
	}{
		{Case: "Match", Expected: "foo", Property: Foo, Props: properties.Map{Foo: "bar"}},
		{Case: "No Match", Expected: nil, Property: Foo, Props: properties.Map{Bar: "foo"}},
	}
	for _, tc := range cases {
		segment := &Segment{
			Properties: tc.Props,
		}
		segment.migratePropertyValue(tc.Property, tc.Expected)
		assert.Equal(t, tc.Expected, segment.Properties[tc.Property], tc.Case)
	}
}

func TestMigratePropertyKey(t *testing.T) {
	cases := []struct {
		Case        string
		Expected    interface{}
		OldProperty properties.Property
		NewProperty properties.Property
		Props       properties.Map
	}{
		{Case: "Match", Expected: "bar", OldProperty: Foo, NewProperty: Bar, Props: properties.Map{Foo: "bar"}},
		{Case: "No match", Expected: nil, OldProperty: Foo, NewProperty: Bar, Props: properties.Map{FooBar: "bar"}},
		{Case: "No migration", Expected: "bar", OldProperty: Foo, NewProperty: Bar, Props: properties.Map{Bar: "bar"}},
	}
	for _, tc := range cases {
		segment := &Segment{
			Properties: tc.Props,
		}
		segment.migratePropertyKey(tc.OldProperty, tc.NewProperty)
		assert.Equal(t, tc.Expected, segment.Properties[tc.NewProperty], tc.Case)
		assert.NotContains(t, segment.Properties, tc.OldProperty, tc.Case)
	}
}

type MockedWriter struct {
	template string
}

func (m *MockedWriter) Enabled() bool {
	return true
}

func (m *MockedWriter) Template() string {
	return m.template
}

func (m *MockedWriter) Init(props properties.Properties, env environment.Environment) {}

func TestIconOverride(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Property properties.Property
		Props    properties.Map
	}{
		{
			Case:     "Match",
			Expected: "hello bar bar",
			Property: Foo,
			Props: properties.Map{
				Foo:                        " bar ",
				properties.SegmentTemplate: "hello foo bar",
			},
		},
		{
			Case:     "No match",
			Expected: "hello foo bar",
			Property: Foo,
			Props: properties.Map{
				Bar:                        " bar ",
				properties.SegmentTemplate: "hello foo bar",
			},
		},
	}
	for _, tc := range cases {
		segment := &Segment{
			Properties: tc.Props,
			writer: &MockedWriter{
				template: tc.Props.GetString(properties.SegmentTemplate, ""),
			},
		}
		segment.migrateIconOverride(tc.Property, " foo ")
		assert.Equal(t, tc.Expected, segment.Properties[properties.SegmentTemplate], tc.Case)
	}
}

func TestColorMigration(t *testing.T) {
	cases := []struct {
		Case       string
		Expected   string
		Property   properties.Property
		Template   string
		Background bool
		NoOverride bool
		Props      properties.Map
	}{
		{
			Case:     "Foreground override",
			Expected: "hello green bar",
			Template: "hello %s bar",
			Property: Foo,
			Props: properties.Map{
				Foo: "green",
			},
		},
		{
			Case:       "Background override",
			Expected:   "hello green bar",
			Template:   "hello %s bar",
			Property:   Foo,
			Background: true,
			Props: properties.Map{
				Foo: "green",
			},
		},
		{
			Case:       "No override",
			Expected:   "hello green bar",
			Template:   "hello %s bar",
			Property:   Foo,
			NoOverride: true,
		},
	}
	for _, tc := range cases {
		segment := &Segment{
			Properties: tc.Props,
		}
		if tc.Background {
			segment.Properties[colorBackground] = true
		}
		segment.migrateColorOverride(tc.Property, tc.Template)
		templates := segment.ForegroundTemplates
		if tc.Background {
			templates = segment.BackgroundTemplates
		}
		if tc.NoOverride {
			assert.Empty(t, templates, tc.Case)
			return
		}
		lastElement := templates[len(templates)-1]
		assert.Equal(t, tc.Expected, lastElement, tc.Case)
	}
}

func TestSegmentTemplateMigration(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Type     SegmentType
		Props    properties.Map
	}{
		{
			Case:     "GIT",
			Expected: " {{ .HEAD }} {{ .BranchStatus }}{{ if .Working.Changed }} working {{ .Working.String }}{{ end }}{{ if and (.Staging.Changed) (.Working.Changed) }} and{{ end }}{{ if .Staging.Changed }} staged {{ .Staging.String }}{{ end }}{{ if gt .StashCount 0}} stash {{ .StashCount }}{{ end }}{{ if gt .WorktreeCount 0}} worktree {{ .WorktreeCount }}{{ end }} ", // nolint: lll
			Type:     GIT,
			Props: properties.Map{
				"local_working_icon":    " working ",
				"local_staged_icon":     " staged ",
				"worktree_count_icon":   " worktree ",
				"stash_count_icon":      " stash ",
				"status_separator_icon": " and",
			},
		},
		{
			Case:     "GIT - Staging and Working Color",
			Expected: " {{ .HEAD }} {{ .BranchStatus }}{{ if .Working.Changed }} working <#123456>{{ .Working.String }}</>{{ end }}{{ if and (.Staging.Changed) (.Working.Changed) }} and{{ end }}{{ if .Staging.Changed }} staged <#123456>{{ .Staging.String }}</>{{ end }}{{ if gt .StashCount 0}} stash {{ .StashCount }}{{ end }}{{ if gt .WorktreeCount 0}} worktree {{ .WorktreeCount }}{{ end }} ", // nolint: lll
			Type:     GIT,
			Props: properties.Map{
				"local_working_icon":    " working ",
				"local_staged_icon":     " staged ",
				"worktree_count_icon":   " worktree ",
				"stash_count_icon":      " stash ",
				"status_separator_icon": " and",
				"working_color":         "#123456",
				"staging_color":         "#123456",
			},
		},
		{
			Case:     "EXIT - No exit Code with Icon overrides",
			Expected: " {{ if gt .Code 0 }}FAIL{{ else }}SUCCESS{{ end }} ",
			Type:     EXIT,
			Props: properties.Map{
				"display_exit_code": false,
				"success_icon":      "SUCCESS",
				"error_icon":        "FAIL",
			},
		},
		{
			Case:     "EXIT - Always numeric",
			Expected: " {{ if gt .Code 0 }}FAIL {{ .Code }}{{ else }}SUCCESS{{ end }} ",
			Type:     EXIT,
			Props: properties.Map{
				"always_numeric": true,
				"success_icon":   "SUCCESS",
				"error_icon":     "FAIL",
			},
		},
		{
			Case:     "BATTERY",
			Expected: ` {{ if not .Error }}{{ $stateList := list "Discharging" "Full" }}{{ if has .State.String $stateList }}{{ .Icon }}{{ .Percentage }}{{ end }}{{ end }}{{ .Error }} `,
			Type:     BATTERY,
			Props: properties.Map{
				"display_charging": false,
			},
		},
		{
			Case:     "SESSION",
			Expected: " {{ if .SSHSession }}SSH {{ end }}{{ .UserName }}@{{ .HostName }} ",
			Type:     SESSION,
			Props: properties.Map{
				"ssh_icon": "SSH ",
			},
		},
		{
			Case:     "SESSION no HOST",
			Expected: " {{ if .SSHSession }}\uf817 {{ end }}{{ .UserName }} ",
			Type:     SESSION,
			Props: properties.Map{
				"display_host": false,
			},
		},
		{
			Case:     "SESSION no USER",
			Expected: " {{ if .SSHSession }}\uf817 {{ end }}{{ .HostName }} ",
			Type:     SESSION,
			Props: properties.Map{
				"display_user": false,
			},
		},
		{
			Case:     "SESSION no USER nor HOST",
			Expected: " {{ if .SSHSession }}\uf817 {{ end }} ",
			Type:     SESSION,
			Props: properties.Map{
				"display_user": false,
				"display_host": false,
			},
		},
		{
			Case:     "SESSION - Color overrides",
			Expected: " {{ if .SSHSession }}\uf817 {{ end }}<#123456>{{ .UserName }}</>@<#789012>{{ .HostName }}</> ",
			Type:     SESSION,
			Props: properties.Map{
				"user_color": "#123456",
				"host_color": "#789012",
			},
		},
	}
	for _, tc := range cases {
		segment := &Segment{
			Type:       tc.Type,
			Properties: tc.Props,
		}
		segment.migrationOne(&mock.MockedEnvironment{})
		assert.Equal(t, tc.Expected, segment.Properties[properties.SegmentTemplate], tc.Case)
	}
}

func TestInlineColorOverride(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Property properties.Property
		Props    properties.Map
	}{
		{
			Case:     "Match",
			Expected: "hello <#123456>foo</> bar",
			Property: Foo,
			Props: properties.Map{
				Foo:                        "#123456",
				properties.SegmentTemplate: "hello foo bar",
			},
		},
		{
			Case:     "No match",
			Expected: "hello foo bar",
			Property: Foo,
			Props: properties.Map{
				Bar:                        "#123456",
				properties.SegmentTemplate: "hello foo bar",
			},
		},
	}
	for _, tc := range cases {
		segment := &Segment{
			Properties: tc.Props,
			writer: &MockedWriter{
				template: tc.Props.GetString(properties.SegmentTemplate, ""),
			},
		}
		segment.migrateInlineColorOverride(tc.Property, "foo")
		assert.Equal(t, tc.Expected, segment.Properties[properties.SegmentTemplate], tc.Case)
	}
}

func TestMigratePreAndPostfix(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Props    properties.Map
	}{
		{
			Case:     "Pre and Postfix",
			Expected: "<background,transparent>\ue0b6</> \uf489 {{ .Name }} <transparent,background>\ue0b2</>",
			Props: properties.Map{
				"postfix":  " <transparent,background>\ue0b2</>",
				"prefix":   "<background,transparent>\ue0b6</> \uf489 ",
				"template": "{{ .Name }}",
			},
		},
		{
			Case:     "Prefix",
			Expected: " {{ .Name }} ",
			Props: properties.Map{
				"prefix":   " ",
				"template": "{{ .Name }}",
			},
		},
		{
			Case:     "Postfix",
			Expected: " {{ .Name }} ",
			Props: properties.Map{
				"postfix":  " ",
				"template": "{{ .Name }} ",
			},
		},
	}
	for _, tc := range cases {
		segment := &Segment{
			Properties: tc.Props,
			writer: &MockedWriter{
				template: tc.Props.GetString(properties.SegmentTemplate, ""),
			},
		}
		segment.migrateTemplate()
		assert.Equal(t, tc.Expected, segment.Properties[properties.SegmentTemplate], tc.Case)
		assert.NotContains(t, segment.Properties, "prefix", tc.Case)
		assert.NotContains(t, segment.Properties, "postfix", tc.Case)
	}
}
