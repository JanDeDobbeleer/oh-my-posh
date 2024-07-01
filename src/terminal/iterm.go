package terminal

import (
	"fmt"
	"slices"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

type iTermFeature string

const (
	PromptMark iTermFeature = "prompt_mark"
	CurrentDir iTermFeature = "current_dir"
	RemoteHost iTermFeature = "remote_host"
)

type ITermFeatures []iTermFeature

func (f ITermFeatures) Contains(feature iTermFeature) bool {
	for _, item := range f {
		if item == feature {
			return true
		}
	}

	return false
}

func RenderItermFeatures(features ITermFeatures, sh, pwd, user, host string) string {
	supportedShells := []string{shell.BASH, shell.ZSH}

	var result strings.Builder
	for _, feature := range features {
		switch feature {
		case PromptMark:
			if !slices.Contains(supportedShells, sh) {
				continue
			}

			result.WriteString(formats.ITermPromptMark)
		case CurrentDir:
			result.WriteString(fmt.Sprintf(formats.ITermCurrentDir, pwd))
		case RemoteHost:
			result.WriteString(fmt.Sprintf(formats.ITermRemoteHost, user, host))
		}
	}

	return result.String()
}
