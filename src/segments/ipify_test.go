package segments

import (
	"errors"
	"net"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

type mockedipAPI struct {
	testify_.Mock
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
			api:     api,
			env:     &mock.Environment{},
			options: options.Map{},
		}

		enabled := ipify.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)

		cache.Device.DeleteAll()

		if !enabled {
			continue
		}

		assert.Equal(t, tc.ExpectedString, renderTemplate(&mock.Environment{}, ipify.Template(), ipify), tc.Case)
	}
}

func TestIpifyURL(t *testing.T) {
	cases := []struct {
		Case        string
		Options     options.Map
		ExpectedURL string
	}{
		{
			Case:        "Default",
			Options:     options.Map{},
			ExpectedURL: "https://api.ipify.org?format=json",
		},
		{
			Case:        "IPv6",
			Options:     options.Map{URL: "https://api64.ipify.org"},
			ExpectedURL: "https://api64.ipify.org?format=json",
		},
	}

	for _, tc := range cases {
		env := &mock.Environment{}
		env.On("HTTPRequest", tc.ExpectedURL).Return([]byte(`{"ip":"127.0.0.1"}`), nil)

		ipify := &IPify{
			env:     env,
			options: tc.Options,
		}

		assert.True(t, ipify.Enabled(), tc.Case)
		assert.Equal(t, "127.0.0.1", ipify.IP, tc.Case)

		cache.Device.DeleteAll()
	}
}
