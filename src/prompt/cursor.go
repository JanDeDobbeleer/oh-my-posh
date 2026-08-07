package prompt

import (
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

// CursorStyle renders only the cursor_style escape sequence, skipping the rest of the
// prompt. Enable-PoshVIMode (omp.ps1) calls this once per vi mode and caches the result so
// a mode change inside a PSSession can update the cursor shape with a direct Console.Write
// instead of a full PSReadLine repaint - see writePrimaryPromptInternal in primary.go for
// the same rendering used inline in the primary prompt.
func (e *Engine) CursorStyle() string {
	if len(e.Config.CursorStyle) == 0 {
		return ""
	}

	style, err := template.RenderTrusted(e.Config.CursorStyle, nil)
	if err != nil {
		return ""
	}

	return terminal.SetCursorStyle(style)
}
