package segments

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

func TestConnection(t *testing.T) {
	type connectionResponse struct {
		Connection *runtime.Connection
		Error      error
	}
	cases := []struct {
		Case            string
		ExpectedString  string
		ConnectionType  string
		Connections     []*connectionResponse
		ExpectedEnabled bool
	}{
		{
			Case:            "WiFi only, enabled",
			ExpectedString:  "\uf1eb",
			ExpectedEnabled: true,
			ConnectionType:  "wifi",
			Connections: []*connectionResponse{
				{
					Connection: &runtime.Connection{
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
					Connection: &runtime.Connection{
						Type: runtime.WIFI,
					},
					Error: fmt.Errorf("no connection"),
				},
			},
		},
		{
			Case:            "WiFi and Ethernet, enabled",
			ConnectionType:  "wifi|ethernet",
			ExpectedString:  "\ueba9",
			ExpectedEnabled: true,
			Connections: []*connectionResponse{
				{
					Connection: &runtime.Connection{
						Type: runtime.WIFI,
					},
					Error: fmt.Errorf("no connection"),
				},
				{
					Connection: &runtime.Connection{
						Type: runtime.ETHERNET,
					},
				},
			},
		},
		{
			Case:           "WiFi and Ethernet, disabled",
			ConnectionType: "wifi|ethernet",
			Connections: []*connectionResponse{
				{
					Connection: &runtime.Connection{
						Type: runtime.WIFI,
					},
					Error: fmt.Errorf("no connection"),
				},
				{
					Connection: &runtime.Connection{
						Type: runtime.ETHERNET,
					},
					Error: fmt.Errorf("no connection"),
				},
			},
		},
	}
	for _, tc := range cases {
		env := &mock.Environment{}
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
