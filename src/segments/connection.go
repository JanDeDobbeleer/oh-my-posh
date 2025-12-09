package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

type Connection struct {
	Base

	runtime.Connection
}

const (
	Type options.Option = "type"
)

func (c *Connection) Template() string {
	return " {{ if eq .Type \"wifi\"}}\uf1eb{{ else if eq .Type \"ethernet\"}}\ueba9{{ end }} "
}

func (c *Connection) Enabled() bool {
	types := c.options.String(Type, "wifi|ethernet")
	connectionTypes := strings.SplitSeq(types, "|")
	for connectionType := range connectionTypes {
		network, err := c.env.Connection(runtime.ConnectionType(connectionType))
		if err != nil {
			continue
		}
		c.Connection = *network
		return true
	}
	return false
}
