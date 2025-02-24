package template

import (
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func funcMap() template.FuncMap {
	funcMap := map[string]any{
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
		"truncE":       truncE,
		"readFile":     readFile,
		"stat":         stat,
		"dir":          filepath.Dir,
		"base":         filepath.Base,
	}

	for key, fun := range sprig.TxtFuncMap() {
		if _, ok := funcMap[key]; !ok {
			funcMap[key] = fun
		}
	}

	return template.FuncMap(funcMap)
}
