package template

import (
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func funcMap() template.FuncMap {
	funcMap := map[string]interface{}{
		"secondsRound": secondsRound,
		"url":          url,
		"path":         path,
		"glob":         glob,
		"matchP":       matchP,
		"replaceP":     replaceP,
	}
	for key, fun := range sprig.TxtFuncMap() {
		if _, ok := funcMap[key]; !ok {
			funcMap[key] = fun
		}
	}
	return template.FuncMap(funcMap)
}
