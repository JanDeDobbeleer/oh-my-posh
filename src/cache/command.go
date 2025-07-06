package cache

import (
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
)

type Command struct {
	Commands *maps.Concurrent[string]
}

func (c *Command) Set(command, path string) {
	c.Commands.Set(command, path)
}

func (c *Command) Get(command string) (string, bool) {
	cacheCommand, found := c.Commands.Get(command)
	if !found {
		return "", false
	}

	return cacheCommand, true
}
