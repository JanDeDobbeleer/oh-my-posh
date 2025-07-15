package cli

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/cli/image"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/prompt"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/spf13/cobra"
)

var (
	author string
	// cursorPadding int
	// rPromptOffset int
	bgColor           string
	outputImage       string
	colorSettingsFile string
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
- color-settings: JSON file with color overrides for ANSI color names

Example usage:

> oh-my-posh config export image --config ~/myconfig.omp.json

Exports the config to an image file called myconfig.png in the current working directory.

> oh-my-posh config export image --config ~/myconfig.omp.json --output ~/mytheme.png

Exports the config to an image file ~/mytheme.png.

> oh-my-posh config export image --config ~/myconfig.omp.json --settings ~/.image.settings.json

Exports the config to an image file using customized output settings.`,
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		cfg, _ := config.Load(configFlag, shell.GENERIC, false)

		flags := &runtime.Flags{
			Config:        cfg.Source,
			Shell:         shell.GENERIC,
			TerminalWidth: 150,
		}

		env := &runtime.Terminal{}
		env.Init(flags)

		template.Init(env, cfg.Var, cfg.Maps)

		defer func() {
			template.SaveCache()
			env.Close()
		}()

		// set sane defaults for things we don't print
		cfg.ConsoleTitleTemplate = ""
		cfg.PWD = ""

		terminal.Init(shell.GENERIC)
		terminal.BackgroundColor = cfg.TerminalBackground.ResolveTemplate()
		terminal.Colors = cfg.MakeColors(env)

		eng := &prompt.Engine{
			Config: cfg,
			Env:    env,
		}

		settings, err := image.LoadSettings(colorSettingsFile)
		if err != nil {
			settings = &image.Settings{
				Colors:          image.NewColors(),
				Author:          author,
				BackgroundColor: bgColor,
			}
		}

		if settings.Colors == nil {
			settings.Colors = image.NewColors()
		}

		primaryPrompt := eng.Primary()

		imageCreator := &image.Renderer{
			AnsiString: primaryPrompt,
			Settings:   *settings,
		}

		if outputImage != "" {
			imageCreator.Path = cleanOutputPath(outputImage)
		}

		err = imageCreator.Init(env)
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
	imageCmd.Flags().StringVar(&colorSettingsFile, "settings", "", "color settings file to override ANSI color codes and metadata")

	// deprecated flags
	_ = imageCmd.Flags().MarkHidden("author")
	_ = imageCmd.Flags().MarkHidden("background-color")

	exportCmd.AddCommand(imageCmd)
}
