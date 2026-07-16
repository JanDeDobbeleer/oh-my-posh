package harness

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// stringsOf converts a [][]byte, as returned by splitDSRQueries, into []string for easier
// assertions.
func stringsOf(bs [][]byte) []string {
	if bs == nil {
		return nil
	}

	out := make([]string, len(bs))
	for i, b := range bs {
		out[i] = string(b)
	}

	return out
}

// TestSplitDSRQueries covers splitDSRQueries given a single, already-complete window (no
// carry involved from a previous read). See TestSplitDSRQueriesAcrossChunks for the case
// where a query is split across two pty reads.
func TestSplitDSRQueries(t *testing.T) {
	cases := []struct {
		Case            string
		Data            string
		ExpectedQueries []string
		ExpectedRest    string
		ExpectedCarry   string
	}{
		{
			Case:            "no query",
			Data:            "plain output, nothing special",
			ExpectedQueries: nil,
			ExpectedRest:    "plain output, nothing special",
			ExpectedCarry:   "",
		},
		{
			Case:            "one query mid-chunk with distinct output before and after",
			Data:            "before-text\x1b[6nafter-text",
			ExpectedQueries: []string{"before-text\x1b[6n"},
			ExpectedRest:    "after-text",
			ExpectedCarry:   "",
		},
		{
			Case:            "multiple queries in one chunk",
			Data:            "A\x1b[6nB\x1b[6nC",
			ExpectedQueries: []string{"A\x1b[6n", "B\x1b[6n"},
			ExpectedRest:    "C",
			ExpectedCarry:   "",
		},
		{
			Case:            "trailing 3-byte prefix held back",
			Data:            "trailing-text\x1b[6",
			ExpectedQueries: nil,
			ExpectedRest:    "trailing-text",
			ExpectedCarry:   "\x1b[6",
		},
		{
			Case:            "trailing 2-byte prefix held back",
			Data:            "trailing-text\x1b[",
			ExpectedQueries: nil,
			ExpectedRest:    "trailing-text",
			ExpectedCarry:   "\x1b[",
		},
		{
			Case:            "trailing 1-byte prefix held back",
			Data:            "trailing-text\x1b",
			ExpectedQueries: nil,
			ExpectedRest:    "trailing-text",
			ExpectedCarry:   "\x1b",
		},
		{
			Case:            "trailing non-prefix not held back",
			Data:            "trailing-text-abc",
			ExpectedQueries: nil,
			ExpectedRest:    "trailing-text-abc",
			ExpectedCarry:   "",
		},
	}

	for _, c := range cases {
		t.Run(c.Case, func(t *testing.T) {
			queries, rest, carry := splitDSRQueries([]byte(c.Data))

			assert.Equal(t, c.ExpectedQueries, stringsOf(queries), "queries")
			assert.Equal(t, c.ExpectedRest, string(rest), "rest")
			assert.Equal(t, c.ExpectedCarry, string(carry), "carry")
		})
	}
}

// TestSplitDSRQueriesAcrossChunks drives splitDSRQueries twice per case, exactly as read()
// does: the carry from the first call is prefixed onto the second chunk before the second
// call. It covers every possible split point of a 4-byte dsrRequest ("\x1b[6n") across two
// reads: after 1, 2, and 3 bytes of the sequence have arrived.
func TestSplitDSRQueriesAcrossChunks(t *testing.T) {
	cases := []struct {
		Case             string
		Chunk1           string
		Chunk2           string
		ExpectedCarry1   string
		ExpectedRest1    string
		ExpectedQueries2 []string
		ExpectedRest2    string
		ExpectedCarry2   string
	}{
		{
			Case:             "split after 1 byte",
			Chunk1:           "before\x1b",
			Chunk2:           "[6nafter",
			ExpectedCarry1:   "\x1b",
			ExpectedRest1:    "before",
			ExpectedQueries2: []string{"\x1b[6n"},
			ExpectedRest2:    "after",
			ExpectedCarry2:   "",
		},
		{
			Case:             "split after 2 bytes",
			Chunk1:           "before\x1b[",
			Chunk2:           "6nafter",
			ExpectedCarry1:   "\x1b[",
			ExpectedRest1:    "before",
			ExpectedQueries2: []string{"\x1b[6n"},
			ExpectedRest2:    "after",
			ExpectedCarry2:   "",
		},
		{
			Case:             "split after 3 bytes",
			Chunk1:           "before\x1b[6",
			Chunk2:           "nafter",
			ExpectedCarry1:   "\x1b[6",
			ExpectedRest1:    "before",
			ExpectedQueries2: []string{"\x1b[6n"},
			ExpectedRest2:    "after",
			ExpectedCarry2:   "",
		},
	}

	for _, c := range cases {
		t.Run(c.Case, func(t *testing.T) {
			queries1, rest1, carry1 := splitDSRQueries([]byte(c.Chunk1))

			assert.Empty(t, queries1, "no complete query expected after the first chunk")
			assert.Equal(t, c.ExpectedRest1, string(rest1), "rest after chunk 1")
			assert.Equal(t, c.ExpectedCarry1, string(carry1), "carry after chunk 1")

			window2 := append(append([]byte(nil), carry1...), []byte(c.Chunk2)...)
			queries2, rest2, carry2 := splitDSRQueries(window2)

			assert.Equal(t, c.ExpectedQueries2, stringsOf(queries2), "queries after chunk 2")
			assert.Equal(t, c.ExpectedRest2, string(rest2), "rest after chunk 2")
			assert.Equal(t, c.ExpectedCarry2, string(carry2), "carry after chunk 2")
		})
	}
}
