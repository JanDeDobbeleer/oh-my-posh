package main

import "strings"

type command struct {
	props Properties
	env   Environment

	Output string
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
			output := c.env.runShellCommand(shell, strings.TrimSpace(cmd))
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
			output += c.env.runShellCommand(shell, strings.TrimSpace(cmd))
		}
		c.Output = output
		return c.Output != ""
	}
	c.Output = c.env.runShellCommand(shell, strings.TrimSpace(command))
	return c.Output != ""
}

func (c *command) string() string {
	segmentTemplate := c.props.getString(SegmentTemplate, "{{.Output}}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  c,
		Env:      c.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (c *command) init(props Properties, env Environment) {
	c.props = props
	c.env = env
}
