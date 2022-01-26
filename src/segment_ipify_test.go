package main

import (
	"errors"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	IPIFYAPIURL = "https://api.ipify.org"
)

func TestIpifySegment(t *testing.T) {
	cases := []struct {
		Case            string
		Response        string
		ExpectedString  string
		ExpectedEnabled bool
		Template        string
		Error           error
	}{
		{
			Case:            "IPv4",
			Response:        `127.0.0.1`,
			ExpectedString:  "127.0.0.1",
			ExpectedEnabled: true,
		},
		{
			Case:            "IPv6 (with template)",
			Response:        `0000:aaaa:1111:bbbb:2222:cccc:3333:dddd`,
			ExpectedString:  "Ext. IP: 0000:aaaa:1111:bbbb:2222:cccc:3333:dddd",
			ExpectedEnabled: true,
			Template:        "Ext. IP: {{.IP}}",
		},
		{
			Case:            "Error in retrieving data",
			Response:        "nonsense",
			Error:           errors.New("Something went wrong"),
			ExpectedEnabled: false,
		},
	}

	for _, tc := range cases {
		env := &mock.MockedEnvironment{}
		props := properties.Map{
			CacheTimeout: 0,
		}
		env.On("HTTPRequest", IPIFYAPIURL).Return([]byte(tc.Response), tc.Error)

		ipify := &IPify{
			props: props,
			env:   env,
		}

		enabled := ipify.enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if !enabled {
			continue
		}

		if tc.Template == "" {
			tc.Template = ipify.template()
		}
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, ipify), tc.Case)
	}
}
