package cli

import (
	"fmt"
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/render"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/spf13/cobra"
)

var (
	outputImage        string
	svgFontFamily      string
	imageTerminalWidth int
	svgCellWidth       float64
	svgLineHeight      float64
	svgFillAscent      float64
	svgFillDescent     float64
)

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Export your config to an SVG image",
	Long: `Export your config to an SVG image.

The image renders straight from the prompt's own Run stream, so it stays crisp at any zoom level
and, unlike a rasterized image, faithfully reproduces every color and style the prompt itself can
render.

You can tweak the output by using additional flags:

- terminal-width: number of columns the image canvas is rendered at; this
  is also the width the prompt is rendered with, so right-aligned blocks
  and rprompts line up with the image the same way they would in a real
  terminal of that width (default 120). Lines that still don't fit either
  collapse their alignment padding or wrap onto the next row
- font-family: CSS font-family value the svg renders text with; defaults to
  a Nerd Font stack
- cell-width: horizontal advance of one monospace cell, as a multiple of
  font-size (e.g. 0.6021, not a pixel count); defaults to Hack Nerd Font's
  own advance ratio, matching the default font-family stack. Set this
  explicitly when font-family names a font with a different advance ratio
- line-height: vertical advance of one row, as a multiple of font-size;
  defaults to Hack Nerd Font's own line height. Set this explicitly
  alongside font-family/cell-width for the same reason
- fill-ascent, fill-descent: how far a segment's background reaches above
  and below the text baseline, as multiples of font-size. Together they are
  the box a powerline separator glyph inks, so the background ends exactly
  where the separator does instead of standing taller than it. Both default
  to Hack Nerd Font's own U+E0B0 ink box; set them alongside font-family for
  the same reason as cell-width and line-height

Example usage:

> oh-my-posh config export image --config ~/myconfig.omp.json

Exports the config to an image file called myconfig.svg in the current working directory.

> oh-my-posh config export image --config ~/myconfig.omp.json --output ~/mytheme.svg

Exports the config to an image file ~/mytheme.svg.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		cache.Init(os.Getenv("POSH_SHELL"))

		setConfigFlag()

		cfg := config.Load(configFlag)

		// CaptureRuns must be set before render.Config's eng.Primary() call so
		// exportSVG can read the structured Run stream back via eng.CapturedRuns(); it
		// stays here rather than in the helper because it's specific to this command,
		// not part of the shared render setup.
		terminal.CaptureRuns = true

		eng, err := render.Config(cfg, imageTerminalWidth, false, func(flags *runtime.Flags) error {
			return applyDataFile(flags, cmd.Flags().Changed)
		})
		if err != nil {
			exitcode = 666
			fmt.Println(err.Error())
			return
		}

		defer func() {
			template.SaveCache()
			cache.Close()
		}()

		metrics := render.FontMetrics{
			CellWidth:   svgCellWidth,
			LineHeight:  svgLineHeight,
			FillAscent:  svgFillAscent,
			FillDescent: svgFillDescent,
		}

		if err := exportSVG(eng, cfg, outputImage, svgFontFamily, imageTerminalWidth, metrics); err != nil {
			exitcode = 666
			fmt.Println(err.Error())
		}
	},
}

func init() {
	imageCmd.Flags().StringVarP(&outputImage, "output", "o", "", "image file (.svg) to export to")
	imageCmd.Flags().StringVar(&svgFontFamily, "font-family", "", "CSS font-family for the exported svg")
	imageCmd.Flags().StringVar(&dataPath, "data", "", "path to a template data file (json/yaml/toml) to render with")
	imageCmd.Flags().IntVar(&imageTerminalWidth, "terminal-width", 120, "number of columns to render the prompt and image at")
	imageCmd.Flags().Float64Var(&svgCellWidth, "cell-width", 0, "horizontal advance of one cell, as a multiple of font-size (defaults to Hack Nerd Font's own ratio)")
	imageCmd.Flags().Float64Var(&svgLineHeight, "line-height", 0, "vertical advance of one row, as a multiple of font-size (defaults to Hack Nerd Font's own ratio)")
	const fillHelp = "how far a segment background reaches %s the baseline, as a multiple of font-size " +
		"(defaults to Hack Nerd Font's own U+E0B0 ink box)"

	imageCmd.Flags().Float64Var(&svgFillAscent, "fill-ascent", 0, fmt.Sprintf(fillHelp, "above"))
	imageCmd.Flags().Float64Var(&svgFillDescent, "fill-descent", 0, fmt.Sprintf(fillHelp, "below"))

	exportCmd.AddCommand(imageCmd)
}

// setConfigFlag resolves which config the export commands act on: the one --config names, else
// the one this shell's session cached when it rendered its prompt.
//
// Neither is an error. An empty configFlag is what config.Load already turns into the built-in
// default (see config/read), which is exactly the prompt `oh-my-posh print primary` renders for
// someone who has not configured anything - so refusing to export it, as this used to, made the
// export commands the only ones that could not act on the default config.
func setConfigFlag() {
	if configFlag != "" {
		return
	}

	if configPath, OK := cache.Get[string](cache.Session, config.SourceKey); OK {
		configFlag = configPath
	}
}
