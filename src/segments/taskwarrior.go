package segments

import (
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

// Taskwarrior option constants
const (
	TaskwarriorCommand  options.Option = "command"
	TaskwarriorDueCmd   options.Option = "due_command"
	TaskwarriorSchedCmd options.Option = "scheduled_command"
	TaskwarriorWaitCmd  options.Option = "waiting_command"
	TaskwarriorCtxCmd   options.Option = "context_command"
)

type Taskwarrior struct {
	Base

	Due       int
	Scheduled int
	Waiting   int
	Context   string
}

func (t *Taskwarrior) Template() string {
	return " \uf4a0 {{.Due}} "
}

func (t *Taskwarrior) Enabled() bool {
	cmd := t.options.String(TaskwarriorCommand, "task")

	if !t.env.HasCommand(cmd) {
		return false
	}

	t.Due = t.getCount(cmd, TaskwarriorDueCmd, "+PENDING due.before:tomorrow count")
	t.Scheduled = t.getCount(cmd, TaskwarriorSchedCmd, "+PENDING scheduled.before:tomorrow count")
	t.Waiting = t.getCount(cmd, TaskwarriorWaitCmd, "+WAITING count")
	t.Context = t.getContext(cmd, TaskwarriorCtxCmd, "_get rc.context")

	return true
}

func (t *Taskwarrior) getCount(cmd string, opt options.Option, defaultArgs string) int {
	args := t.options.String(opt, defaultArgs)
	output, err := t.env.RunCommand(cmd, strings.Fields(args)...)
	if err != nil {
		log.Error(err)
		return 0
	}

	count, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil {
		log.Error(err)
		return 0
	}

	return count
}

func (t *Taskwarrior) getContext(cmd string, opt options.Option, defaultArgs string) string {
	args := t.options.String(opt, defaultArgs)
	output, err := t.env.RunCommand(cmd, strings.Fields(args)...)
	if err != nil {
		log.Error(err)
		return ""
	}

	return strings.TrimSpace(output)
}
