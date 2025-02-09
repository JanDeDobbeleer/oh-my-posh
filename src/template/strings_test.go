package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrunc(t *testing.T) {
	cases := []struct {
		Case        string
		Expected    string
		Template    string
		ShouldError bool
	}{
		{Case: "5 length integer", Expected: "Hello", Template: `{{ trunc 5 "Hello World" }}`},
		{Case: "5 length stringteger", Expected: "Hello", Template: `{{ trunc "5" "Hello World" }}`},
		{Case: "5 length float", Expected: "Hello", Template: `{{ trunc 5.0 "Hello World" }}`},
		{Case: "invalid", ShouldError: true, Template: `{{ trunc "foo" "Hello World" }}`},
		{Case: "smaller than length", Expected: "Hello World", Template: `{{ trunc 20 "Hello World" }}`},
		{Case: "negative", Expected: "ld", Template: `{{ trunc -2 "Hello World" }}`},
	}

	for _, tc := range cases {
		tmpl := &Text{
			Template: tc.Template,
			Context:  nil,
		}

		text, err := tmpl.Render()
		if tc.ShouldError {
			assert.Error(t, err)
			continue
		}

		assert.Equal(t, tc.Expected, text, tc.Case)
	}
}

func TestTruncE(t *testing.T) {
	cases := []struct {
		name      string
		length    any
		input     string
		expected  string
		wantPanic bool
	}{
		{
			name:     "normal truncation",
			length:   5,
			input:    "hello world",
			expected: "hell…",
		},
		{
			name:     "no truncation needed",
			length:   20,
			input:    "short",
			expected: "short",
		},
		{
			name:     "negative length",
			length:   -3,
			input:    "hello world",
			expected: "…ld",
		},
		{
			name:     "zero length",
			length:   0,
			input:    "hello",
			expected: "…",
		},
		{
			name:     "unicode characters",
			length:   4,
			input:    "你好世界",
			expected: "你好世…",
		},
		{
			name:     "empty string",
			length:   5,
			input:    "",
			expected: "",
		},
		{
			name:      "invalid length type",
			length:    "invalid",
			input:     "hello",
			wantPanic: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("expected panic but got none")
					}
				}()
			}

			result := truncE(tc.length, tc.input)
			if result != tc.expected {
				t.Errorf("expected %q but got %q", tc.expected, result)
			}
		})
	}
}
