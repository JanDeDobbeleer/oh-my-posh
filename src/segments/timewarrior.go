package segments

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

const (
	TimewarriorCommand      options.Option = "command"
	TimewarriorTagFormat    options.Option = "tag_format"
	TimewarriorTagSeparator options.Option = "tag_separator"
)

type Timewarrior struct {
	Base

	Start      time.Time
	Annotation string
	Duration   string
	Tags       string
	Active     bool
}

func (t *Timewarrior) Template() string {
	return "{{ if .Active }} \uf252 {{ if .Tags }}{{ .Tags }}{{ else }}\u2205{{ end }}" +
		"{{ if .Annotation }} | {{ .Annotation }}{{ end }}" +
		"{{ else }} Not tracking{{ end }} "
}

func (t *Timewarrior) Enabled() bool {
	command := t.options.String(TimewarriorCommand, "timew")
	if !t.env.HasCommand(command) {
		return false
	}

	status, err := t.env.RunCommand(command, "get", "dom.active")
	if err != nil {
		log.Error(err)
		return false
	}

	alwaysEnabled := t.options.Bool(options.AlwaysEnabled, false)
	if strings.TrimSpace(status) == "0" {
		return alwaysEnabled
	}
	if strings.TrimSpace(status) != "1" {
		return false
	}

	output, err := t.env.RunCommand(command, "get", "dom.active.json")
	if err != nil {
		log.Error(err)
		return false
	}

	var interval struct {
		Start      string   `json:"start"`
		End        string   `json:"end"`
		Annotation string   `json:"annotation"`
		Tags       []string `json:"tags"`
	}
	if err := json.Unmarshal([]byte(output), &interval); err != nil {
		log.Error(err)
		return false
	}

	if interval.End != "" {
		return alwaysEnabled
	}

	const layout = "20060102T150405Z"
	t.Start, err = time.Parse(layout, interval.Start)
	if err != nil {
		log.Error(err)
		return false
	}

	t.Active = true
	t.Annotation = interval.Annotation
	t.Duration = time.Since(t.Start).Truncate(time.Second).String()
	tags := make([]string, len(interval.Tags))
	tagFormat := t.options.String(TimewarriorTagFormat, "#%s")
	for index, tag := range interval.Tags {
		tags[index] = fmt.Sprintf(tagFormat, tag)
	}
	t.Tags = strings.Join(tags, t.options.String(TimewarriorTagSeparator, " "))
	return true
}
