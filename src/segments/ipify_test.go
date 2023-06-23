package segments

import (
	"errors"
	"net"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"

	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

const (
	IPIFYAPIURL = "https://api.ipify.org"
)

type mockedipAPI struct {
	mock2.Mock
}

func (s *mockedipAPI) Get() (*ipData, error) {
	args := s.Called()
	return args.Get(0).(*ipData), args.Error(1)
}

func TestIpifySegment(t *testing.T) {
	cases := []struct {
		Case            string
		IPDate          *ipData
		Error           error
		ExpectedString  string
		ExpectedEnabled bool
	}{
		{
			Case:            "IP data",
			IPDate:          &ipData{IP: "127.0.0.1"},
			ExpectedString:  "127.0.0.1",
			ExpectedEnabled: true,
		},
		{
			Case:            "Error",
			Error:           errors.New("network is unreachable"),
			ExpectedEnabled: false,
		},
		{
			Case:            "Offline",
			ExpectedString:  OFFLINE,
			Error:           &net.DNSError{IsNotFound: true},
			ExpectedEnabled: true,
		},
	}

	for _, tc := range cases {
		api := &mockedipAPI{}
		api.On("Get").Return(tc.IPDate, tc.Error)

		ipify := &IPify{
			api: api,
		}

		enabled := ipify.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)

		if !enabled {
			continue
		}

		assert.Equal(t, tc.ExpectedString, renderTemplate(&mock.MockedEnvironment{}, ipify.Template(), ipify), tc.Case)
	}
}
