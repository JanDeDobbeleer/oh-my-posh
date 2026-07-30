package cli

import (
	"github.com/jandedobbeleer/oh-my-posh/src/cmdtree"
)

func NoArgsOrOneValidArg(cmd *cmdtree.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}

	if err := cmdtree.ExactArgs(1)(cmd, args); err != nil {
		return err
	}

	return cmdtree.OnlyValidArgs(cmd, args)
}
