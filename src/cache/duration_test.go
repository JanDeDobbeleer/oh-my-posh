package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeconds(t *testing.T) {
	cases := []struct {
		Case     string
		Duration Duration
		Expected int
	}{
		{
			Case:     "2 seconds",
			Duration: "2s",
			Expected: 2,
		},
		{
			Case:     "1 minute",
			Duration: "1m",
			Expected: 60,
		},
		{
			Case:     "2 hours",
			Duration: "2h",
			Expected: 7200,
		},
		{
			Case:     "2 days",
			Duration: "48h",
			Expected: 172800,
		},
		{
			Case:     "invalid",
			Duration: "foo",
			Expected: 0,
		},
		{
			Case:     "1 fortnight",
			Duration: "1fortnight",
			Expected: 0,
		},
		{
			Case:     "infinite",
			Duration: "infinite",
			Expected: -1,
		},
	}
	for _, tc := range cases {
		got := tc.Duration.Seconds()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
