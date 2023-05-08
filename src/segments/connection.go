package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Connection struct {
	props properties.Properties
	env   platform.Environment

	platform.Connection
}

const (
	Type properties.Property = "type"
)

func (c *Connection) Template() string {
	return " {{ if eq .Type \"wifi\"}}\uf1eb{{ else if eq .Type \"ethernet\"}}\U000f0200{{ end }} "
}

func (c *Connection) Enabled() bool {
	types := c.props.GetString(Type, "wifi|ethernet")
	connectionTypes := strings.Split(types, "|")
	for _, connectionType := range connectionTypes {
		network, err := c.env.Connection(platform.ConnectionType(connectionType))
		if err != nil {
			continue
		}
		c.Connection = *network
		return true
	}
	return false
}

func (c *Connection) Init(props properties.Properties, env platform.Environment) {
	c.props = props
	c.env = env
}
