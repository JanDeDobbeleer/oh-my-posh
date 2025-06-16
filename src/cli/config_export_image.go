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
	bgColor         string
	outputImage     string
	fontRegularPath string // For --font and --font-regular
	fontBoldPath    string
	fontItalicPath  string
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
- font: the path to the regular .ttf or .otf font file (NerdFont recommended)
- font-regular: alias for --font
- font-bold: the path to the bold .ttf or .otf font file (NerdFont recommended)
- font-italic: the path to the italic .ttf or .otf font file (NerdFont recommended)

If --font points to a .ttc or .otc font collection, the first font in the collection will be used as the regular font.
Bold and italic styles must be specified with --font-bold and --font-italic respectively.
Oh My Posh will fall back to a bundled Nerd Font if custom fonts are not specified or cannot be loaded.

Example usage:

> oh-my-posh config export image --config ~/myconfig.omp.json

Exports the config to an image file called myconfig.png in the current working directory.

> oh-my-posh config export image --config ~/myconfig.omp.json --output ~/mytheme.png

Exports the config to an image file ~/mytheme.png.

> oh-my-posh config export image --config ~/myconfig.omp.json --author "John Doe"

Exports the config to an image file using customized output options.

> oh-my-posh config export image --config ~/myconfig.omp.json --font "/path/to/MyNerdFont-Regular.ttf"

Exports the config to an image using a single NerdFont file for regular, bold, and italic styles.

> oh-my-posh config export image --config ~/myconfig.omp.json --font "/path/to/MyNerdFont-Regular.ttf" --font-bold "/path/to/MyNerdFont-Bold.ttf" --font-italic "/path/to/MyNerdFont-Italic.otf"

Exports the config to an image using specific NerdFont files for regular, bold, and italic styles.`,
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		configFile := config.Path(configFlag)
		cfg, _ := config.Load(configFile, shell.GENERIC, false)

		flags := &runtime.Flags{
			Config:        configFile,
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

		primaryPrompt := eng.Primary()

		imageCreator := &image.Renderer{
			AnsiString:      primaryPrompt,
			Author:          author,
			BgColor:         bgColor,
			FontRegularPath: fontRegularPath,
			FontBoldPath:    fontBoldPath,
			FontItalicPath:  fontItalicPath,
		}

		if outputImage != "" {
			imageCreator.Path = cleanOutputPath(outputImage)
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

	imageCmd.Flags().StringVar(&fontRegularPath, "font", "", "path to the regular .ttf or .otf font file (NerdFont recommended)")
	imageCmd.Flags().StringVar(&fontRegularPath, "font-regular", "", "path to the regular .ttf or .otf font file (alias for --font)")
	imageCmd.Flags().StringVar(&fontBoldPath, "font-bold", "", "path to the bold .ttf or .otf font file (NerdFont recommended)")
	imageCmd.Flags().StringVar(&fontItalicPath, "font-italic", "", "path to the italic .ttf or .otf font file (NerdFont recommended)")

	exportCmd.AddCommand(imageCmd)
}
