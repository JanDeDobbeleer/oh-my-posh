package main

import (
	"oh-my-posh/environment"
	"strings"
)

type command struct {
	props Properties
	env   environment.Environment

	Output string
}

const (
	// ExecutableShell to execute command in
	ExecutableShell Property = "shell"
	// Command to execute
	Command Property = "command"
)

func (c *command) template() string {
	return "{{ .Output }}"
}

func (c *command) enabled() bool {
	shell := c.props.getString(ExecutableShell, "bash")
	if !c.env.HasCommand(shell) {
		return false
	}
	command := c.props.getString(Command, "echo no command specified")
	if strings.Contains(command, "||") {
		commands := strings.Split(command, "||")
		for _, cmd := range commands {
			output := c.env.RunShellCommand(shell, strings.TrimSpace(cmd))
			if output != "" {
				c.Output = output
				return true
			}
		}
	}
	if strings.Contains(command, "&&") {
		var output string
		commands := strings.Split(command, "&&")
		for _, cmd := range commands {
			output += c.env.RunShellCommand(shell, strings.TrimSpace(cmd))
		}
		c.Output = output
		return c.Output != ""
	}
	c.Output = c.env.RunShellCommand(shell, strings.TrimSpace(command))
	return c.Output != ""
}

func (c *command) init(props Properties, env environment.Environment) {
	c.props = props
	c.env = env
}
