package main

import "strings"

type command struct {
	props properties
	env   environmentInfo
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
	if !c.env.hasCommand(shell) {
		return false
	}
	command := c.props.getString(Command, "echo no command specified")
	if strings.Contains(command, "||") {
		commands := strings.Split(command, "||")
		for _, cmd := range commands {
			output := c.env.runShellCommand(shell, cmd)
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
			output += c.env.runShellCommand(shell, cmd)
		}
		c.value = output
		return c.value != ""
	}
	c.value = c.env.runShellCommand(shell, command)
	return c.value != ""
}

func (c *command) string() string {
	return c.value
}

func (c *command) init(props properties, env environmentInfo) {
	c.props = props
	c.env = env
}
