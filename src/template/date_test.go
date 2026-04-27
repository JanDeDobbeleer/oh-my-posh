package template

import (
	"fmt"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

// knownEpoch is 2019-06-13 20:39:39 UTC, matching sprig's own test fixture.
const knownEpoch = int64(1560458379)

func TestDateFromStringEpoch(t *testing.T) {
	// This is the core regression: unixEpoch returns a string, and piping that
	// string into `date` must produce the correct formatted date, not time.Now().
	epochStr := fmt.Sprintf("%d", knownEpoch)

	// Use date_in_zone with UTC so tests are timezone-independent.
	cases := []struct {
		Case     string
		Expected string
		Template string
		Context  any
	}{
		{
			Case:     "string epoch via date_in_zone",
			Expected: "13 Jun 19 20:39 +0000",
			Template: `{{ date_in_zone "02 Jan 06 15:04 -0700" .Epoch "UTC" }}`,
			Context:  struct{ Epoch string }{Epoch: epochStr},
		},
		{
			Case:     "int64 epoch via date_in_zone",
			Expected: "13 Jun 19 20:39 +0000",
			Template: `{{ date_in_zone "02 Jan 06 15:04 -0700" .Epoch "UTC" }}`,
			Context:  struct{ Epoch int64 }{Epoch: knownEpoch},
		},
		{
			Case:     "time.Time via date_in_zone",
			Expected: "13 Jun 19 20:39 +0000",
			Template: `{{ date_in_zone "02 Jan 06 15:04 -0700" .Epoch "UTC" }}`,
			Context:  struct{ Epoch time.Time }{Epoch: time.Unix(knownEpoch, 0).UTC()},
		},
		{
			Case:     "string epoch direct call dateInZone",
			Expected: "13 Jun 19 20:39 +0000",
			Template: `{{ dateInZone "02 Jan 06 15:04 -0700" .Epoch "UTC" }}`,
			Context:  struct{ Epoch string }{Epoch: epochStr},
		},
		{
			Case:     "htmlDate with zero string epoch",
			Expected: "1970-01-01",
			Template: `{{ htmlDateInZone .Epoch "UTC" }}`,
			Context:  struct{ Epoch string }{Epoch: "0"},
		},
		{
			Case:     "non-numeric string falls back to current time (not panic)",
			Template: `{{ date_in_zone "02 Jan 06" .Epoch "UTC" }}`,
			Context:  struct{ Epoch string }{Epoch: "not-a-number"},
			// Expected is time-dependent; we just verify no error is returned.
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Shell").Return("foo")
		Cache = new(cache.Template)
		Init(env, nil, nil)

		text, err := Render(tc.Template, tc.Context)
		assert.NoError(t, err, tc.Case)
		if tc.Expected != "" {
			assert.Equal(t, tc.Expected, text, tc.Case)
		} else {
			assert.NotEmpty(t, text, tc.Case)
		}
	}
}
