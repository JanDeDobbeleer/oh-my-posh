package cli

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/font"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"

	"github.com/spf13/cobra"
)

var (
	user bool

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
				env := &platform.Shell{}
				env.Init()
				defer env.Close()

				// Windows users need to specify the --user flag if they want to install the font as user
				// If the user does not specify the --user flag, the font will be installed as a system font
				// and therefore we need to be administrator
				system := env.Root()
				if env.GOOS() == platform.WINDOWS && !user && !system {
					fmt.Println(`
    You need to be administrator to install a font as system font.
    You can either run this command as administrator or specify the --user flag to install the font for your user only:

    oh-my-posh font install --user
    `)
					return
				}

				font.Run(fontName, system)
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
	RootCmd.AddCommand(fontCmd)
	fontCmd.Flags().BoolVar(&user, "user", false, "install font as user")
}
