package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Cmd struct {
	props properties.Properties
	env   platform.Environment

	Output string
}

const (
	// ExecutableShell to execute command in
	ExecutableShell properties.Property = "shell"
	// Command to execute
	Command properties.Property = "command"
	// Command to execute
	Script properties.Property = "script"
	// Interpret execution, or not
	Interpret properties.Property = "interpret"
)

func (c *Cmd) Template() string {
	return " {{ .Output }} "
}

func (c *Cmd) Enabled() bool {
	shell := c.props.GetString(ExecutableShell, "bash")
	if !c.env.HasCommand(shell) {
		return false
	}

	command := c.props.GetString(Command, "")
	if len(command) != 0 {
		return c.runCommand(shell, command)
	}

	script := c.props.GetString(Script, "")
	if len(script) != 0 {
		return c.runScript(shell, script)
	}

	return false
}

func (c *Cmd) runCommand(shell, command string) bool {
	interpret := c.props.GetBool(Interpret, true)

	if !interpret {
		c.Output = c.env.RunShellCommand(shell, command)
		return len(c.Output) != 0
	}

	if strings.Contains(command, "||") {
		commands := strings.Split(command, "||")
		for _, cmd := range commands {
			output := c.env.RunShellCommand(shell, strings.TrimSpace(cmd))
			if len(output) != 0 {
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
		return len(c.Output) != 0
	}

	c.Output = c.env.RunShellCommand(shell, strings.TrimSpace(command))
	return len(c.Output) != 0
}

func (c *Cmd) runScript(shell, script string) bool {
	c.Output = c.env.RunShellCommand(shell, script)
	return len(c.Output) != 0
}

func (c *Cmd) Init(props properties.Properties, env platform.Environment) {
	c.props = props
	c.env = env
}
