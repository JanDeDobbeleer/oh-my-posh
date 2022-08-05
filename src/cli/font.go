package cli

import (
	"fmt"
	"oh-my-posh/font"

	"github.com/spf13/cobra"
)

var (
	// fontCmd can work with fonts
	fontCmd = &cobra.Command{
		Use:   "font [install|configure]",
		Short: "Manage fonts",
		Long: `Manage fonts.

This command is used to install fonts and configure the font in your terminal.

  - install: oh-my-posh font install 3270`,
		ValidArgs: []string{
			"install",
			"configure",
		},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = cmd.Help()
				return
			}
			switch args[0] {
			case "install":
				var fontName string
				if len(args) > 1 {
					fontName = args[1]
				}
				font.Run(fontName)
				return
			case "configure":
				fmt.Println("not implemented")
			default:
				_ = cmd.Help()
			}
		},
	}
)

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(fontCmd)
}
