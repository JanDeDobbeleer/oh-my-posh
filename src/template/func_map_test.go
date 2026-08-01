package template

import (
	"sort"
	"testing"

	"github.com/Masterminds/sprig/v3"

	"github.com/stretchr/testify/assert"
)

// TestRestrictedAllowedSprigFuncsCoversAllSprigFuncs is the drift guard: a
// sprig upgrade that adds a new function must fail this test rather than
// silently expose that function to untrusted templates. Every sprig function
// must be triaged into exactly one of dangerousFuncs or
// restrictedAllowedSprigFuncs before it can be used at all — unless it's
// shadowed by an oh-my-posh local function of the same name (localFuncMap),
// which always wins over the sprig original and so is never reachable.
func TestRestrictedAllowedSprigFuncsCoversAllSprigFuncs(t *testing.T) {
	localOverrides := localFuncMap()

	for name := range sprig.TxtFuncMap() {
		if dangerousFuncs[name] || restrictedAllowedSprigFuncs[name] {
			continue
		}

		if _, ok := localOverrides[name]; ok {
			continue
		}

		t.Errorf("sprig function %q is untriaged: add it to dangerousFuncs if it does I/O, network, "+
			"or heavy-CPU work, or to restrictedAllowedSprigFuncs if it's confirmed safe for untrusted input", name)
	}
}

// TestRestrictedFuncMapExactKeySet pins the exact set of functions exposed to
// untrusted templates for the current sprig version, guarding against
// accidental behavior changes to restrictedFuncMap.
func TestRestrictedFuncMapExactKeySet(t *testing.T) {
	expected := []string{
		"abbrev", "abbrevboth", "add", "add1", "add1f", "addf", "adler32sum", "ago", "all", "any", "append", "atoi",
		"b32dec", "b32enc", "b64dec", "b64enc", "base", "biggest", "camelcase", "cat", "ceil", "chunk", "clean",
		"coalesce", "compact", "concat", "contains", "date", "dateInZone", "dateModify", "date_in_zone", "date_modify",
		"deepCopy", "deepEqual", "default", "dict", "dig", "dir", "div", "divf", "duration", "durationRound", "empty",
		"ext", "fail", "findP", "first", "float64", "floor", "fromJson", "get", "gt", "has", "hasKey", "hasPrefix",
		"hasSuffix", "hello", "hresult", "htmlDate", "htmlDateInZone", "indent", "initial", "initials", "int", "int64",
		"isAbs", "join", "kebabcase", "keys", "kindIs", "kindOf", "last", "list", "localeShortDate", "localeShortTime",
		"lower", "lt", "matchP", "max", "maxf", "merge", "mergeOverwrite", "min", "minf", "mod", "mul", "mulf",
		"mustAppend", "mustChunk", "mustCompact", "mustDateModify", "mustDeepCopy", "mustFirst", "mustFromJson",
		"mustHas", "mustInitial", "mustLast", "mustMerge", "mustMergeOverwrite", "mustPrepend", "mustPush",
		"mustRegexFind", "mustRegexFindAll", "mustRegexMatch", "mustRegexReplaceAll", "mustRegexReplaceAllLiteral",
		"mustRegexSplit", "mustRest", "mustReverse", "mustSlice", "mustToDate", "mustToJson", "mustToPrettyJson",
		"mustToRawJson", "mustUniq", "mustWithout", "must_date_modify", "nindent", "nospace", "now", "omit", "osBase",
		"osClean", "osDir", "osExt", "osIsAbs", "path", "pick", "pluck", "plural", "prepend", "push", "quote",
		"randAlpha", "randAlphaNum", "randAscii", "randInt", "randNumeric", "random", "reason", "regexFind",
		"regexFindAll", "regexMatch", "regexQuoteMeta", "regexReplaceAll", "regexReplaceAllLiteral", "regexSplit",
		"repeat", "replace", "replaceP", "rest", "reverse", "round", "secondsRound", "semver", "semverCompare", "seq",
		"set", "sha1sum", "sha256sum", "sha512sum", "shuffle", "slice", "snakecase", "sortAlpha", "split", "splitList",
		"splitn", "squote", "sub", "subf", "substr", "swapcase", "ternary", "title", "toDate", "toDecimal", "toJson",
		"toPrettyJson", "toRawJson", "toString", "toStrings", "trim", "trimAll", "trimPrefix", "trimSuffix", "trimall",
		"trunc", "truncE", "tuple", "typeIs", "typeIsLike", "typeOf", "uniq", "unixEpoch", "unset", "until", "untilStep",
		"untitle", "upper", "url", "urlJoin", "urlParse", "uuidv4", "values", "without", "wrap", "wrapWith",
	}

	fm := funcMap(false)

	actual := make([]string, 0, len(fm))
	for key := range fm {
		actual = append(actual, key)
	}

	sort.Strings(actual)
	sort.Strings(expected)

	assert.Equal(t, expected, actual)
}
