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
			Case:     "1 second",
			Duration: "1second",
			Expected: 1,
		},
		{
			Case:     "2 seconds",
			Duration: "2seconds",
			Expected: 2,
		},
		{
			Case:     "1 minute",
			Duration: "1minute",
			Expected: 60,
		},
		{
			Case:     "2 minutes",
			Duration: "2minutes",
			Expected: 120,
		},
		{
			Case:     "1 hour",
			Duration: "1hour",
			Expected: 3600,
		},
		{
			Case:     "2 hours",
			Duration: "2hours",
			Expected: 7200,
		},
		{
			Case:     "1 day",
			Duration: "1day",
			Expected: 86400,
		},
		{
			Case:     "2 days",
			Duration: "2days",
			Expected: 172800,
		},
		{
			Case:     "1 week",
			Duration: "1week",
			Expected: 604800,
		},
		{
			Case:     "2 weeks",
			Duration: "2weeks",
			Expected: 1209600,
		},
		{
			Case:     "1 month",
			Duration: "1month",
			Expected: 2592000,
		},
		{
			Case:     "2 months",
			Duration: "2month",
			Expected: 5184000,
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
