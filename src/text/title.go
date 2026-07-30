package text

import (
	"strings"
	"unicode"
)

// Title uppercases the first letter of each word and lowercases every other
// letter. Words break on any character other than a letter, digit or
// underscore, and digits never count as a word's first letter - matching how
// x/text's cases.Title(language.English) handled the ASCII identifiers this
// replaces it for: "nix-shell" -> "Nix-Shell", "ui5tooling" -> "Ui5tooling",
// "foo_bar" -> "Foo_bar".
func Title(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))

	first := true

	for _, r := range s {
		switch {
		case unicode.IsLetter(r):
			if first {
				sb.WriteRune(unicode.ToUpper(r))
				first = false
				continue
			}

			sb.WriteRune(unicode.ToLower(r))
		case unicode.IsDigit(r) || r == '_':
			sb.WriteRune(r)
		default:
			sb.WriteRune(r)
			first = true
		}
	}

	return sb.String()
}
