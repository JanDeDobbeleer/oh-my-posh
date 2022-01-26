package segments

import (
	"errors"
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWiFiSegment(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		ExpectedEnabled bool
		Network         *environment.WifiInfo
		WifiError       error
		DisplayError    bool
	}{
		{
			Case: "No error and nil network",
		},
		{
			Case:      "Error and nil network",
			WifiError: errors.New("oh noes"),
		},
		{
			Case:            "Display error and nil network",
			WifiError:       errors.New("oh noes"),
			ExpectedString:  "oh noes",
			DisplayError:    true,
			ExpectedEnabled: true,
		},
		{
			Case:            "Display wifi state",
			ExpectedString:  "pretty fly for a wifi",
			ExpectedEnabled: true,
			Network: &environment.WifiInfo{
				SSID: "pretty fly for a wifi",
			},
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Platform").Return(environment.WindowsPlatform)
		env.On("IsWsl").Return(false)
		env.On("WifiNetwork").Return(tc.Network, tc.WifiError)

		w := &Wifi{
			env: env,
			props: properties.Map{
				properties.DisplayError: tc.DisplayError,
			},
		}

		assert.Equal(t, tc.ExpectedEnabled, w.Enabled(), tc.Case)
		if tc.Network != nil || tc.DisplayError {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, "{{ if .Error }}{{ .Error }}{{ else }}{{ .SSID }}{{ end }}", w), tc.Case)
		}
	}
}
