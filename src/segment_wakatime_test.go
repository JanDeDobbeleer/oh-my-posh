package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	WTAPIURL = "https://wakatime.com/api/v1/users/current/summaries?api_key=key&end=today&start=today"
)

func TestWTTrackedTime(t *testing.T) {
	cases := []struct {
		Case     string
		Hours    int
		Minutes  int
		Expected string
	}{
		{
			Case:     "0h 0m",
			Hours:    0,
			Minutes:  0,
			Expected: "0m",
		},
		{
			Case:     "0h 25m",
			Hours:    0,
			Minutes:  25,
			Expected: "25m",
		},
		{
			Case:     "2h 0m",
			Hours:    2,
			Minutes:  0,
			Expected: "2h 0m",
		},
		{
			Case:     "2h 45m",
			Hours:    2,
			Minutes:  45,
			Expected: "2h 45m",
		},
	}

	for _, tc := range cases {
		env := &MockedEnvironment{}

		sec := float64(tc.Hours*3600+tc.Minutes*60) + 30.123
		response := fmt.Sprintf(`{"cummulative_total": {"decimal": "x", "digital": "x", "seconds": %.2f, "text": "x"}}`, sec)

		env.On("doGet", WTAPIURL).Return([]byte(response), nil)

		w := &wakatime{
			props: map[Property]interface{}{
				APIKey:       "key",
				CacheTimeout: 0,
			},
			env: env,
		}

		assert.Nil(t, w.setStatus(), tc.Case+" - Error")
		assert.Equal(t, tc.Hours, w.Hours, tc.Case+" - Hours")
		assert.Equal(t, tc.Minutes, w.Minutes, tc.Case+" - Minutes")
		assert.Equal(t, tc.Expected, w.string(), tc.Case+" - String")
	}
}
