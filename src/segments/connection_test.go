package segments

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/mock"
	"github.com/jandedobbeleer/oh-my-posh/platform"
	"github.com/jandedobbeleer/oh-my-posh/properties"

	"github.com/stretchr/testify/assert"
)

func TestConnection(t *testing.T) {
	type connectionResponse struct {
		Connection *platform.Connection
		Error      error
	}
	cases := []struct {
		Case            string
		ExpectedString  string
		ExpectedEnabled bool
		ConnectionType  string
		Connections     []*connectionResponse
	}{
		{
			Case:            "WiFi only, enabled",
			ExpectedString:  "\uf1eb",
			ExpectedEnabled: true,
			ConnectionType:  "wifi",
			Connections: []*connectionResponse{
				{
					Connection: &platform.Connection{
						Name: "WiFi",
						Type: "wifi",
					},
				},
			},
		},
		{
			Case:           "WiFi only, disabled",
			ConnectionType: "wifi",
			Connections: []*connectionResponse{
				{
					Connection: &platform.Connection{
						Type: platform.WIFI,
					},
					Error: fmt.Errorf("no connection"),
				},
			},
		},
		{
			Case:            "WiFi and Ethernet, enabled",
			ConnectionType:  "wifi|ethernet",
			ExpectedString:  "\uf6ff",
			ExpectedEnabled: true,
			Connections: []*connectionResponse{
				{
					Connection: &platform.Connection{
						Type: platform.WIFI,
					},
					Error: fmt.Errorf("no connection"),
				},
				{
					Connection: &platform.Connection{
						Type: platform.ETHERNET,
					},
				},
			},
		},
		{
			Case:           "WiFi and Ethernet, disabled",
			ConnectionType: "wifi|ethernet",
			Connections: []*connectionResponse{
				{
					Connection: &platform.Connection{
						Type: platform.WIFI,
					},
					Error: fmt.Errorf("no connection"),
				},
				{
					Connection: &platform.Connection{
						Type: platform.ETHERNET,
					},
					Error: fmt.Errorf("no connection"),
				},
			},
		},
	}
	for _, tc := range cases {
		env := &mock.MockedEnvironment{}
		for _, con := range tc.Connections {
			env.On("Connection", con.Connection.Type).Return(con.Connection, con.Error)
		}
		c := &Connection{
			env: env,
			props: &properties.Map{
				Type: tc.ConnectionType,
			},
		}
		assert.Equal(t, tc.ExpectedEnabled, c.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, c.Template(), c), fmt.Sprintf("Failed in case: %s", tc.Case))
		}
	}
}
