package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/segments"

	"github.com/spf13/cobra"
)

func codexStatusDataSource(cmd *cobra.Command) ([]byte, error) {
	options, err := codexLocalStatusOptionsFromCommand(cmd)
	if err != nil {
		return nil, err
	}

	data, err := segments.CodexStatusFromLocalSessions(options)
	if err != nil {
		return nil, err
	}

	return json.Marshal(data)
}

func codexLocalStatusOptionsFromCommand(cmd *cobra.Command) (segments.CodexLocalStatusOptions, error) {
	sessionID, err := cmd.Flags().GetString("session")
	if err != nil {
		return segments.CodexLocalStatusOptions{}, err
	}

	codexHome, err := cmd.Flags().GetString("codex-home")
	if err != nil {
		return segments.CodexLocalStatusOptions{}, err
	}

	sessionRoot, err := cmd.Flags().GetString("session-root")
	if err != nil {
		return segments.CodexLocalStatusOptions{}, err
	}

	codexHomeEnv := os.Getenv("CODEX_HOME")
	userHome := ""
	if codexHome == "" && codexHomeEnv == "" && sessionRoot == "" {
		userHome, err = os.UserHomeDir()
		if err != nil {
			return segments.CodexLocalStatusOptions{}, fmt.Errorf("failed to resolve home directory for codex status discovery: %w", err)
		}
	}

	return segments.ResolveCodexLocalStatusOptions(segments.CodexLocalStatusOptions{
		CodexHome:   codexHome,
		SessionRoot: sessionRoot,
		SessionID:   sessionID,
	}, codexHomeEnv, userHome), nil
}
