package text

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripJSONComments(t *testing.T) {
	cases := []struct {
		Case     string
		Input    string
		Expected string
	}{
		{
			Case:     "no comments",
			Input:    `{"a": 1}`,
			Expected: `{"a": 1}`,
		},
		{
			Case:     "line comment",
			Input:    "{\n  // comment\n  \"a\": 1\n}",
			Expected: "{\n  \n  \"a\": 1\n}",
		},
		{
			Case:     "block comment",
			Input:    `{"a": /* inline */ 1}`,
			Expected: `{"a":  1}`,
		},
		{
			Case:     "comment markers inside strings survive",
			Input:    `{"url": "https://example.com", "glob": "/*"}`,
			Expected: `{"url": "https://example.com", "glob": "/*"}`,
		},
		{
			Case:     "escaped quote inside string",
			Input:    `{"a": "say \"hi\" // not a comment"}`,
			Expected: `{"a": "say \"hi\" // not a comment"}`,
		},
		{
			Case:     "unterminated block comment",
			Input:    `{"a": 1} /* trailing`,
			Expected: `{"a": 1}`,
		},
		{
			Case:     "trailing line comment without newline",
			Input:    `{"a": 1} // done`,
			Expected: `{"a": 1}`,
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.Expected, StripJSONComments(tc.Input), tc.Case)
	}

	// the result must stay valid JSON for the consumers that unmarshal it
	jsonc := "{\n  // leading\n  \"theme\": \"dark\", /* block */\n  \"count\": 3 // trailing\n}"
	var parsed map[string]any
	assert.NoError(t, json.Unmarshal([]byte(StripJSONComments(jsonc)), &parsed))
	assert.Equal(t, "dark", parsed["theme"])
	assert.Equal(t, float64(3), parsed["count"])
}
