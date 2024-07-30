package segments

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

func TestWTTrackedTime(t *testing.T) {
	cases := []struct {
		Case     string
		Seconds  int
		Expected string
		Template string
		Error    error
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
			Case:     "no cache 2h 45m",
			Seconds:  9900,
			Expected: "2h 45m",
		},
		{
			Case:     "api error",
			Seconds:  2,
			Expected: "0s",
			Error:    errors.New("api error"),
		},
	}

	for _, tc := range cases {
		env := &mock.Environment{}
		response := fmt.Sprintf(`{"cumulative_total": {"seconds": %.2f, "text": "x"}}`, float64(tc.Seconds))

		env.On("HTTPRequest", FAKEAPIURL).Return([]byte(response), tc.Error)

		env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)
		env.On("Flags").Return(&runtime.Flags{})

		env.On("TemplateCache").Return(&cache.Template{
			Env: map[string]string{"HELLO": "hello"},
		})

		w := &Wakatime{
			props: properties.Map{
				URL: FAKEAPIURL,
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
		env := &mock.Environment{}

		env.On("Error", testify_.Anything)
		env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)
		env.On("TemplateCache").Return(&cache.Template{
			Env: map[string]string{"HELLO": "hello"},
		})
		env.On("Flags").Return(&runtime.Flags{})

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
