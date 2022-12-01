package engine

import (
	"encoding/json"
	"oh-my-posh/mock"
	"oh-my-posh/platform"
	"oh-my-posh/properties"
	"oh-my-posh/segments"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	cwd = "Projects/oh-my-posh"
)

func TestMapSegmentWriterCanMap(t *testing.T) {
	sc := &Segment{
		Type: SESSION,
	}
	env := new(mock.MockedEnvironment)
	err := sc.mapSegmentWithWriter(env)
	assert.NoError(t, err)
	assert.NotNil(t, sc.writer)
}

func TestMapSegmentWriterCannotMap(t *testing.T) {
	sc := &Segment{
		Type: "nilwriter",
	}
	env := new(mock.MockedEnvironment)
	err := sc.mapSegmentWithWriter(env)
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
			"properties": {
				"style": "folder",
				"exclude_folders": [
					"/super/secret/project"
				]
			}
		}
		`
	segment := &Segment{}
	err := json.Unmarshal([]byte(segmentJSON), segment)
	assert.NoError(t, err)
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
		env := new(mock.MockedEnvironment)
		env.On("GOOS").Return(platform.LINUX)
		env.On("Home").Return("")
		env.On("Pwd").Return(cwd)
		env.On("DirMatchesOneOf", cwd, []string{"Projects/oh-my-posh"}).Return(tc.Included)
		env.On("DirMatchesOneOf", cwd, []string{"Projects/nope"}).Return(tc.Excluded)
		segment := &Segment{
			Properties: properties.Map{
				properties.IncludeFolders: []string{"Projects/oh-my-posh"},
				properties.ExcludeFolders: []string{"Projects/nope"},
			},
			env: env,
		}
		got := segment.shouldIncludeFolder()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestGetColors(t *testing.T) {
	cases := []struct {
		Case          string
		Background    bool
		ExpectedColor string
		Templates     []string
		DefaultColor  string
		Region        string
		Profile       string
	}{
		{Case: "No template - foreground", ExpectedColor: "color", Background: false, DefaultColor: "color"},
		{Case: "No template - background", ExpectedColor: "color", Background: true, DefaultColor: "color"},
		{Case: "Nil template", ExpectedColor: "color", DefaultColor: "color", Templates: nil},
		{
			Case:          "Template - default",
			ExpectedColor: "color",
			DefaultColor:  "color",
			Templates: []string{
				"{{if contains \"john\" .Profile}}color2{{end}}",
			},
			Profile: "doe",
		},
		{
			Case:          "Template - override",
			ExpectedColor: "color2",
			DefaultColor:  "color",
			Templates: []string{
				"{{if contains \"john\" .Profile}}color2{{end}}",
			},
			Profile: "john",
		},
		{
			Case:          "Template - override multiple",
			ExpectedColor: "color3",
			DefaultColor:  "color",
			Templates: []string{
				"{{if contains \"doe\" .Profile}}color2{{end}}",
				"{{if contains \"john\" .Profile}}color3{{end}}",
			},
			Profile: "john",
		},
		{
			Case:          "Template - override multiple no match",
			ExpectedColor: "color",
			DefaultColor:  "color",
			Templates: []string{
				"{{if contains \"doe\" .Profile}}color2{{end}}",
				"{{if contains \"philip\" .Profile}}color3{{end}}",
			},
			Profile: "john",
		},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("TemplateCache").Return(&platform.TemplateCache{
			Env: make(map[string]string),
		})
		segment := &Segment{
			writer: &segments.Aws{
				Profile: tc.Profile,
				Region:  tc.Region,
			},
			env: env,
		}
		if tc.Background {
			segment.Background = tc.DefaultColor
			segment.BackgroundTemplates = tc.Templates
			color := segment.background()
			assert.Equal(t, tc.ExpectedColor, color, tc.Case)
			continue
		}
		segment.Foreground = tc.DefaultColor
		segment.ForegroundTemplates = tc.Templates
		color := segment.foreground()
		assert.Equal(t, tc.ExpectedColor, color, tc.Case)
	}
}

func TestShouldHideForCols(t *testing.T) {
	cases := []struct {
		Case     string
		MinWidth int
		MaxWidth int
		Width    int
		Error    error
		Expected bool
	}{
		{Case: "No settings"},
		{Case: "Min cols - hide", MinWidth: 10, Width: 9, Expected: true},
		{Case: "Min cols - show", MinWidth: 10, Width: 20, Expected: false},
		{Case: "Max cols - hide", MaxWidth: 10, Width: 11, Expected: true},
		{Case: "Max cols - show", MaxWidth: 10, Width: 8, Expected: false},
		{Case: "Min & Max cols - hide", MinWidth: 10, MaxWidth: 20, Width: 21, Expected: true},
		{Case: "Min & Max cols - hide 2", MinWidth: 10, MaxWidth: 20, Width: 8, Expected: true},
		{Case: "Min & Max cols - show", MinWidth: 10, MaxWidth: 20, Width: 11, Expected: false},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("TerminalWidth").Return(tc.Width, tc.Error)
		segment := &Segment{
			env:      env,
			MaxWidth: tc.MaxWidth,
			MinWidth: tc.MinWidth,
		}
		got := segment.shouldHideForWidth()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
