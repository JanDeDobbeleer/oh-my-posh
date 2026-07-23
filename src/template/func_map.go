package template

import (
	"path/filepath"
	"sync"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// dangerousFuncs execute host commands or touch the filesystem/environment
// directly. They are only exposed to templates rendered via RenderTrusted
// (see text.go) — never to RenderUntrusted, which may contain or be composed
// from runtime data.
var dangerousFuncs = map[string]bool{
	"cmd":       true,
	"readFile":  true,
	"stat":      true,
	"glob":      true,
	"env":       true,
	"expandenv": true,
}

// baseFuncMap returns the funcs available regardless of trust level.
func baseFuncMap() map[string]any {
	fm := map[string]any{
		"secondsRound": secondsRound,
		"url":          url,
		"path":         filePath,
		"matchP":       matchP,
		"findP":        findP,
		"replaceP":     replaceP,
		"gt":           gt,
		"lt":           lt,
		"random":       random,
		"reason":       GetReasonFromStatus,
		"hresult":      hresult,
		"trunc":        trunc,
		"truncE":       TruncE,
		"dir":          filepath.Dir,
		"base":         filepath.Base,
		// Locale-aware date/time formatting using OS regional settings.
		"localeShortDate": localeShortDate,
		"localeShortTime": localeShortTime,
		// Override sprig date functions to support string epoch values (e.g. output of unixEpoch).
		"date":           ompDate,
		"date_in_zone":   ompDateInZone,
		"dateInZone":     ompDateInZone,
		"htmlDate":       ompHTMLDate,
		"htmlDateInZone": ompHTMLDateInZone,
	}

	for key, fun := range sprig.TxtFuncMap() {
		if dangerousFuncs[key] {
			continue
		}

		if _, ok := fm[key]; !ok {
			fm[key] = fun
		}
	}

	return fm
}

// sharedFuncMap is built exactly once and reused across all trusted template constructions.
var sharedFuncMap = sync.OnceValue(func() template.FuncMap {
	fm := baseFuncMap()

	fm["cmd"] = cmd
	fm["readFile"] = readFile
	fm["stat"] = stat
	fm["glob"] = glob

	sprigFuncs := sprig.TxtFuncMap()
	if fn, ok := sprigFuncs["env"]; ok {
		fm["env"] = fn
	}

	if fn, ok := sprigFuncs["expandenv"]; ok {
		fm["expandenv"] = fn
	}

	return template.FuncMap(fm)
})

// restrictedFuncMap is built exactly once and reused across all restricted template
// constructions (see RenderRestricted). It never contains dangerousFuncs.
var restrictedFuncMap = sync.OnceValue(func() template.FuncMap {
	return template.FuncMap(baseFuncMap())
})

// funcMap returns the merged FuncMap for the given trust level (built once per
// level, reused everywhere).
func funcMap(trusted bool) template.FuncMap {
	if trusted {
		return sharedFuncMap()
	}

	return restrictedFuncMap()
}
