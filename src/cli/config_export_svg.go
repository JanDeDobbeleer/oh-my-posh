package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/prompt"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/render"
)

// exportSVG renders eng's already-captured Run stream (terminal.CaptureRuns
// must have been true for the eng.Primary() call that preceded this) to a
// non-rasterized SVG file. columns is the --terminal-width the prompt was
// rendered with (see render.SVGOptions). cellWidth/lineHeight are the
// --cell-width/--line-height flags' raw values (see render.SVGOptions).
//
// The option-building and the CapturedRuns/CursorAnchor/svg.Encode glue both
// moved to the render package (render.SVGOptions/render.SVG) so the js/wasm
// entrypoint can reuse them without linking cli; this function keeps only
// what stays genuinely CLI-shaped - resolving the output path and writing
// the file.
func exportSVG(eng *prompt.Engine, cfg *config.Config, output, fontFamily string, columns int, metrics render.FontMetrics) error {
	opts := render.SVGOptions(fontFamily, columns, metrics)

	doc := render.SVG(eng, opts)

	path := svgOutputPath(cfg.Source, output)

	return os.WriteFile(path, []byte(doc), 0o644) //nolint:gosec
}

// svgOutputPath derives the export path: an explicit --output wins (with its
// extension swapped to .svg, so -o mytheme.png still produces mytheme.svg),
// otherwise the config's own basename does, falling back to "prompt.svg"
// when neither yields a usable name.
func svgOutputPath(configPath, output string) string {
	const svgExt = ".svg"

	if output != "" {
		path := cleanOutputPath(output)
		return strings.TrimSuffix(path, filepath.Ext(path)) + svgExt
	}

	base := filepath.Base(configPath)

	match := regex.FindNamedRegexMatch(`(\.?)(?P<STR>.*)\.(json|yaml|yml|toml|jsonc)`, base)

	// TrimSuffix, not TrimRight: TrimRight takes a cutset, so `.omp` there
	// strips any trailing run of '.', 'o', 'm' and 'p' rather than the
	// suffix. The retired PNG renderer carried that bug (image.go's
	// setOutputPath) and it came along with the port — demo.json exported as
	// de.svg, promo.json as pr.svg.
	name := strings.TrimSuffix(match["STR"], ".omp")

	if name == "" {
		name = "prompt"
	}

	return name + svgExt
}
