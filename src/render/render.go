// Package render builds the prompt engine and SVG-encoding options every
// caller that turns a config into a rendered image needs, regardless of
// where that caller runs: the CLI's `config export image`/`config export
// data` commands (cli/config_export_image.go, cli/config_export_data.go) and
// the js/wasm entrypoint (wasm/main.go). The wasm entrypoint cannot import
// cli at all - cli wires up cobra, flag variables, and file/URL I/O, none of
// which exist in a browser - so the render setup and SVG-option resolution
// those commands used to inline had to move somewhere both sides can reach.
// This package is that shared middle layer: everything here is pure engine
// setup and option resolution, with no cobra, no flag parsing, and no file
// writing - those stay in cli, and the wasm entrypoint supplies its own
// equivalents (JS function arguments instead of flags, a returned string
// instead of a written file).
package render

import (
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/prompt"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/svg"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

// rpromptBreathingRoom is how many cells a render leaves between the prompt and an rprompt,
// against prompt.DefaultRPromptBreathingRoom's 30 for an interactive shell. That margin exists so
// a command being typed does not run into the right-aligned block; nothing is ever typed into an
// exported image, so holding 30 cells back only cost the export a block it had room to draw. The
// default config is the case in point: its prompt is 67 cells and its rprompt 24, so at the
// export's own default 120 columns it came up exactly one cell short and the rprompt vanished.
const rpromptBreathingRoom = 20

// Config builds the runtime environment and prompt engine shared by every
// caller that needs to render cfg for its own segment/CapturedRuns output:
// assemble runtime.Flags from cfg, initialize the environment, seed the
// template cache, clear the config fields a render never wants to print
// (ConsoleTitleTemplate/PWD/ShellIntegration), initialize the terminal for
// the GENERIC shell and cfg's own colors, build the Engine, and render the
// primary prompt. This is the ~30-line sequence the CLI's image command,
// single-config data command, and recordThemeSanitized's per-theme loop each
// used to inline before it was first extracted to cli/config_export_data.go
// as renderConfigForExport, and the js/wasm entrypoint (wasm/main.go) needs
// the exact same sequence a second time - a browser has no config file and
// no cobra flags, but it still has to build the same runtime.Flags/
// runtime.Terminal/prompt.Engine trio, which is why this now lives where
// both cli and wasm can import it instead of staying cli-only.
//
// applyData, when non-nil, runs against the freshly built flags before
// env.Init: this is how the CLI image command's own --data replay hooks in,
// and how the wasm entrypoint applies its dataJSON argument and sets
// runtime.Flags.DataOnly. The CLI data command and recordThemeSanitized have
// no --data flag of their own and pass nil.
//
// resetTemplateCache selects a template.ResetCache() call before
// template.Init. The single-config paths (image, data) rely on the
// process-wide template cache starting fresh for their one render; the wasm
// entrypoint's render() function is called repeatedly against the same
// long-lived wasm instance - once per keystroke in a studio preview, say -
// and must reset between calls for the exact same reason
// recordThemeSanitized resets between per-theme renders in the same CLI
// process: otherwise one render's Var/Maps would leak into the next (the
// same isolation prompt/golden_test.go's renderTheme relies on between
// themes).
//
// Ordering matters and is preserved exactly as the original call sites had
// it: template.Init must run before the terminal colors are resolved, since
// cfg.TerminalBackground.ResolveTemplate() and cfg.MakeColors(env) can both
// reference template state. terminal.CaptureRuns = true is NOT set here - it
// must be set before eng.Primary() runs, but is otherwise unrelated to this
// shared setup, so every caller (the CLI image command, the wasm
// entrypoint) sets it at its own call site instead.
func Config(cfg *config.Config, terminalWidth int, resetTemplateCache bool, applyData func(*runtime.Flags) error) (*prompt.Engine, error) {
	flags := &runtime.Flags{
		ConfigPath:    cfg.Source,
		Shell:         shell.GENERIC,
		TerminalWidth: terminalWidth,
	}

	if applyData != nil {
		if err := applyData(flags); err != nil {
			return nil, err
		}
	}

	env := &runtime.Terminal{}
	env.Init(flags)

	if resetTemplateCache {
		template.ResetCache()
	}

	template.Init(env, cfg.Var, cfg.Maps)

	// set sane defaults for things we don't print/need while rendering for export
	cfg.ConsoleTitleTemplate = ""
	cfg.PWD = ""
	cfg.ShellIntegration = false

	terminal.Init(shell.GENERIC)
	terminal.BackgroundColor = cfg.TerminalBackground.ResolveTemplate()
	terminal.Colors = cfg.MakeColors(env)

	eng := &prompt.Engine{
		Config:               cfg,
		Env:                  env,
		RPromptBreathingRoom: rpromptBreathingRoom,
	}

	eng.Primary()

	return eng, nil
}

// FontMetrics groups the font-shape values every SVG-rendering caller can
// quote, each a multiple of font-size (see SVGOptions). They travel together
// because they describe one font: mixing a CellWidth from one with a
// LineHeight from another lays the grid out for a font that doesn't exist.
// The CLI's image command fills this from its --cell-width/--line-height/
// --fill-ascent/--fill-descent flags; the wasm entrypoint fills it from its
// JS optionsObject argument's equivalent fields.
type FontMetrics struct {
	CellWidth   float64
	LineHeight  float64
	FillAscent  float64
	FillDescent float64
}

// SVGOptions resolves the two colors a Run stream can't carry itself — the
// real terminal background and the resolved accent — from the same
// terminal.Colors/terminal.BackgroundColor the ANSI writer already uses, so
// the SVG's "default"/"accent" colors match what a real terminal would show.
// fontFamily is the raw font-family value a caller supplies (the CLI's
// --font-family flag, the wasm entrypoint's optionsObject); an empty string
// leaves svg.Options.FontFamily unset, so Encode falls back to its own
// default Nerd Font stack. columns is the terminal width the prompt engine
// rendered at (see runtime.Flags.TerminalWidth in Config above): the
// exported canvas must use the same value, or right-aligned blocks/rprompts
// computed for that width would land in a canvas shrink-wrapped (or fixed)
// to a different one.
//
// metrics carries the four font-metric values, all expressed as a multiple
// of font-size (e.g. 0.6021, not a pixel count) — the same units the
// measurements behind svg.Options' own defaults are quoted in, and the
// natural unit for a caller quoting a font's own ratios. svg.Options' fields
// themselves are absolute (the same unit FontSize is in, not a ratio of it —
// see their own doc comments), so a nonzero metric value is scaled by
// svg.DefaultFontSize here before being stored. Zero leaves the corresponding
// svg.Options field unset, so Encode falls back to its own Hack-derived
// default, correct for the bundled font stack but not for a caller (e.g. the
// website, rendering in Victor Mono) using a different font.
//
// This resolution deliberately stays in this package rather than becoming a
// constructor on svg.Options itself: svg.Options' own doc comment explains
// that taking colors as explicit fields (instead of Encode reaching into
// package state) is what keeps color resolution unit-testable, and
// terminal.Colors/terminal.BackgroundColor are exactly the kind of live
// package state that comment is warning against baking into svg. Putting the
// resolution here, one level up, keeps svg pure while still giving both the
// CLI and the wasm entrypoint one shared place to build a populated
// svg.Options from - the plain struct remains just as usable directly for a
// caller (like svg's own tests) that wants to skip this resolution entirely.
func SVGOptions(fontFamily string, columns int, metrics FontMetrics) svg.Options {
	var opts svg.Options

	opts.FontFamily = fontFamily
	opts.Columns = columns

	if metrics.CellWidth != 0 {
		opts.CellWidth = metrics.CellWidth * svg.DefaultFontSize
	}

	if metrics.LineHeight != 0 {
		opts.LineHeight = metrics.LineHeight * svg.DefaultFontSize
	}

	if metrics.FillAscent != 0 {
		opts.FillAscent = metrics.FillAscent * svg.DefaultFontSize
	}

	if metrics.FillDescent != 0 {
		opts.FillDescent = metrics.FillDescent * svg.DefaultFontSize
	}

	if accentFg, ok := svg.ParseTrueColorSGR(terminal.Colors.ToAnsi(color.Accent, false)); ok {
		opts.AccentForeground = accentFg
	}

	if accentBg, ok := svg.ParseTrueColorSGR(terminal.Colors.ToAnsi(color.Accent, true)); ok {
		opts.AccentBackground = accentBg
	}

	if bg, ok := svg.ResolveStaticRGB(terminal.BackgroundColor, true, &opts); ok {
		opts.TerminalBackground = bg
	}

	return opts
}

// SVG renders eng's already-captured Run stream (terminal.CaptureRuns must
// have been true for the eng.Primary() call inside Config that produced eng)
// into a self-contained SVG document string. This is exportSVG's own glue
// before it wrote anything to disk - CapturedRuns()/CursorAnchor() feeding
// svg.Encode - split out so the wasm entrypoint can reuse the exact same
// glue instead of reimplementing it; cli's exportSVG (config_export_svg.go)
// now calls this and only keeps the CLI-specific tail: resolving the output
// path and writing the file.
//
//nolint:gocritic
func SVG(eng *prompt.Engine, opts svg.Options) string {
	rows := eng.CapturedRuns()

	if row, run, ok := eng.CursorAnchor(); ok {
		opts.Cursor = &svg.Cursor{Row: row, Run: run}
	}

	return svg.Encode(rows, opts)
}
