package text

import "strings"

// StripJSONComments removes // line and /* */ block comments from JSONC
// input so it can be fed to encoding/json. String contents are left
// untouched, comment markers included.
func StripJSONComments(src string) string {
	var sb strings.Builder
	sb.Grow(len(src))

	for i := 0; i < len(src); {
		c := src[i]

		switch {
		case c == '"':
			start := i
			i++

			for i < len(src) {
				if src[i] == '\\' {
					i += 2
					continue
				}

				if src[i] == '"' {
					i++
					break
				}

				i++
			}

			if i > len(src) {
				i = len(src)
			}

			sb.WriteString(src[start:i])
		case c == '/' && i+1 < len(src) && src[i+1] == '/':
			for i < len(src) && src[i] != '\n' {
				i++
			}
		case c == '/' && i+1 < len(src) && src[i+1] == '*':
			i += 2
			for i+1 < len(src) && (src[i] != '*' || src[i+1] != '/') {
				i++
			}

			i += 2
			if i > len(src) {
				i = len(src)
			}
		default:
			sb.WriteByte(c)
			i++
		}
	}

	return strings.TrimSpace(sb.String())
}
