package cli

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cli/font"
	"github.com/jandedobbeleer/oh-my-posh/src/dsc"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/spf13/cobra"
)

var (
	zipFolder string

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

				flags := &runtime.Flags{
					SaveCache: true,
				}

				env := &runtime.Terminal{}
				env.Init(flags)
				defer env.Close()

				terminal.Init(env.Shell())

				if !strings.HasPrefix(zipFolder, "/") {
					zipFolder += "/"
				}

				fontName, err := font.Run(fontName, env.Cache(), zipFolder)
				if err != nil {
					log.Error(err)
					exitcode = 70
					return
				}

				if env.Root() {
					// do not update the DSC cache if we are running as root
					return
				}

				fontDSC := font.DSC()
				fontDSC.Load(env.Cache())
				fontDSC.Add(fontName)
				fontDSC.Save()

				return
			case "configure":
				fmt.Println("not implemented")
			default:
				_ = cmd.Help()
			}
		},
	}
)

func init() {
	fontCmd.Flags().StringVar(&zipFolder, "zip-folder", "", "the folder inside the zip file to install fonts from")
	fontCmd.AddCommand(dsc.Command(font.DSC()))
	RootCmd.AddCommand(fontCmd)
}
