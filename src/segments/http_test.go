package segments

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestHTTPSegmentEnabled(t *testing.T) {
	cases := []struct {
		expected    any
		name        string
		url         string
		method      string
		response    string
		shouldError bool
	}{
		{
			name:        "Valid URL with GET response",
			url:         "https://jsonplaceholder.typicode.com/posts/1",
			method:      "GET",
			response:    `{"id": "1"}`,
			expected:    "1",
			shouldError: false,
		},
		{
			name:        "Valid URL with POST response",
			url:         "https://jsonplaceholder.typicode.com/posts",
			method:      "POST",
			response:    `{"id": "101"}`,
			expected:    "101",
			shouldError: false,
		},
		{
			name:        "Valid URL with error response",
			url:         "https://api.example.com/data",
			method:      "GET",
			shouldError: true,
		},
		{
			name:        "Empty URL",
			url:         "",
			method:      "GET",
			shouldError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := new(mock.Environment)
			props := options.Map{
				URL:    tc.url,
				METHOD: tc.method,
			}

			env.On("HTTPRequest", tc.url).Return([]byte(tc.response), func() error {
				if tc.shouldError {
					return errors.New("error")
				}
				return nil
			}())

			cs := &HTTP{
				Base: Base{
					env:     env,
					options: props,
				},
			}

			_ = cs.Enabled()
			assert.Equal(t, tc.expected, cs.Body["id"], tc.name)
		})
	}
}

func TestHTTPSegmentCache(t *testing.T) {
	// Simulate what happens when caching
	response := `{"version": "39.2.6", "count": 42, "enabled": true}`

	// Create and populate HTTP segment
	original := &HTTP{
		Base: Base{
			Segment: &Segment{
				Text:  " Electron: v39.2.6 ",
				Index: 1,
			},
		},
	}

	var result map[string]any
	err := json.Unmarshal([]byte(response), &result)
	assert.NoError(t, err)
	original.Body = result

	// Marshal to JSON (like setCache does)
	data, err := json.Marshal(original)
	assert.NoError(t, err)

	// Unmarshal back (like restoreCache does)
	restored := &HTTP{
		Base: Base{
			Segment: &Segment{},
		},
	}

	err = json.Unmarshal(data, restored)
	assert.NoError(t, err)

	// Verify Body is restored correctly
	assert.NotNil(t, restored.Body, "Body should not be nil")
	assert.Equal(t, "39.2.6", restored.Body["version"], "version should be restored")
	assert.Equal(t, float64(42), restored.Body["count"], "count should be restored")
	assert.Equal(t, true, restored.Body["enabled"], "enabled should be restored")
}
