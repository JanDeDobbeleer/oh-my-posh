package segments

import (
	"fmt"
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnection(t *testing.T) {
	type connectionResponse struct {
		Connection *environment.Connection
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
					Connection: &environment.Connection{
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
					Connection: &environment.Connection{
						Type: environment.WIFI,
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
					Connection: &environment.Connection{
						Type: environment.WIFI,
					},
					Error: fmt.Errorf("no connection"),
				},
				{
					Connection: &environment.Connection{
						Type: environment.ETHERNET,
					},
				},
			},
		},
		{
			Case:           "WiFi and Ethernet, disabled",
			ConnectionType: "wifi|ethernet",
			Connections: []*connectionResponse{
				{
					Connection: &environment.Connection{
						Type: environment.WIFI,
					},
					Error: fmt.Errorf("no connection"),
				},
				{
					Connection: &environment.Connection{
						Type: environment.ETHERNET,
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
