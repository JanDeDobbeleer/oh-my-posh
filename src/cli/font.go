package cli

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/font"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/dsc"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/jandedobbeleer/oh-my-posh/src/cmdtree"
)

var (
	zipFolder string

	fontCmd = &cmdtree.Command{
		Use:   "font",
		Short: "Manage fonts",
		Long: `Manage fonts.

List the available Nerd Fonts and install one:

  oh-my-posh font list
  oh-my-posh font install Meslo`,
	}

	fontListCmd = &cmdtree.Command{
		Use:   "list",
		Short: "List the available Nerd Fonts",
		Long: `List the available Nerd Fonts.

Prints one font name per line, so it can be searched or piped:

  oh-my-posh font list | grep -i mono`,
		Args: cmdtree.NoArgs,
		Run: func(_ *cmdtree.Command, _ []string) {
			fonts, err := font.List()
			if err != nil {
				log.Error(err)
				exitcode = 70

				return
			}

			for _, f := range fonts {
				fmt.Println(f.Name)
			}
		},
	}

	fontInstallCmd = &cmdtree.Command{
		Use:   "install <font>",
		Short: "Install a Nerd Font",
		Long: `Install a Nerd Font.

Takes a font name from ` + "`oh-my-posh font list`" + `, a URL, or the path to a local zip file:

  oh-my-posh font install Meslo
  oh-my-posh font install https://example.com/font.zip
  oh-my-posh font install ./CascadiaCode.zip`,
		Args: cmdtree.ExactArgs(1),
		Run: func(_ *cmdtree.Command, args []string) {
			env := &runtime.Terminal{}
			env.Init(&runtime.Flags{})

			sh := env.Shell()

			cache.Init(sh, cache.Persist)

			defer cache.Close()

			terminal.Init(sh)

			cfg := config.Get(configFlag, false)
			cfg.TerminalFeatures.Apply()

			if zipFolder != "" && !strings.HasPrefix(zipFolder, "/") {
				zipFolder += "/"
			}

			fontName, err := font.Install(args[0], zipFolder)
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
			fontDSC.Load()
			fontDSC.Add(fontName)
			fontDSC.Save()
		},
	}
)

func init() {
	fontInstallCmd.Flags().StringVar(&zipFolder, "zip-folder", "", "the folder inside the zip file to install fonts from")

	fontCmd.AddCommand(fontListCmd)
	fontCmd.AddCommand(fontInstallCmd)
	fontCmd.AddCommand(dsc.Command(font.DSC()))

	RootCmd.AddCommand(fontCmd)
}
