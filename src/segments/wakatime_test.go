package segments

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

func TestWTTrackedTime(t *testing.T) {
	cases := []struct {
		Error          error
		Case           string
		Expected       string
		Template       string
		Seconds        int
		CacheTimeout   int
		CacheFoundFail bool
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
