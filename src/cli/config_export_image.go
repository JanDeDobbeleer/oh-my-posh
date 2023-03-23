package cli

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/ansi"
	"github.com/jandedobbeleer/oh-my-posh/src/engine"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/spf13/cobra"
)

var (
	author        string
	cursorPadding int
	rPromptOffset int
	bgColor       string
	outputImage   string
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
	Run: func(cmd *cobra.Command, args []string) {
		env := &platform.Shell{
			CmdFlags: &platform.Flags{
				Config:  config,
				Shell:   shell.GENERIC,
				Version: cliVersion,
			},
		}
		env.Init()
		defer env.Close()
		cfg := engine.LoadConfig(env)

		// set dsane defaults for things we don't print
		cfg.ConsoleTitleTemplate = ""
		cfg.PWD = ""

		writerColors := cfg.MakeColors()
		writer := &ansi.Writer{
			TerminalBackground: shell.ConsoleBackgroundColor(env, cfg.TerminalBackground),
			AnsiColors:         writerColors,
			TrueColor:          env.CmdFlags.TrueColor,
		}
		writer.Init(shell.GENERIC)
		eng := &engine.Engine{
			Config: cfg,
			Env:    env,
			Writer: writer,
		}

		prompt := eng.PrintPrimary()

		imageCreator := &engine.ImageRenderer{
			AnsiString:    prompt,
			Author:        author,
			CursorPadding: cursorPadding,
			RPromptOffset: rPromptOffset,
			BgColor:       bgColor,
			Ansi:          writer,
		}
		if outputImage != "" {
			imageCreator.Path = cleanOutputPath(outputImage, env)
		}
		imageCreator.Init(env.Flags().Config)
		err := imageCreator.SavePNG()

		if err != nil {
			fmt.Print(err.Error())
		}
	},
}

func init() { //nolint:gochecknoinits
	imageCmd.Flags().StringVar(&author, "author", "", "config author")
	imageCmd.Flags().StringVar(&bgColor, "background-color", "", "image background color")
	imageCmd.Flags().IntVar(&cursorPadding, "cursor-padding", 0, "prompt cursor padding")
	imageCmd.Flags().IntVar(&rPromptOffset, "rprompt-offset", 0, "right prompt offset")
	imageCmd.Flags().StringVarP(&outputImage, "output", "o", "", "image file (.png) to export to")
	exportCmd.AddCommand(imageCmd)
}
