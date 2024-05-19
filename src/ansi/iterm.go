package ansi

import (
	"fmt"
	"strings"
)

type iTermFeature string

const (
	PromptMark iTermFeature = "prompt_mark"
	CurrentDir iTermFeature = "current_dir"
	RemoteHost iTermFeature = "remote_host"
)

type ITermFeatures []iTermFeature

func (w *Writer) RenderItermFeatures(features ITermFeatures, pwd, user, host string) string {
	var result strings.Builder
	for _, feature := range features {
		switch feature {
		case PromptMark:
			result.WriteString(w.iTermPromptMark)
		case CurrentDir:
			result.WriteString(fmt.Sprintf(w.iTermCurrentDir, pwd))
		case RemoteHost:
			result.WriteString(fmt.Sprintf(w.iTermRemoteHost, user, host))
		}
	}

	return result.String()
}
