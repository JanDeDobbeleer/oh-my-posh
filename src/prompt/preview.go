package prompt

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

func (e *Engine) Preview() string {
	builder := text.NewBuilder()

	printPrompt := func(title, prompt string) {
		builder.WriteString(log.Text(fmt.Sprintf("\n%s:\n\n", title)).Bold().Plain().String())
		builder.WriteString(prompt)
		builder.WriteString("\n")
	}

	printPrompt("Primary", e.Primary())

	right := e.RPrompt()
	if len(right) > 0 {
		printPrompt("Right", right)
	}

	if e.Config.SecondaryPrompt != nil {
		printPrompt("Secondary", e.ExtraPrompt(Secondary))
	}

	if e.Config.TransientPrompt != nil {
		printPrompt("Transient", e.ExtraPrompt(Transient))
	}

	if e.Config.DebugPrompt != nil {
		printPrompt("Debug", e.ExtraPrompt(Debug))
	}

	if e.Config.ValidLine != nil {
		printPrompt("Valid", e.ExtraPrompt(Valid))
	}

	if e.Config.ErrorLine != nil {
		printPrompt("Error", e.ExtraPrompt(Error))
	}

	builder.WriteString("\n")

	return builder.String()
}
