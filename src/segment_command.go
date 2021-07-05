package main

import (
	"strings"

	"oh-my-posh/runtime"
)

type command struct {
	props *properties
	env   runtime.Environment
	value string
}

const (
	// ExecutableShell to execute command in
	ExecutableShell Property = "shell"
	// Command to execute
	Command Property = "command"
)

func (c *command) enabled() bool {
	shell := c.props.getString(ExecutableShell, "bash")
	if !c.env.HasCommand(shell) {
		return false
	}
	command := c.props.getString(Command, "echo no command specified")
	if strings.Contains(command, "||") {
		commands := strings.Split(command, "||")
		for _, cmd := range commands {
			output := c.env.RunShellCommand(shell, cmd)
			if output != "" {
				c.value = output
				return true
			}
		}
	}
	if strings.Contains(command, "&&") {
		var output string
		commands := strings.Split(command, "&&")
		for _, cmd := range commands {
			output += c.env.RunShellCommand(shell, cmd)
		}
		c.value = output
		return c.value != ""
	}
	c.value = c.env.RunShellCommand(shell, command)
	return c.value != ""
}

func (c *command) string() string {
	return c.value
}

func (c *command) init(props *properties, env runtime.Environment) {
	c.props = props
	c.env = env
}
