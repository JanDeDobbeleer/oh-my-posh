//go:build js && wasm

// Package main is the js/wasm entrypoint that lets a browser render a config
// straight to an SVG string, without shelling out to the CLI: the studio
// (see the website-redesign-direction note) hands over a config's own text
// and a recorded --data file's JSON, rather than a file path or a URL,
// because a browser has neither a filesystem to read nor permission to
// fetch a config on the user's behalf. This package is deliberately thin -
// it is an adapter over the render package (src/render) and config's memory
// parsers (config.ParseBytes/config.ParseData), not a second renderer: every
// piece of actual render logic (engine setup, SVG-option resolution, the
// CapturedRuns/CursorAnchor/svg.Encode glue) already lives in src/render,
// shared with the CLI's `config export image`, precisely so this file never
// has a reason to duplicate any of it.
//
// The js && wasm build constraint keeps this package out of every normal
// build: `go build ./...` on any other GOOS/GOARCH simply has no buildable
// files here and skips the directory, exactly like a _test.go-only package
// would.
package main

import (
	"fmt"
	"strings"
	"syscall/js"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/render"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

func main() {
	js.Global().Set("render", js.FuncOf(renderJS))

	// A wasm_exec.js-hosted module's main must never return: the host glue
	// tears down the Go scheduler as soon as it does, and every function
	// registered above (js.FuncOf) stops working on the very next call. This
	// package exists only to expose renderJS, never to run a program of its
	// own, so blocking forever is exactly what it should do.
	select {}
}

// renderJS is the JS-callable render(configText, format, dataJSON,
// optionsObject) function described in the package doc comment. It returns
// {svg: string} on success or {error: string} on failure - a returned value,
// never a panic: an uncaught panic unwinding out of a js.Func call kills the
// whole wasm instance, not just the one call, which would take down every
// future render() call in the same browser tab along with it. The recover
// below is the backstop for that; renderSVG itself is expected to return an
// error instead of panicking for every input this function's own callers
// (JS, so already untrusted) can produce.
func renderJS(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if r := recover(); r != nil {
			result = map[string]any{"error": fmt.Sprintf("panic: %v", r)}
		}
	}()

	doc, err := renderSVG(args)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}

	return map[string]any{"svg": doc}
}

// renderSVG does the actual work: parse the config and optional data from
// memory, render them, and encode the result - all through config's memory
// parsers and the render package, never through config.Load or anything
// else that would touch a file or the network (config.ParseBytes/
// config.ParseData are exactly the memory-only counterparts to config.Load's
// file-and-URL-reading Parse; see their own doc comments for why a config
// parsed this way can never pull in a remote "extends").
func renderSVG(args []js.Value) (string, error) {
	if len(args) < 4 {
		return "", fmt.Errorf("render expects 4 arguments (configText, format, dataJSON, options), got %d", len(args))
	}

	configText := jsString(args[0])
	format := jsString(args[1])
	dataJSON := jsString(args[2])
	options := args[3]

	// An empty config means the built-in default, the same as it does for the CLI: config.Load
	// with no path returns config.Default(). Parsing "" instead would produce an empty Config and
	// render nothing, so a caller that wants to see what oh-my-posh looks like out of the box -
	// the website's own homepage does - has no other way to ask for it.
	var cfg *config.Config

	var err error

	if strings.TrimSpace(configText) == "" {
		cfg = config.Default(nil)
	} else {
		cfg, err = config.ParseBytes(format, []byte(configText))
		if err != nil {
			return "", fmt.Errorf("failed to parse config: %w", err)
		}
	}

	var data *config.Data

	if dataJSON != "" {
		data, err = config.ParseData([]byte(dataJSON))
		if err != nil {
			return "", fmt.Errorf("failed to parse data: %w", err)
		}
	}

	// applyData plays the same role here as the CLI image command's own
	// --data closure (cli/config_export_image.go, applyDataFile): it runs
	// against render.Config's freshly built flags before env.Init.
	applyData := func(flags *runtime.Flags) error {
		// DataOnly makes the recorded data the only source a segment may
		// render from (see runtime.Flags.DataOnly's own doc comment). The
		// CLI's --data-only is an opt-in a user can leave off, in which case
		// an uncovered segment falls through to probing the real machine;
		// here that fallback would mean probing the *browser's* fake
		// environment, which can never be what a caller of this function
		// wants, so this is mandatory rather than a choice exposed to JS.
		flags.DataOnly = true

		if data == nil {
			return nil
		}

		// changed is nil: there is no CLI flag of data's own env keys (pwd,
		// status, ...) to consult for "was this overridden explicitly" from
		// this call, exactly like prompt/golden_test.go's fixture harness -
		// see config.Data.ApplyFlags' own doc comment.
		return data.ApplyFlags(flags, nil)
	}

	// terminal.CaptureRuns must be set before render.Config's own
	// eng.Primary() call runs - render.Config's doc comment explains why
	// that line can't live inside render.Config itself, and every caller
	// (the CLI image command, this one) sets it at its own call site instead.
	terminal.CaptureRuns = true

	columns := jsInt(optionsGet(options, "columns"), 120)

	// resetTemplateCache is true, unlike the CLI's single-shot image/data
	// commands: this wasm instance stays alive across many calls to
	// render() - once per keystroke in a studio preview, say - so each call
	// must start the template cache fresh, or one render's Var/Maps would
	// leak into the next. See render.Config's own doc comment for the exact
	// same reasoning applied to recordThemeSanitized's per-theme loop.
	eng, err := render.Config(cfg, columns, true, applyData)
	if err != nil {
		return "", fmt.Errorf("failed to render config: %w", err)
	}

	metrics := render.FontMetrics{
		CellWidth:   jsFloat(optionsGet(options, "cellWidth")),
		LineHeight:  jsFloat(optionsGet(options, "lineHeight")),
		FillAscent:  jsFloat(optionsGet(options, "fillAscent")),
		FillDescent: jsFloat(optionsGet(options, "fillDescent")),
	}

	opts := render.SVGOptions(jsString(optionsGet(options, "fontFamily")), columns, metrics)

	return render.SVG(eng, opts), nil
}

// optionsGet reads key off options, tolerating every shape a JS caller might
// hand in for its optional fourth argument: omitted (js.Undefined()), an
// explicit null, or simply not an object at all. js.Value.Get panics if
// called on a value that isn't a JavaScript object, so this guards that
// instead of letting a loosely-typed JS caller crash the whole render() call
// (see renderJS's own doc comment on why nothing here should ever panic).
// A real object with the key simply absent needs no such guard - Get already
// answers that with js.Undefined(), same as a missing map key would in Go.
func optionsGet(options js.Value, key string) js.Value {
	if options.Type() != js.TypeObject {
		return js.Undefined()
	}

	return options.Get(key)
}

// jsString reads v as a string, treating undefined/null as "" rather than
// panicking - the natural empty value for every string argument this
// package reads (an omitted dataJSON, an omitted optionsObject.fontFamily).
func jsString(v js.Value) string {
	if v.IsUndefined() || v.IsNull() {
		return ""
	}

	return v.String()
}

// jsFloat reads v as a float64, treating undefined/null as 0 - the same
// "unset" sentinel render.SVGOptions already gives every FontMetrics field
// (see its own doc comment: a zero metric falls back to svg's own default).
func jsFloat(v js.Value) float64 {
	if v.IsUndefined() || v.IsNull() {
		return 0
	}

	return v.Float()
}

// jsInt reads v as an int, falling back to fallback for undefined/null
// instead of panicking - used for optionsObject.columns, where 0 would be a
// nonsensical canvas width rather than a meaningful "unset".
func jsInt(v js.Value, fallback int) int {
	if v.IsUndefined() || v.IsNull() {
		return fallback
	}

	return v.Int()
}
