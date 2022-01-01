package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWiFiSegment(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		ExpectedEnabled bool
		Network         *wifiInfo
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
			Network: &wifiInfo{
				SSID: "pretty fly for a wifi",
			},
		},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getPlatform", nil).Return(windowsPlatform)
		env.On("isWsl", nil).Return(false)
		env.On("getWifiNetwork", nil).Return(tc.Network, tc.WifiError)

		w := &wifi{
			env: env,
			props: properties{
				DisplayError:    tc.DisplayError,
				SegmentTemplate: "{{ if .Error }}{{ .Error }}{{ else }}{{ .SSID }}{{ end }}",
			},
		}

		assert.Equal(t, tc.ExpectedEnabled, w.enabled(), tc.Case)
		if tc.Network != nil || tc.DisplayError {
			assert.Equal(t, tc.ExpectedString, w.string(), tc.Case)
		}
	}
}
