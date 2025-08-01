package path

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

// Base returns the last element of path.
// Trailing path separators are removed before extracting the last element.
// If the path consists entirely of separators, Base returns a single separator.
func Base(input string) string {
	volumeName := filepath.VolumeName(input)
	// Strip trailing slashes.
	for len(input) > 0 && IsSeparator(input[len(input)-1]) {
		input = input[0 : len(input)-1]
	}

	if input == "" {
		return Separator()
	}

	if volumeName == input {
		return input
	}

	// Throw away volume name
	input = input[len(filepath.VolumeName(input)):]
	// Find the last element
	i := len(input) - 1
	for i >= 0 && !IsSeparator(input[i]) {
		i--
	}

	if i >= 0 {
		input = input[i+1:]
	}

	// If empty now, it had only slashes.
	if input == "" {
		return Separator()
	}

	return input
}

func Clean(input string) string {
	if input == "" {
		return input
	}

	cleaned := input
	separator := Separator()

	// The prefix can be empty for a relative path.
	var prefix string
	if IsSeparator(cleaned[0]) {
		prefix = separator
	}

	if runtime.GOOS == windows {
		// Normalize (forward) slashes to backslashes on Windows.
		cleaned = strings.ReplaceAll(cleaned, "/", `\`)

		// Clean the prefix for a UNC path, if any.
		if regex.MatchString(`^\\{2}[^\\]+`, cleaned) {
			cleaned = strings.TrimPrefix(cleaned, `\\.\UNC\`)
			if cleaned == "" {
				return cleaned
			}
			prefix = `\\`
		}

		// Always use an uppercase drive letter on Windows.
		driveLetter, err := regex.GetCompiledRegex(`^[a-z]:`)
		if err == nil {
			cleaned = driveLetter.ReplaceAllStringFunc(cleaned, strings.ToUpper)
		}
	}

	sb := text.NewBuilder()

	sb.WriteString(prefix)

	// Clean slashes.
	matches := regex.FindAllNamedRegexMatch(fmt.Sprintf(`(?P<element>[^\%s]+)`, separator), cleaned)
	n := len(matches) - 1
	for i, m := range matches {
		sb.WriteString(m["element"])
		if i != n {
			sb.WriteString(separator)
		}
	}

	return sb.String()
}

func ReplaceHomeDirPrefixWithTilde(path string) string {
	home := Home()
	if !strings.HasPrefix(path, home) {
		return path
	}

	rem := path[len(home):]
	if rem == "" || IsSeparator(rem[0]) {
		return "~" + rem
	}

	return path
}

func ReplaceTildePrefixWithHomeDir(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	rem := path[1:]
	if rem == "" || IsSeparator(rem[0]) {
		return Home() + rem
	}

	return path
}
