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
		expected    any
		name        string
		url         string
		method      string
		timeout     int
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
		{
			name:        "Timeout test",
			url:         "https://apifastmock.com/mock/SsGLwD_ht5I2_9NebQH0N5NaOaRMwMkww4eU37UC9s8fHBzBIcYERO5MSzH6LUWGevCemFq5LWgP8bG7aalV_Q0VYUFEakiqZ8xyHwszmnmqzRkZ5UGm9cjvjITfUNCB8K15SzW021Dm3yipsANIMNTO4QGGYdvNRcgx89w",
			method:      "GET",
			shouldError: false,
			timeout:     1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := new(mock.Environment)
			props := properties.Map{
				URL:     tc.url,
				METHOD:  tc.method,
				TIMEOUT: tc.timeout,
			}

			env.On("HTTPRequest", tc.url).Return([]byte(tc.response), func() error {
				if tc.shouldError {
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

			_ = cs.Enabled()
			assert.Equal(t, tc.expected, cs.Body["id"], tc.name)
		})
	}
}
