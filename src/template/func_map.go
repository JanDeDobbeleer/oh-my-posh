package template

import (
	"path/filepath"
	"sync"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// sharedFuncMap is built exactly once and reused across all template constructions.
var sharedFuncMap = sync.OnceValue(func() template.FuncMap {
	fm := map[string]any{
		"secondsRound": secondsRound,
		"url":          url,
		"path":         filePath,
		"glob":         glob,
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
		"cmd":          cmd,
		"readFile":     readFile,
		"stat":         stat,
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
		if _, ok := fm[key]; !ok {
			fm[key] = fun
		}
	}

	return template.FuncMap(fm)
})

// funcMap returns the shared merged FuncMap (built once, reused everywhere).
func funcMap() template.FuncMap {
	return sharedFuncMap()
}
