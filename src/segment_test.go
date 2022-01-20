package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	cwd = "Projects/oh-my-posh"
)

func TestMapSegmentWriterCanMap(t *testing.T) {
	sc := &Segment{
		Type: Session,
	}
	env := new(MockedEnvironment)
	err := sc.mapSegmentWithWriter(env)
	assert.NoError(t, err)
	assert.NotNil(t, sc.writer)
}

func TestMapSegmentWriterCannotMap(t *testing.T) {
	sc := &Segment{
		Type: "nilwriter",
	}
	env := new(MockedEnvironment)
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
				"prefix": " \uE5FF ",
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
		Case           string
		IncludeFolders []string
		ExcludeFolders []string
		Expected       bool
	}{
		{Case: "Base Case", IncludeFolders: nil, ExcludeFolders: nil, Expected: true},
		{Case: "Base Case Empty Arrays", IncludeFolders: []string{}, ExcludeFolders: []string{}, Expected: true},

		{Case: "Include", IncludeFolders: []string{"Projects/oh-my-posh"}, ExcludeFolders: nil, Expected: true},
		{Case: "Include Regex", IncludeFolders: []string{"Projects.*"}, ExcludeFolders: nil, Expected: true},
		{Case: "Include Mismatch", IncludeFolders: []string{"Projects/nope"}, ExcludeFolders: nil, Expected: false},
		{Case: "Include Regex Mismatch", IncludeFolders: []string{"zProjects.*"}, ExcludeFolders: nil, Expected: false},

		{Case: "Exclude", IncludeFolders: nil, ExcludeFolders: []string{"Projects/oh-my-posh"}, Expected: false},
		{Case: "Exclude Regex", IncludeFolders: nil, ExcludeFolders: []string{"Projects.*"}, Expected: false},
		{Case: "Exclude Mismatch", IncludeFolders: nil, ExcludeFolders: []string{"Projects/nope"}, Expected: true},
		{Case: "Exclude Regex Mismatch", IncludeFolders: nil, ExcludeFolders: []string{"zProjects.*"}, Expected: true},

		{Case: "Include Match / Exclude Match", IncludeFolders: []string{"Projects.*"}, ExcludeFolders: []string{"Projects/oh-my-posh"}, Expected: false},
		{Case: "Include Match / Exclude Mismatch", IncludeFolders: []string{"Projects.*"}, ExcludeFolders: []string{"Projects/nope"}, Expected: true},
		{Case: "Include Mismatch / Exclude Match", IncludeFolders: []string{"zProjects.*"}, ExcludeFolders: []string{"Projects/oh-my-posh"}, Expected: false},
		{Case: "Include Mismatch / Exclude Mismatch", IncludeFolders: []string{"zProjects.*"}, ExcludeFolders: []string{"Projects/nope"}, Expected: false},
	}
	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getRuntimeGOOS").Return(linuxPlatform)
		env.On("homeDir").Return("")
		env.On("pwd").Return(cwd)
		segment := &Segment{
			Properties: properties{
				IncludeFolders: tc.IncludeFolders,
				ExcludeFolders: tc.ExcludeFolders,
			},
			env: env,
		}
		got := segment.shouldIncludeFolder()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestShouldIncludeFolderRegexInverted(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("getRuntimeGOOS").Return(linuxPlatform)
	env.On("homeDir").Return("")
	env.On("pwd").Return(cwd)
	segment := &Segment{
		Properties: properties{
			ExcludeFolders: []string{"(?!Projects[\\/]).*"},
		},
		env: env,
	}
	// detect panic(thrown by MustCompile)
	defer func() {
		if err := recover(); err != nil {
			// display a message explaining omp failed(with the err)
			assert.Equal(t, "regexp: Compile(`^(?!Projects[\\/]).*$`): error parsing regexp: invalid or unsupported Perl syntax: `(?!`", err)
		}
	}()
	segment.shouldIncludeFolder()
}

func TestShouldIncludeFolderRegexInvertedNonEscaped(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("getRuntimeGOOS").Return(linuxPlatform)
	env.On("homeDir").Return("")
	env.On("pwd").Return(cwd)
	segment := &Segment{
		Properties: properties{
			ExcludeFolders: []string{"(?!Projects/).*"},
		},
		env: env,
	}
	// detect panic(thrown by MustCompile)
	defer func() {
		if err := recover(); err != nil {
			// display a message explaining omp failed(with the err)
			assert.Equal(t, "regexp: Compile(`^(?!Projects/).*$`): error parsing regexp: invalid or unsupported Perl syntax: `(?!`", err)
		}
	}()
	segment.shouldIncludeFolder()
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
		env := new(MockedEnvironment)
		env.onTemplate()
		segment := &Segment{
			writer: &aws{
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
