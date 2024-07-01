package cli

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/engine"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/spf13/cobra"
)

var (
	author string
	// cursorPadding int
	// rPromptOffset int
	bgColor     string
	outputImage string
)

// imageCmd represents the image command
var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Export your config to an image",
	Long: `Export your config to an image.

You can tweak the output by using additional flags:

- author: displays the author below the prompt
- cursor-padding: the padding of the prompt cursor
- rprompt-offset: the offset of the right prompt
- background-color: the background color of the image

Example usage:

> oh-my-posh config export image --config ~/myconfig.omp.json

Exports the config to an image file called myconfig.png in the current working directory.

> oh-my-posh config export image --config ~/myconfig.omp.json --output ~/mytheme.png

Exports the config to an image file ~/mytheme.png.

> oh-my-posh config export image --config ~/myconfig.omp.json --author "John Doe"

Exports the config to an image file using customized output options.`,
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		env := &platform.Shell{
			CmdFlags: &platform.Flags{
				Config:        config,
				Shell:         shell.GENERIC,
				TerminalWidth: 150,
			},
		}

		env.Init()
		defer env.Close()
		cfg := engine.LoadConfig(env)

		// set sane defaults for things we don't print
		cfg.ConsoleTitleTemplate = ""
		cfg.PWD = ""
		cfg.OSC99 = false

		// add variables to the environment
		env.Var = cfg.Var

		terminal.Init(shell.GENERIC)
		terminal.BackgroundColor = shell.ConsoleBackgroundColor(env, cfg.TerminalBackground)
		terminal.AnsiColors = cfg.MakeColors()

		eng := &engine.Engine{
			Config: cfg,
			Env:    env,
		}

		prompt := eng.Primary()

		imageCreator := &engine.ImageRenderer{
			AnsiString: prompt,
			Author:     author,
			BgColor:    bgColor,
		}

		if outputImage != "" {
			imageCreator.Path = cleanOutputPath(outputImage, env)
		}

		err := imageCreator.Init(env)
		if err != nil {
			fmt.Print(err.Error())
			return
		}

		err = imageCreator.SavePNG()
		if err != nil {
			fmt.Print(err.Error())
		}
	},
}

func init() {
	imageCmd.Flags().StringVar(&author, "author", "", "config author")
	imageCmd.Flags().StringVar(&bgColor, "background-color", "", "image background color")
	imageCmd.Flags().StringVarP(&outputImage, "output", "o", "", "image file (.png) to export to")
	exportCmd.AddCommand(imageCmd)
}
