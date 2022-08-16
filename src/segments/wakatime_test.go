package segments

import (
	"errors"
	"fmt"
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
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
		env := &mock.MockedEnvironment{}

		response := fmt.Sprintf(`{"cummulative_total": {"seconds": %.2f, "text": "x"}}`, float64(tc.Seconds))

		env.On("HTTPRequest", FAKEAPIURL).Return([]byte(response), tc.Error)

		cache := &mock.MockedCache{}
		cache.On("Get", FAKEAPIURL).Return(response, !tc.CacheFoundFail)
		cache.On("Set", FAKEAPIURL, response, tc.CacheTimeout).Return()
		env.On("Cache").Return(cache)

		env.On("TemplateCache").Return(&environment.TemplateCache{
			Env: map[string]string{"HELLO": "hello"},
		})

		w := &Wakatime{
			props: properties.Map{
				properties.CacheTimeout: tc.CacheTimeout,
				URL:                     FAKEAPIURL,
			},
			env: env,
		}

		assert.ErrorIs(t, tc.Error, w.setAPIData(), tc.Case+" - Error")
		assert.Equal(t, tc.Expected, renderTemplate(env, w.Template(), w), tc.Case+" - String")
	}
}

func TestWTGetUrl(t *testing.T) {
	cases := []struct {
		Case        string
		Expected    string
		URL         string
		ShouldError bool
	}{
		{
			Case:     "no template",
			Expected: "test",
			URL:      "test",
		},
		{
			Case:     "template",
			URL:      "{{ .Env.HELLO }} world",
			Expected: "hello world",
		},
		{
			Case:        "error",
			URL:         "{{ .BURR }}",
			ShouldError: true,
		},
	}

	for _, tc := range cases {
		env := &mock.MockedEnvironment{}

		env.On("Log", mock2.Anything, mock2.Anything, mock2.Anything)
		env.On("TemplateCache").Return(&environment.TemplateCache{
			Env: map[string]string{"HELLO": "hello"},
		})

		w := &Wakatime{
			props: properties.Map{
				URL: tc.URL,
			},
			env: env,
		}

		got, err := w.getURL()

		if tc.ShouldError {
			assert.Error(t, err, tc.Case)
			continue
		}
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
