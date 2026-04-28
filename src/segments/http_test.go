package segments

import (
	"encoding/json"
	"errors"
	"io"
	"testing"

	runtimehttp "github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

// timeoutCapturingEnv wraps the mock environment to record the timeout argument
// passed to HTTPRequest, so tests can assert it flows through correctly.
type timeoutCapturingEnv struct {
	*mock.Environment
	capturedTimeout int
}

func (e *timeoutCapturingEnv) HTTPRequest(url string, body io.Reader, timeout int, modifiers ...runtimehttp.RequestModifier) ([]byte, error) {
	e.capturedTimeout = timeout
	return e.Environment.HTTPRequest(url, body, timeout, modifiers...)
}

func TestHTTPSegmentEnabled(t *testing.T) {
	cases := []struct {
		expected    any
		name        string
		url         string
		method      string
		response    string
		timeout     int
		shouldError bool
	}{
		{
			name:        "Valid URL with GET response",
			url:         "https://jsonplaceholder.typicode.com/posts/1",
			method:      "GET",
			timeout:     0,
			response:    `{"id": "1"}`,
			expected:    "1",
			shouldError: false,
		},
		{
			name:        "Valid URL with POST response",
			url:         "https://jsonplaceholder.typicode.com/posts",
			method:      "POST",
			timeout:     0,
			response:    `{"id": "101"}`,
			expected:    "101",
			shouldError: false,
		},
		{
			name:        "Valid URL with error response",
			url:         "https://api.example.com/data",
			method:      "GET",
			timeout:     0,
			shouldError: true,
		},
		{
			name:        "Empty URL",
			url:         "",
			method:      "GET",
			timeout:     0,
			shouldError: false,
		},
		{
			name:        "Custom timeout",
			url:         "https://jsonplaceholder.typicode.com/posts/1",
			method:      "GET",
			timeout:     5000,
			response:    `{"id": "2"}`,
			expected:    "2",
			shouldError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inner := new(mock.Environment)
			props := options.Map{
				URL:    tc.url,
				METHOD: tc.method,
			}

			if tc.timeout > 0 {
				props[options.HTTPTimeout] = tc.timeout
			}

			inner.On("HTTPRequest", tc.url).Return([]byte(tc.response), func() error {
				if tc.shouldError {
					return errors.New("error")
				}
				return nil
			}())

			capturing := &timeoutCapturingEnv{Environment: inner}

			cs := &HTTP{
				Base: Base{
					env:     capturing,
					options: props,
				},
			}

			_ = cs.Enabled()
			assert.Equal(t, tc.expected, cs.Body["id"], tc.name)
			if tc.timeout > 0 {
				assert.Equal(t, tc.timeout, capturing.capturedTimeout, "expected custom timeout to be passed to HTTPRequest")
			}
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
