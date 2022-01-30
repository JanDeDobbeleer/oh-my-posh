package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWTTrackedTime(t *testing.T) {
	cases := []struct {
		Case           string
		Seconds        int
		Expected       string
		Template       string
		CacheTimeout   int
		CacheFoundFail bool
		Error          error
	}{
		{
			Case:     "nothing tracked",
			Seconds:  0,
			Expected: "0s",
		},
		{
			Case:     "25 minutes",
			Seconds:  1500,
			Expected: "25m",
		},
		{
			Case:     "2 hours",
			Seconds:  7200,
			Expected: "2h",
		},
		{
			Case:     "2h 45m",
			Seconds:  9900,
			Expected: "2h 45m",
		},
		{
			Case:     "negative number",
			Seconds:  -9900,
			Expected: "2h 45m",
		},
		{
			Case:         "cache 2h 45m",
			Seconds:      9900,
			Expected:     "2h 45m",
			CacheTimeout: 20,
		},
		{
			Case:           "no cache 2h 45m",
			Seconds:        9900,
			Expected:       "2h 45m",
			CacheTimeout:   20,
			CacheFoundFail: true,
		},
		{
			Case:           "api error",
			Seconds:        2,
			Expected:       "0s",
			CacheTimeout:   20,
			CacheFoundFail: true,
			Error:          errors.New("api error"),
		},
	}

	for _, tc := range cases {
		env := &MockedEnvironment{}

		response := fmt.Sprintf(`{"cummulative_total": {"seconds": %.2f, "text": "x"}}`, float64(tc.Seconds))

		env.On("HTTPRequest", FAKEAPIURL).Return([]byte(response), tc.Error)

		cache := &MockedCache{}
		cache.On("get", FAKEAPIURL).Return(response, !tc.CacheFoundFail)
		cache.On("set", FAKEAPIURL, response, tc.CacheTimeout).Return()
		env.On("cache").Return(cache)
		env.onTemplate()

		w := &wakatime{
			props: properties{
				APIKey:       "key",
				CacheTimeout: tc.CacheTimeout,
				URL:          FAKEAPIURL,
			},
			env: env,
		}

		assert.ErrorIs(t, tc.Error, w.setAPIData(), tc.Case+" - Error")
		assert.Equal(t, tc.Expected, w.string(), tc.Case+" - String")
	}
}
