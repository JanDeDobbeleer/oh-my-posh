package segments

import (
	"strings"

	c "golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

const (
	TaskwarriorCommand  options.Option = "command"
	TaskwarriorCommands options.Option = "commands"
)

type Taskwarrior struct {
	Base

	// Commands is keyed by the capitalized command name.
	Commands map[string]string
}

func (t *Taskwarrior) Template() string {
	return " \uf4a0 {{ range $k, $v := .Commands }}{{ $k }}:{{ $v }} {{ end }}"
}

func (t *Taskwarrior) Enabled() bool {
	cmd := t.options.String(TaskwarriorCommand, "task")

	if !t.env.HasCommand(cmd) {
		return false
	}

	defaultCommands := map[string]string{
		"due":       "+PENDING due.before:tomorrow count",
		"scheduled": "+PENDING scheduled.before:tomorrow count",
		"waiting":   "+WAITING count",
		"context":   "_get rc.context",
	}

	configuredCommands := t.options.KeyValueMap(TaskwarriorCommands, defaultCommands)

	t.Commands = make(map[string]string, len(configuredCommands))

	for name, args := range configuredCommands {
		key := c.Title(language.English).String(name)
		t.Commands[key] = t.runCommand(cmd, args)
	}

	return true
}

func (t *Taskwarrior) runCommand(cmd, args string) string {
	output, err := t.env.RunCommand(cmd, strings.Fields(args)...)
	if err != nil {
		log.Error(err)
		return ""
	}

	return strings.TrimSpace(output)
}
