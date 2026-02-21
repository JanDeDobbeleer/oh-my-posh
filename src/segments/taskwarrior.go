package segments

import (
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

// Taskwarrior option constants
const (
	TaskwarriorCommand options.Option = "command"
	TaskwarriorTags    options.Option = "tags"
	TaskwarriorCtxCmd  options.Option = "context_command"
)

// Taskwarrior displays task counts and context from Taskwarrior.
// The Tags field is a map from capitalized tag name to task count, populated
// from the "tags" configuration option. Each entry in the config map has the
// tag name as key and a Taskwarrior filter expression as value (without the
// trailing "count" word, which is appended automatically).
type Taskwarrior struct {
	Base

	// Tags holds a count per tag name (first letter uppercased).
	Tags    map[string]int
	Context string
}

func (t *Taskwarrior) Template() string {
	return " \uf4a0 {{ range $k, $v := .Tags }}{{ $k }}:{{ $v }} {{ end }}"
}

func (t *Taskwarrior) Enabled() bool {
	cmd := t.options.String(TaskwarriorCommand, "task")

	if !t.env.HasCommand(cmd) {
		return false
	}

	defaultTags := map[string]string{
		"due":       "+PENDING due.before:tomorrow",
		"scheduled": "+PENDING scheduled.before:tomorrow",
		"waiting":   "+WAITING",
	}

	configuredTags := t.options.KeyValueMap(TaskwarriorTags, defaultTags)

	t.Tags = make(map[string]int, len(configuredTags))

	for name, filter := range configuredTags {
		count := t.queryCount(cmd, filter)
		key := strings.ToUpper(name[:1]) + name[1:]
		t.Tags[key] = count
	}

	contextCmd := t.options.String(TaskwarriorCtxCmd, "")
	if contextCmd != "" {
		t.Context = t.queryContext(cmd, contextCmd)
	}

	return true
}

func (t *Taskwarrior) queryCount(cmd, filter string) int {
	args := strings.Fields(filter)
	args = append(args, "count")

	output, err := t.env.RunCommand(cmd, args...)
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

func (t *Taskwarrior) queryContext(cmd, args string) string {
	output, err := t.env.RunCommand(cmd, strings.Fields(args)...)
	if err != nil {
		log.Error(err)
		return ""
	}

	return strings.TrimSpace(output)
}
