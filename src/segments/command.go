package segments

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"strings"
)

type Cmd struct {
	props properties.Properties
	env   environment.Environment

	Output string
}

const (
	// ExecutableShell to execute command in
	ExecutableShell properties.Property = "shell"
	// Command to execute
	Command properties.Property = "command"
)

func (c *Cmd) Template() string {
	return "{{ .Output }}"
}

func (c *Cmd) Enabled() bool {
	shell := c.props.GetString(ExecutableShell, "bash")
	if !c.env.HasCommand(shell) {
		return false
	}
	command := c.props.GetString(Command, "echo no command specified")
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

func (c *Cmd) Init(props properties.Properties, env environment.Environment) {
	c.props = props
	c.env = env
}
