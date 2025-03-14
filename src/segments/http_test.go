package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

func TestHTTPSegmentEnabled(t *testing.T) {
	cases := []struct {
		caseName   string
		url        string
		method     string
		response   string
		responseID string
		isError    bool
		expected   bool
	}{
		{
			caseName:   "Valid URL with GET response",
			url:        "https://jsonplaceholder.typicode.com/posts/1",
			method:     "GET",
			response:   `{"id": "1"}`,
			responseID: "1",
			isError:    false,
			expected:   true,
		},
		{
			caseName:   "Valid URL with POST response",
			url:        "https://jsonplaceholder.typicode.com/posts",
			method:     "POST",
			response:   `{"id": "101"}`,
			responseID: "101",
			isError:    false,
			expected:   true,
		},
		{
			caseName:   "Valid URL with error response",
			url:        "https://api.example.com/data",
			method:     "GET",
			response:   ``,
			responseID: ``,
			isError:    true,
			expected:   false,
		},
		{
			caseName:   "Empty URL",
			url:        "",
			method:     "GET",
			response:   ``,
			responseID: ``,
			isError:    false,
			expected:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			env := new(mock.Environment)
			props := properties.Map{
				URL:    tc.url,
				METHOD: tc.method,
			}

			env.On("HTTPRequest", tc.url).Return([]byte(tc.response), func() error {
				if tc.isError {
					return errors.New("error")
				}
				return nil
			}())

			cs := &HTTP{
				base: base{
					env:   env,
					props: props,
				},
			}

			enabled := cs.Enabled()
			assert.Equal(t, tc.expected, enabled)
			if enabled {
				assert.Equal(t, tc.responseID, cs.Result["id"])
			}
		})
	}
}

func TestHTTPSegmentTemplate(t *testing.T) {
	env := new(mock.Environment)
	props := properties.Map{
		URL: "https://jsonplaceholder.typicode.com/posts/1",
	}

	env.On("HTTPRequest", "https://jsonplaceholder.typicode.com/posts/1").Return([]byte(`{"key": "value"}`), nil)

	cs := &HTTP{
		base: base{
			env:   env,
			props: props,
		},
	}

	cs.Enabled()
	template := cs.Template()
	expectedTemplate := " {{ .Result }} "
	assert.Equal(t, expectedTemplate, template)
}
