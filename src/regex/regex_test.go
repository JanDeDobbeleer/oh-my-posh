package regex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindStringMatch(t *testing.T) {
	cases := []struct {
		Case     string
		Pattern  string
		Text     string
		Expected string
		Index    int
	}{
		{
			Case:     "Full match at index 0",
			Pattern:  `\w+`,
			Text:     "hello",
			Index:    0,
			Expected: "hello",
		},
		{
			Case:     "Capture group at index 1",
			Pattern:  `hello (\w+)`,
			Text:     "hello world",
			Index:    1,
			Expected: "world",
		},
		{
			Case:     "No matches returns original text",
			Pattern:  `\d+`,
			Text:     "hello",
			Index:    0,
			Expected: "hello",
		},
		{
			Case:     "Invalid pattern returns original text",
			Pattern:  `[invalid`,
			Text:     "hello",
			Index:    0,
			Expected: "hello",
		},
		{
			Case:     "Empty text returns empty string",
			Pattern:  `\w+`,
			Text:     "",
			Index:    0,
			Expected: "",
		},
		{
			Case:     "Index out of bounds returns original text",
			Pattern:  `(\w+)`,
			Text:     "hello",
			Index:    2,
			Expected: "hello",
		},
		{
			Case:     "Multiple capture groups",
			Pattern:  `(\w+)\s(\w+)`,
			Text:     "hello world",
			Index:    2,
			Expected: "world",
		},
	}

	for _, tc := range cases {
		got, _ := FindStringMatch(tc.Pattern, tc.Text, tc.Index)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestFindStringMatchIndex(t *testing.T) {
	cases := []struct {
		Case       string
		Pattern    string
		Text       string
		Expected   string
		Index      int
		ExpectedOK bool
	}{
		{
			Case:       "Full match at index 0",
			Pattern:    `\w+`,
			Text:       "hello",
			Index:      0,
			Expected:   "hello",
			ExpectedOK: true,
		},
		{
			Case:       "Capture group at index 1",
			Pattern:    `hello (\w+)`,
			Text:       "hello world",
			Index:      1,
			Expected:   "world",
			ExpectedOK: true,
		},
		{
			Case:       "No matches",
			Pattern:    `\d+`,
			Text:       "hello",
			Index:      0,
			ExpectedOK: false,
		},
		{
			Case:       "Invalid pattern",
			Pattern:    `[invalid`,
			Text:       "hello",
			Index:      0,
			ExpectedOK: false,
		},
		{
			Case:       "Index out of bounds",
			Pattern:    `(\w+)`,
			Text:       "hello",
			Index:      2,
			ExpectedOK: false,
		},
		{
			Case:       "Non-participating capture group",
			Pattern:    `(a)|(b)`,
			Text:       "a",
			Index:      2,
			ExpectedOK: false,
		},
		{
			Case:       "Duplicate text: index points at the actually captured occurrence",
			Pattern:    `.*/(src)/deep`,
			Text:       "/home/src/project/src/deep",
			Index:      1,
			Expected:   "src",
			ExpectedOK: true,
		},
	}

	for _, tc := range cases {
		start, end, ok := FindStringMatchIndex(tc.Pattern, tc.Text, tc.Index)
		assert.Equal(t, tc.ExpectedOK, ok, tc.Case)
		if !ok {
			continue
		}

		assert.Equal(t, tc.Expected, tc.Text[start:end], tc.Case)
	}
}
