package terminal

import (
	"fmt"
	"slices"

	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

const (
	PromptMark = "prompt_mark"
	CurrentDir = "current_dir"
	RemoteHost = "remote_host"
)

func HasITermFeatures() bool {
	return SupportsFeature(PromptMark) || SupportsFeature(CurrentDir) || SupportsFeature(RemoteHost)
}

func RenderItermFeatures(sh, pwd, user, host string) string {
	result := text.NewBuilder()

	supportedShells := []string{shell.BASH, shell.ZSH}

	if SupportsFeature(PromptMark) && slices.Contains(supportedShells, sh) {
		result.WriteString(formats.ITermPromptMark)
	}

	if SupportsFeature(CurrentDir) {
		result.WriteString(fmt.Sprintf(formats.ITermCurrentDir, pwd))
	}

	if SupportsFeature(RemoteHost) {
		result.WriteString(fmt.Sprintf(formats.ITermRemoteHost, user, host))
	}

	return result.String()
}
