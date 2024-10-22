package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Connection struct {
	base

	runtime.Connection
}

const (
	Type properties.Property = "type"
)

func (c *Connection) Template() string {
	return " {{ if eq .Type \"wifi\"}}\uf1eb{{ else if eq .Type \"ethernet\"}}\ueba9{{ end }} "
}

func (c *Connection) Enabled() bool {
	types := c.props.GetString(Type, "wifi|ethernet")
	connectionTypes := strings.Split(types, "|")
	for _, connectionType := range connectionTypes {
		network, err := c.env.Connection(runtime.ConnectionType(connectionType))
		if err != nil {
			continue
		}
		c.Connection = *network
		return true
	}
	return false
}
