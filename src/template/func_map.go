package template

import (
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func funcMap() template.FuncMap {
	funcMap := map[string]any{
		"secondsRound": secondsRound,
		"url":          url,
		"path":         path,
		"glob":         glob,
		"matchP":       matchP,
		"replaceP":     replaceP,
		"gt":           gt,
		"lt":           lt,
		"reason":       GetReasonFromStatus,
		"hresult":      hresult,
		"trunc":        trunc,
		"readFile":     readFile,
	}
	for key, fun := range sprig.TxtFuncMap() {
		if _, ok := funcMap[key]; !ok {
			funcMap[key] = fun
		}
	}
	return template.FuncMap(funcMap)
}
