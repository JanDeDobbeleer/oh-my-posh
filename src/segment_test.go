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
	assert.NotNil(t, sc.props)
	assert.NoError(t, err)
	assert.NotNil(t, sc.writer)
}

func TestMapSegmentWriterCannotMap(t *testing.T) {
	sc := &Segment{
		Type: "nilwriter",
	}
	env := new(MockedEnvironment)
	err := sc.mapSegmentWithWriter(env)
	assert.Nil(t, sc.props)
	assert.Error(t, err)
}

func TestParseTestSettings(t *testing.T) {
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
				"ignore_folders": [
					"/super/secret/project"
				]
			}
		}
		`
	segment := &Segment{}
	err := json.Unmarshal([]byte(segmentJSON), segment)
	assert.NoError(t, err)
	cwd := "/super/secret/project"
	got := segment.shouldIgnoreFolder(cwd)
	assert.True(t, got)
}

func TestShouldIgnoreFolderRegex(t *testing.T) {
	segment := &Segment{
		Properties: map[Property]interface{}{
			IgnoreFolders: []string{"Projects[\\/].*"},
		},
	}
	got := segment.shouldIgnoreFolder(cwd)
	assert.True(t, got)
}

func TestShouldIgnoreFolderRegexNonEscaped(t *testing.T) {
	segment := &Segment{
		Properties: map[Property]interface{}{
			IgnoreFolders: []string{"Projects/.*"},
		},
	}
	got := segment.shouldIgnoreFolder(cwd)
	assert.True(t, got)
}

func TestShouldIgnoreFolderRegexInverted(t *testing.T) {
	segment := &Segment{
		Properties: map[Property]interface{}{
			IgnoreFolders: []string{"(?!Projects[\\/]).*"},
		},
	}
	// detect panic(thrown by MustCompile)
	defer func() {
		if err := recover(); err != nil {
			// display a message explaining omp failed(with the err)
			assert.Equal(t, "regexp: Compile(`^(?!Projects[\\/]).*$`): error parsing regexp: invalid or unsupported Perl syntax: `(?!`", err)
		}
	}()
	segment.shouldIgnoreFolder(cwd)
}

func TestShouldIgnoreFolderRegexInvertedNonEscaped(t *testing.T) {
	segment := &Segment{
		Properties: map[Property]interface{}{
			IgnoreFolders: []string{"(?!Projects/).*"},
		},
	}
	// detect panic(thrown by MustCompile)
	defer func() {
		if err := recover(); err != nil {
			// display a message explaining omp failed(with the err)
			assert.Equal(t, "regexp: Compile(`^(?!Projects/).*$`): error parsing regexp: invalid or unsupported Perl syntax: `(?!`", err)
		}
	}()
	segment.shouldIgnoreFolder(cwd)
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
		segment := &Segment{
			writer: &aws{
				Profile: tc.Profile,
				Region:  tc.Region,
			},
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
