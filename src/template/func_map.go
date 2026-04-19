package template

import (
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func normalizeUnixEpoch(value any) any {
	text, ok := value.(string)
	if !ok {
		return value
	}

	epoch, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return value
	}

	return epoch
}

func funcMap() template.FuncMap {
	sprigFuncMap := sprig.TxtFuncMap()
	sprigDate := sprigFuncMap["date"].(func(string, interface{}) string)
	sprigDateInZone := sprigFuncMap["date_in_zone"].(func(string, interface{}, string) string)
	sprigHTMLDate := sprigFuncMap["htmlDate"].(func(interface{}) string)
	sprigHTMLDateInZone := sprigFuncMap["htmlDateInZone"].(func(interface{}, string) string)

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
		"truncE":       TruncE,
		"readFile":     readFile,
		"stat":         stat,
		"dir":          filepath.Dir,
		"base":         filepath.Base,
		"date": func(layout string, value any) string {
			return sprigDate(layout, normalizeUnixEpoch(value))
		},
		"date_in_zone": func(layout string, value any, zone string) string {
			return sprigDateInZone(layout, normalizeUnixEpoch(value), zone)
		},
		"dateInZone": func(layout string, value any, zone string) string {
			return sprigDateInZone(layout, normalizeUnixEpoch(value), zone)
		},
		"htmlDate": func(value any) string {
			return sprigHTMLDate(normalizeUnixEpoch(value))
		},
		"htmlDateInZone": func(value any, zone string) string {
			return sprigHTMLDateInZone(normalizeUnixEpoch(value), zone)
		},
	}

	for key, fun := range sprigFuncMap {
		if _, ok := funcMap[key]; !ok {
			funcMap[key] = fun
		}
	}

	return template.FuncMap(funcMap)
}
