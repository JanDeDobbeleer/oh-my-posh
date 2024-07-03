package cache

import (
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
)

type Command struct {
	Commands *maps.Concurrent
}

func (c *Command) Set(command, path string) {
	c.Commands.Set(command, path)
}

func (c *Command) Get(command string) (string, bool) {
	cacheCommand, found := c.Commands.Get(command)
	if !found {
		return "", false
	}
	command, ok := cacheCommand.(string)
	return command, ok
}
