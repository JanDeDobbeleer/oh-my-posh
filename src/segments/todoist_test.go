package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

const TodoistTestURL = "https://api.todoist.com/api/v1/tasks/filter?query=due today"

func TestTodoistSegment(t *testing.T) {
	cases := []struct {
		Error           error
		Case            string
		JSONResponse    string
		ExpectedCount   int
		ExpectedEnabled bool
	}{
		{
			Case:            "No tasks",
			JSONResponse:    `{"results": []}`,
			ExpectedCount:   0,
			ExpectedEnabled: true,
		},
		{
			Case:            "Single task",
			JSONResponse:    `{"results": [{"id": "123"}]}`,
			ExpectedCount:   1,
			ExpectedEnabled: true,
		},
		{
			Case:            "Multiple tasks",
			JSONResponse:    `{"results": [{"id": "1"}, {"id": "2"}, {"id": "3"}]}`,
			ExpectedCount:   3,
			ExpectedEnabled: true,
		},
		{
			Case:            "API error",
			JSONResponse:    ``,
			ExpectedCount:   0,
			ExpectedEnabled: false,
			Error:           errors.New("API request failed"),
		},
		{
			Case:            "Invalid JSON response",
			JSONResponse:    `invalid json`,
			ExpectedCount:   0,
			ExpectedEnabled: false,
		},
		{
			Case:            "Task with additional fields",
			JSONResponse:    `{"results": [{"id": "456", "content": "Buy milk", "due": {"date": "2024-01-15"}}]}`,
			ExpectedCount:   1,
			ExpectedEnabled: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			env := new(mock.Environment)
			props := options.Map{
				APIKey: "fake-api-key",
			}

			env.On("HTTPRequest", TodoistTestURL).Return([]byte(tc.JSONResponse), tc.Error)

			todoist := &Todoist{}
			todoist.Init(props, env)

			enabled := todoist.Enabled()
			assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)

			if enabled {
				assert.Equal(t, tc.ExpectedCount, todoist.TaskCount, tc.Case)
			}
		})
	}
}

func TestTodoistTemplate(t *testing.T) {
	todoist := &Todoist{}
	assert.Equal(t, "{{ .TaskCount }} ", todoist.Template())
}

func TestTodoistTemplateRendering(t *testing.T) {
	cases := []struct {
		Case           string
		JSONResponse   string
		Template       string
		ExpectedString string
	}{
		{
			Case:           "Default template with no tasks",
			JSONResponse:   `{"results": []}`,
			Template:       "{{ .TaskCount }}",
			ExpectedString: "0",
		},
		{
			Case:           "Default template with tasks",
			JSONResponse:   `{"results": [{"id": "1"}, {"id": "2"}]}`,
			Template:       "{{ .TaskCount }}",
			ExpectedString: "2",
		},
		{
			Case:           "Custom template with icon",
			JSONResponse:   `{"results": [{"id": "1"}, {"id": "2"}, {"id": "3"}]}`,
			Template:       "📋 {{ .TaskCount }} tasks",
			ExpectedString: "📋 3 tasks",
		},
		{
			Case:           "Conditional template - has tasks",
			JSONResponse:   `{"results": [{"id": "1"}]}`,
			Template:       "{{ if gt .TaskCount 0 }}📋 {{ .TaskCount }}{{ end }}",
			ExpectedString: "📋 1",
		},
		{
			Case:           "Conditional template - no tasks",
			JSONResponse:   `{"results": []}`,
			Template:       "{{ if gt .TaskCount 0 }}📋 {{ .TaskCount }}{{ else }}✅{{ end }}",
			ExpectedString: "✅",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			env := new(mock.Environment)
			props := options.Map{
				APIKey: "fake-api-key",
			}

			env.On("HTTPRequest", TodoistTestURL).Return([]byte(tc.JSONResponse), nil)

			todoist := &Todoist{}
			todoist.Init(props, env)

			enabled := todoist.Enabled()
			assert.True(t, enabled, tc.Case)

			result := renderTemplate(env, tc.Template, todoist)
			assert.Equal(t, tc.ExpectedString, result, tc.Case)
		})
	}
}
