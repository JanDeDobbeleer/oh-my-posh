package template

import (
	"fmt"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// dangerousFuncs execute host commands, touch the filesystem/environment
// directly, perform network I/O, or are CPU-expensive enough to hang the
// shell if called in a loop. They are only exposed to templates rendered via
// RenderTrusted (see text.go) — never to RenderUntrusted, which may contain
// or be composed from runtime data.
var dangerousFuncs = map[string]bool{
	"cmd":       true,
	"readFile":  true,
	"stat":      true,
	"glob":      true,
	"env":       true,
	"expandenv": true,

	// getHostByName does a live DNS lookup, which can be used to exfiltrate
	// data (e.g. an env var) from an untrusted template over DNS.
	"getHostByName": true,

	// CPU-expensive sprig functions: reachable untrusted text (e.g. a
	// crafted folder name) can call these in a loop to hang the shell.
	"genPrivateKey":            true,
	"genCA":                    true,
	"genCAWithKey":             true,
	"genSelfSignedCert":        true,
	"genSelfSignedCertWithKey": true,
	"genSignedCert":            true,
	"genSignedCertWithKey":     true,
	"buildCustomCert":          true,
	"bcrypt":                   true,
	"htpasswd":                 true,
	"derivePassword":           true,
	"encryptAES":               true,
	"decryptAES":               true,
	"randBytes":                true,
}

// restrictedAllowedSprigFuncs is a frozen allowlist of sprig functions that
// have been reviewed and confirmed safe to expose to untrusted templates
// (see restrictedFuncMap). It is intentionally NOT derived from
// dangerousFuncs: the two lists must be maintained independently so that a
// sprig function is only reachable from untrusted input after an explicit,
// reviewed addition here, rather than merely being absent from a blocklist.
// A sprig upgrade that adds new functions must not silently expose them —
// see TestRestrictedAllowedSprigFuncsCoversAllSprigFuncs, which fails CI
// until every new sprig function is triaged into one list or the other.
var restrictedAllowedSprigFuncs = map[string]bool{
	"abbrev": true, "abbrevboth": true, "add": true, "add1": true, "add1f": true, "addf": true, "adler32sum": true,
	"ago": true, "all": true, "any": true, "append": true, "atoi": true, "b32dec": true, "b32enc": true,
	"b64dec": true, "b64enc": true, "biggest": true, "camelcase": true, "cat": true, "ceil": true, "chunk": true,
	"clean": true, "coalesce": true, "compact": true, "concat": true, "contains": true, "dateModify": true,
	"date_modify": true, "deepCopy": true, "deepEqual": true, "default": true, "dict": true, "dig": true, "div": true,
	"divf": true, "duration": true, "durationRound": true, "empty": true, "ext": true, "fail": true, "first": true,
	"float64": true, "floor": true, "fromJson": true, "get": true, "has": true, "hasKey": true, "hasPrefix": true,
	"hasSuffix": true, "hello": true, "indent": true, "initial": true, "initials": true, "int": true, "int64": true,
	"isAbs": true, "join": true, "kebabcase": true, "keys": true, "kindIs": true, "kindOf": true, "last": true,
	"list": true, "lower": true, "max": true, "maxf": true, "merge": true, "mergeOverwrite": true, "min": true,
	"minf": true, "mod": true, "mul": true, "mulf": true, "mustAppend": true, "mustChunk": true, "mustCompact": true,
	"mustDateModify": true, "mustDeepCopy": true, "mustFirst": true, "mustFromJson": true, "mustHas": true,
	"mustInitial": true, "mustLast": true, "mustMerge": true, "mustMergeOverwrite": true, "mustPrepend": true,
	"mustPush": true, "mustRegexFind": true, "mustRegexFindAll": true, "mustRegexMatch": true,
	"mustRegexReplaceAll": true, "mustRegexReplaceAllLiteral": true, "mustRegexSplit": true, "mustRest": true,
	"mustReverse": true, "mustSlice": true, "mustToDate": true, "mustToJson": true, "mustToPrettyJson": true,
	"mustToRawJson": true, "mustUniq": true, "mustWithout": true, "must_date_modify": true, "nindent": true,
	"nospace": true, "now": true, "omit": true, "osBase": true, "osClean": true, "osDir": true, "osExt": true,
	"osIsAbs": true, "pick": true, "pluck": true, "plural": true, "prepend": true, "push": true, "quote": true,
	"randAlpha": true, "randAlphaNum": true, "randAscii": true, "randInt": true, "randNumeric": true,
	"regexFind": true, "regexFindAll": true, "regexMatch": true, "regexQuoteMeta": true, "regexReplaceAll": true,
	"regexReplaceAllLiteral": true, "regexSplit": true, "repeat": true, "replace": true, "rest": true,
	"reverse": true, "round": true, "semver": true, "semverCompare": true, "seq": true, "set": true, "sha1sum": true,
	"sha256sum": true, "sha512sum": true, "shuffle": true, "slice": true, "snakecase": true, "sortAlpha": true,
	"split": true, "splitList": true, "splitn": true, "squote": true, "sub": true, "subf": true, "substr": true,
	"swapcase": true, "ternary": true, "title": true, "toDate": true, "toDecimal": true, "toJson": true,
	"toPrettyJson": true, "toRawJson": true, "toString": true, "toStrings": true, "trim": true, "trimAll": true,
	"trimPrefix": true, "trimSuffix": true, "trimall": true, "tuple": true, "typeIs": true, "typeIsLike": true,
	"typeOf": true, "uniq": true, "unixEpoch": true, "unset": true, "until": true, "untilStep": true, "untitle": true,
	"upper": true, "urlJoin": true, "urlParse": true, "uuidv4": true, "values": true, "without": true, "wrap": true,
	"wrapWith": true,
}

// escapeFuncName is the internal function appended to every print action's
// pipeline after parsing (see escapePrintActions). It must resolve in both
// trust levels, so it lives in the shared local map.
const escapeFuncName = "__esc"

// localFuncMap holds oh-my-posh's own template functions, including ones
// that intentionally override a sprig function of the same name (e.g. date,
// to support string epoch values). These are unconditionally available in
// both the trusted and restricted func maps.
func localFuncMap() map[string]any {
	fm := map[string]any{
		escapeFuncName: escapeActionValue,
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
		"dir":          filepath.Dir,
		"base":         filepath.Base,
		"trunc":        trunc,
		"truncE":       TruncE,
		// The print builtins format through fmt and return plain strings, so
		// they would drop the trust of a Markup argument; these overrides go
		// through markupAware like every other function.
		"print":   fmt.Sprint,
		"printf":  fmt.Sprintf,
		"println": fmt.Sprintln,
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

	return fm
}

// wrapMarkupAware replaces every function in fm with its Markup-aware form
// (see markupAware). The escape function is the one exception: it is the
// boundary the wrapper protects, and must see the raw value.
func wrapMarkupAware(fm map[string]any) template.FuncMap {
	for name, fn := range fm {
		if name == escapeFuncName {
			continue
		}

		fm[name] = markupAware(name, fn)
	}

	return template.FuncMap(fm)
}

func baseFuncMap() map[string]any {
	fm := localFuncMap()

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

var sharedFuncMap = sync.OnceValue(func() template.FuncMap {
	fm := baseFuncMap()

	fm["cmd"] = cmd
	fm["readFile"] = readFile
	fm["stat"] = stat
	fm["glob"] = glob

	sprigFuncs := sprig.TxtFuncMap()
	for _, name := range []string{
		"env",
		"expandenv",
		"getHostByName",
		"genPrivateKey",
		"genCA",
		"genCAWithKey",
		"genSelfSignedCert",
		"genSelfSignedCertWithKey",
		"genSignedCert",
		"genSignedCertWithKey",
		"buildCustomCert",
		"bcrypt",
		"htpasswd",
		"derivePassword",
		"encryptAES",
		"decryptAES",
		"randBytes",
	} {
		if fn, ok := sprigFuncs[name]; ok {
			fm[name] = fn
		}
	}

	return wrapMarkupAware(fm)
})

// Reused across all restricted template constructions (see RenderUntrusted).
// Built as an allowlist, not baseFuncMap() minus dangerousFuncs: only sprig
// functions named in restrictedAllowedSprigFuncs are exposed, so a sprig
// upgrade can never widen what untrusted templates can call without a
// reviewed change to that list.
var restrictedFuncMap = sync.OnceValue(func() template.FuncMap {
	fm := localFuncMap()

	for key, fun := range sprig.TxtFuncMap() {
		if dangerousFuncs[key] {
			continue
		}

		if !restrictedAllowedSprigFuncs[key] {
			continue
		}

		if _, ok := fm[key]; !ok {
			fm[key] = fun
		}
	}

	return wrapMarkupAware(fm)
})

func funcMap(trusted bool) template.FuncMap {
	if trusted {
		return sharedFuncMap()
	}

	return restrictedFuncMap()
}
