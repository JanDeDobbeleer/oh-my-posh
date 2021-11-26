package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type netshStringArgs struct {
	state          string
	ssid           string
	radioType      string
	authentication string
	channel        int
	receiveRate    int
	transmitRate   int
	signal         int
}

func getNetshString(args *netshStringArgs) string {
	const netshString string = `
	There is 1 interface on the system:

	Name                   : Wi-Fi
	Description            : Intel(R) Wireless-AC 9560 160MHz
	GUID                   : 6bb8def2-9af2-4bd4-8be2-6bd54e46bdc9
	Physical address       : d4:3b:04:e6:10:40
	State                  : %s
	SSID                   : %s
	BSSID                  : 5c:7d:7d:82:c5:73
	Network type           : Infrastructure
	Radio type             : %s
	Authentication         : %s
	Cipher                 : CCMP
	Connection mode        : Profile
	Channel                : %d
	Receive rate (Mbps)    : %d
	Transmit rate (Mbps)   : %d
	Signal                 : %d%%
	Profile                : ohsiggy

	Hosted network status  : Not available`

	return fmt.Sprintf(netshString, args.state, args.ssid, args.radioType, args.authentication, args.channel, args.receiveRate, args.transmitRate, args.signal)
}

func TestWiFiSegment(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		ExpectedEnabled bool
		CommandNotFound bool
		CommandOutput   string
		CommandError    error
		DisplayError    bool
		Template        string
		ExpectedState   string
	}{
		{
			Case:            "not enabled on windows when netsh command not found",
			ExpectedEnabled: false,
			ExpectedString:  "",
			CommandNotFound: true,
		},
		{
			Case:            "not enabled on windows when netsh command fails",
			ExpectedEnabled: false,
			ExpectedString:  "",
			CommandError:    errors.New("intentional testing failure"),
		},
		{
			Case:            "enabled on windows with DisplayError=true",
			ExpectedEnabled: true,
			ExpectedString:  "WIFI ERR: intentional testing failure",
			CommandError:    errors.New("intentional testing failure"),
			DisplayError:    true,
			Template:        "{{.State}}",
		},
		{
			Case:            "enabled on windows with every property in template",
			ExpectedEnabled: true,
			ExpectedString:  "connected testing 802.11ac WPA2-Personal 99 500 400 80",
			CommandOutput: getNetshString(&netshStringArgs{
				state:          "connected",
				ssid:           "testing",
				radioType:      "802.11ac",
				authentication: "WPA2-Personal",
				channel:        99,
				receiveRate:    500.0,
				transmitRate:   400.0,
				signal:         80,
			}),
			Template: "{{.State}} {{.SSID}} {{.RadioType}} {{.Authentication}} {{.Channel}} {{.ReceiveRate}} {{.TransmitRate}} {{.Signal}}",
		},
		{
			Case:            "enabled on windows but wifi not connected",
			ExpectedEnabled: true,
			ExpectedString:  "disconnected",
			CommandOutput: getNetshString(&netshStringArgs{
				state: "disconnected",
			}),
			Template: "{{if not .Connected}}{{.State}}{{end}}",
		},
		{
			Case:            "enabled on windows but template is invalid",
			ExpectedEnabled: true,
			ExpectedString:  "unable to create text based on template",
			CommandOutput:   getNetshString(&netshStringArgs{}),
			Template:        "{{.DoesNotExist}}",
		},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getPlatform", nil).Return(windowsPlatform)
		env.On("isWsl", nil).Return(false)
		env.On("hasCommand", "netsh.exe").Return(!tc.CommandNotFound)
		env.On("runCommand", mock.Anything, mock.Anything).Return(tc.CommandOutput, tc.CommandError)

		w := &wifi{
			env: env,
			props: map[Property]interface{}{
				DisplayError:    tc.DisplayError,
				SegmentTemplate: tc.Template,
			},
		}

		assert.Equal(t, tc.ExpectedEnabled, w.enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, w.string(), tc.Case)
	}
}
