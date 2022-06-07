package cli

import (
	"fmt"
	"oh-my-posh/font"
	"os"

	"github.com/spf13/cobra"
)

var (
	fontName string

	// fontCmd can work with fonts
	fontCmd = &cobra.Command{
		Use:   "font [install|configure]",
		Short: "Manage fonts",
		Long: `Manage fonts.

This command is used to install fonts and configure the font in your terminal.

  - install: oh-my-posh font install https://github.com/ryanoasis/nerd-fonts/releases/download/v2.1.0/3270.zip`,
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
				if len(fontName) == 0 {
					font.Run()
					return
				}
				err := font.Install(fontName)
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}
				return
			case "configure":
				fmt.Println("not implemented")
			default:
				_ = cmd.Help()
			}
		},
	}
)

func init() { // nolint:gochecknoinits
	fontCmd.Flags().StringVarP(&fontName, "font", "f", "", "the font name to install")
	rootCmd.AddCommand(fontCmd)
}
